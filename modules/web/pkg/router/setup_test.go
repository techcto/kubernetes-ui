package router

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestAuthPaths(t *testing.T) {
	tests := map[string]bool{
		"/api/v1/login":                true,
		"/api/v1/csrftoken/login":      true,
		"/api/v1/me":                   true,
		"/api/v1/oidc":                 true,
		"/api/v1/oidc/login":           true,
		"/api/v1/oidc/callback":        true,
		"/api/v1/csrftoken/deployment": false,
		"/api/v1/node":                 false,
		"/":                            false,
	}

	for path, expected := range tests {
		if actual := isAuthPath(path); actual != expected {
			t.Errorf("isAuthPath(%q) = %v, want %v", path, actual, expected)
		}
	}
}

func TestGatewayRoutesAuthAndAPI(t *testing.T) {
	auth := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, _ = fmt.Fprint(writer, "auth:"+request.URL.Path)
	}))
	defer auth.Close()
	api := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, _ = fmt.Fprint(writer, "api:"+request.URL.Path)
	}))
	defer api.Close()

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(gatewayMiddleware(auth.URL, api.URL))
	engine.NoRoute(func(c *gin.Context) { c.String(http.StatusOK, "web:"+c.Request.URL.Path) })
	gateway := httptest.NewServer(engine)
	defer gateway.Close()

	tests := map[string]string{
		"/api/v1/oidc/login": "auth:/api/v1/oidc/login",
		"/api/v1/node":       "api:/api/v1/node",
		"/metrics":           "api:/metrics",
		"/":                  "web:/",
	}
	for path, expected := range tests {
		response, err := http.Get(gateway.URL + path)
		if err != nil {
			t.Fatalf("GET %s: %v", path, err)
		}
		body, err := io.ReadAll(response.Body)
		_ = response.Body.Close()
		if err != nil {
			t.Fatalf("read GET %s: %v", path, err)
		}
		if response.StatusCode != http.StatusOK || string(body) != expected {
			t.Errorf("GET %s = (%d, %q), want (200, %q)", path, response.StatusCode, string(body), expected)
		}
	}
}
