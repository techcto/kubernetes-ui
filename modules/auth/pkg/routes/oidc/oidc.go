// Copyright 2017 The Kubernetes Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package oidc

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"regexp"
	"sort"

	gooidc "github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
	"k8s.io/klog/v2"

	"k8s.io/dashboard/auth/pkg/args"
	"k8s.io/dashboard/client"
)

// kubernetesRolePriority mirrors Cloud's ClusterMember::getKubernetesRole()
// mapping - Keycloak client role names match the K8s built-in ClusterRole
// names 1:1 (view/edit/admin) for legacy clusters. New clusters preserve
// organization-scoped group names so Kubernetes RoleBindings can enforce the
// namespace boundary, while owner remains a distinct cluster-wide group.
var kubernetesRolePriority = []string{"admin", "edit", "view"}

var namespaceRolePattern = regexp.MustCompile(`^spacemade:namespace:[a-z0-9](?:[-a-z0-9]*[a-z0-9])?:(?:admin|edit|view)$`)

type provider struct {
	oauth2Config *oauth2.Config
	verifier     *gooidc.IDTokenVerifier
}

var current *provider

// Setup discovers the configured OIDC issuer and prepares the OAuth2/OIDC
// clients used by the login/callback handlers. A no-op (OIDC left disabled)
// if no issuer/client id was configured - not every install has SSO wired up.
func Setup() error {
	if !args.OIDCEnabled() {
		klog.InfoS("OIDC not configured, SSO login disabled")
		return nil
	}

	oidcProvider, err := gooidc.NewProvider(context.Background(), args.OIDCIssuerURL())
	if err != nil {
		return fmt.Errorf("could not discover OIDC provider at %q: %w", args.OIDCIssuerURL(), err)
	}

	current = &provider{
		oauth2Config: &oauth2.Config{
			ClientID:     args.OIDCClientID(),
			ClientSecret: args.OIDCClientSecret(),
			RedirectURL:  args.OIDCRedirectURL(),
			Endpoint:     oidcProvider.Endpoint(),
			Scopes:       args.OIDCScopes(),
		},
		verifier: oidcProvider.Verifier(&gooidc.Config{ClientID: args.OIDCClientID()}),
	}

	return nil
}

func enabled() bool {
	return current != nil
}

func randomState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func generateVerifier() string {
	return oauth2.GenerateVerifier()
}

// loginRedirectURL returns Keycloak's authorization endpoint URL. state and
// verifier must be stashed by the caller (e.g. in cookies) to be checked
// against / reused on callback.
func loginRedirectURL(state, verifier string) string {
	return current.oauth2Config.AuthCodeURL(state, oauth2.S256ChallengeOption(verifier))
}

// exchange trades an authorization code for tokens, verifies the ID token,
// and extracts the caller's Keycloak-issued username and Cloud role (view/
// edit/admin - see resource_access.<client_id>.roles in the ID token, set by
// KeycloakService::createRolesForClient / ClusterMember::getKubernetesRole
// on the cloud side).
func exchange(ctx context.Context, code, verifier string) (username string, groups []string, err error) {
	token, err := current.oauth2Config.Exchange(ctx, code, oauth2.VerifierOption(verifier))
	if err != nil {
		return "", nil, fmt.Errorf("code exchange failed: %w", err)
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return "", nil, fmt.Errorf("no id_token in token response")
	}

	idToken, err := current.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return "", nil, fmt.Errorf("id_token verification failed: %w", err)
	}

	var claims struct {
		PreferredUsername string `json:"preferred_username"`
		ResourceAccess    map[string]struct {
			Roles []string `json:"roles"`
		} `json:"resource_access"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return "", nil, fmt.Errorf("could not parse id_token claims: %w", err)
	}

	groups = kubernetesGroups(claims.ResourceAccess[args.OIDCClientID()].Roles)
	if len(groups) == 0 {
		return "", nil, fmt.Errorf("id_token has no authorized Kubernetes role")
	}

	return claims.PreferredUsername, groups, nil
}

func kubernetesGroups(roles []string) []string {
	has := make(map[string]bool, len(roles))
	for _, role := range roles {
		has[role] = true
	}
	if has["owner"] {
		return []string{"owner"}
	}

	groups := make([]string, 0)
	for role := range has {
		if namespaceRolePattern.MatchString(role) {
			groups = append(groups, role)
		}
	}
	if len(groups) > 0 {
		sort.Strings(groups)
		return groups
	}

	for _, role := range kubernetesRolePriority {
		if has[role] {
			return []string{role}
		}
	}

	return nil
}

func highestRole(roles []string) string {
	groups := kubernetesGroups(roles)
	if len(groups) > 0 {
		return groups[0]
	}
	return ""
}

// issueSession signs a short-lived session token carrying the user's
// identity and role, using the same signing key + claims shape modules/api
// verifies in its own request pipeline (see k8s.io/dashboard/client's
// SignSession/configFromRequest) - the frontend presents this as its own
// Authorization bearer token on subsequent requests.
func issueSession(username string, groups []string) (string, error) {
	return client.SignSessionGroups(username, groups)
}
