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
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"

	"k8s.io/dashboard/auth/pkg/router"
)

const (
	stateCookie     = "oidc_state"
	verifierCookie  = "oidc_verifier"
	namespaceCookie = "oidc_namespace"
	cookieMaxAge    = 300 // 5 minutes - just long enough for the Keycloak redirect round trip
)

func init() {
	if err := Setup(); err != nil {
		klog.ErrorS(err, "Could not set up OIDC provider - SSO login will be unavailable")
		return
	}

	router.V1().GET("/oidc/login", handleLogin)
	router.V1().GET("/oidc/callback", handleCallback)
}

func handleLogin(c *gin.Context) {
	if !enabled() {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "OIDC is not configured"})
		return
	}

	state, err := randomState()
	if err != nil {
		klog.ErrorS(err, "Could not generate OIDC state")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not start login"})
		return
	}
	verifier := generateVerifier()
	namespace := c.Query("namespace")
	if namespace != "" && !validNamespace(namespace) {
		redirectToLogin(c, "invalid_namespace", "")
		return
	}

	c.SetCookie(stateCookie, state, cookieMaxAge, "/api/v1/oidc", "", true, true)
	c.SetCookie(verifierCookie, verifier, cookieMaxAge, "/api/v1/oidc", "", true, true)
	if namespace != "" {
		c.SetCookie(namespaceCookie, namespace, cookieMaxAge, "/api/v1/oidc", "", true, true)
	} else {
		c.SetCookie(namespaceCookie, "", -1, "/api/v1/oidc", "", true, true)
	}

	c.Redirect(http.StatusFound, loginRedirectURL(state, verifier))
}

// handleCallback lands the browser back here directly (it's the OAuth
// redirect_uri), so on both success and failure it redirects into the SPA's
// (hash-routed) login page rather than returning JSON a browser can't do
// anything useful with. LoginComponent picks up ssoToken/ssoError from the
// query string - see modules/web/src/login/component.ts.
func handleCallback(c *gin.Context) {
	if !enabled() {
		redirectToLogin(c, "sso_not_configured", "")
		return
	}

	state, stateErr := c.Cookie(stateCookie)
	verifier, verifierErr := c.Cookie(verifierCookie)
	namespace, _ := c.Cookie(namespaceCookie)

	// Clear all regardless of outcome - they're single-use.
	c.SetCookie(stateCookie, "", -1, "/api/v1/oidc", "", true, true)
	c.SetCookie(verifierCookie, "", -1, "/api/v1/oidc", "", true, true)
	c.SetCookie(namespaceCookie, "", -1, "/api/v1/oidc", "", true, true)

	if stateErr != nil || state == "" || state != c.Query("state") {
		redirectToLogin(c, "invalid_state", namespace)
		return
	}
	if verifierErr != nil || verifier == "" {
		redirectToLogin(c, "invalid_state", namespace)
		return
	}

	code := c.Query("code")
	if code == "" {
		redirectToLogin(c, "missing_code", namespace)
		return
	}

	username, groups, err := exchange(c.Request.Context(), code, verifier)
	if err != nil {
		klog.ErrorS(err, "OIDC callback failed")
		redirectToLogin(c, "sso_failed", namespace)
		return
	}

	session, err := issueSession(username, groups)
	if err != nil {
		klog.ErrorS(err, "Could not issue session")
		redirectToLogin(c, "sso_failed", namespace)
		return
	}

	redirectURL := "/#/login?ssoToken=" + url.QueryEscape(session)
	if namespace != "" {
		redirectURL += "&namespace=" + url.QueryEscape(namespace)
	}
	c.Redirect(http.StatusFound, redirectURL)
}

func redirectToLogin(c *gin.Context, reason, namespace string) {
	redirectURL := "/#/login?ssoError=" + url.QueryEscape(reason)
	if namespace != "" {
		redirectURL += "&namespace=" + url.QueryEscape(namespace)
	}
	c.Redirect(http.StatusFound, redirectURL)
}
