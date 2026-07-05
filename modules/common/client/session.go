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

package client

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// SessionIssuer identifies tokens minted by modules/auth's OIDC callback
// (see modules/auth/pkg/routes/oidc), as opposed to a raw Kubernetes bearer
// token a user pasted in directly - both arrive in the same Authorization
// header, so this is how the two are told apart (see tryParseSession).
const SessionIssuer = "k8s-dashboard-auth"

// SessionClaims is the payload of the session token issued after a
// successful OIDC/SSO login. Username and Role come straight from the
// Keycloak ID token's preferred_username and resource_access claims.
type SessionClaims struct {
	jwt.RegisteredClaims
	Username string `json:"username"`
	Role     string `json:"role"`
}

var sessionSigningKey []byte

// setSessionSigningKey is called from Init via WithSessionSigningKey. Both
// modules/auth (which signs sessions) and modules/api (which verifies them
// in configFromRequest) must be given the same key.
func setSessionSigningKey(key string) {
	sessionSigningKey = []byte(key)
}

// SignSession mints a signed, short-lived session token for the given
// identity. Used by modules/auth after a successful OIDC login.
func SignSession(username, role string) (string, error) {
	if len(sessionSigningKey) == 0 {
		return "", fmt.Errorf("session signing key not configured")
	}

	claims := SessionClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    SessionIssuer,
			Subject:   username,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(8 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		Username: username,
		Role:     role,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(sessionSigningKey)
}

// tryParseSession verifies token as one of our own session tokens. Returns
// ok=false (not an error) whenever the token doesn't verify as one of ours -
// e.g. it's a real Kubernetes bearer token a user pasted in directly instead
// - in which case the caller should fall back to treating it as a raw token.
func tryParseSession(token string) (*SessionClaims, bool) {
	if len(sessionSigningKey) == 0 || token == "" {
		return nil, false
	}

	claims := &SessionClaims{}
	parsed, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		if _, isHMAC := t.Method.(*jwt.SigningMethodHMAC); !isHMAC {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return sessionSigningKey, nil
	})
	if err != nil || !parsed.Valid || claims.Issuer != SessionIssuer {
		return nil, false
	}

	return claims, true
}
