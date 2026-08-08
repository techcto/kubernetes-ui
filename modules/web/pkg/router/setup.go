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

package router

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"

	"k8s.io/dashboard/web/pkg/environment"
)

var (
	router *gin.Engine
	root   *gin.RouterGroup
)

func init() {
	if !environment.IsDev() {
		gin.SetMode(gin.ReleaseMode)
	}

	router = gin.New()
	router.Use(gin.Recovery())
	router.Use(gatewayMiddleware(
		os.Getenv("KUBERNETES_DASHBOARD_AUTH_URL"),
		os.Getenv("KUBERNETES_DASHBOARD_API_URL"),
	))
	_ = router.SetTrustedProxies(nil)
	root = router.Group("/")
}

func gatewayMiddleware(authAddress, apiAddress string) gin.HandlerFunc {
	authProxy := newReverseProxy(authAddress, "http://kubernetes-dashboard-auth:8000")
	apiProxy := newReverseProxy(apiAddress, "http://kubernetes-dashboard-api:8000")

	return func(c *gin.Context) {
		var proxy *httputil.ReverseProxy
		switch {
		case isAuthPath(c.Request.URL.Path):
			proxy = authProxy
		case strings.HasPrefix(c.Request.URL.Path, "/api/"), c.Request.URL.Path == "/metrics":
			proxy = apiProxy
		default:
			c.Next()
			return
		}

		c.Abort()
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

func isAuthPath(path string) bool {
	return path == "/api/v1/login" ||
		path == "/api/v1/csrftoken/login" ||
		path == "/api/v1/me" ||
		path == "/api/v1/oidc" ||
		strings.HasPrefix(path, "/api/v1/oidc/")
}

func newReverseProxy(address, fallback string) *httputil.ReverseProxy {
	if address == "" {
		address = fallback
	}
	target, err := url.Parse(address)
	if err != nil || target.Scheme == "" || target.Host == "" {
		panic("invalid dashboard gateway target: " + address)
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.FlushInterval = -1
	proxy.ErrorHandler = func(writer http.ResponseWriter, request *http.Request, proxyErr error) {
		klog.ErrorS(proxyErr, "Dashboard gateway request failed", "path", request.URL.Path, "target", target.String())
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusBadGateway)
		_, _ = writer.Write([]byte(`{"error":"dashboard backend unavailable"}`))
	}
	proxy.Transport = &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: time.Second,
	}
	return proxy
}

func Root() *gin.RouterGroup {
	return root
}

func Router() *gin.Engine {
	return router
}
