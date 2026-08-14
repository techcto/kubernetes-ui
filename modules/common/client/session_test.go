package client

import (
	"reflect"
	"testing"
)

func TestSessionPreservesAllKubernetesGroups(t *testing.T) {
	setSessionSigningKey("test-session-signing-key")
	groups := []string{
		"spacemade:namespace:osirus:view",
		"spacemade:namespace:techcto:admin",
	}
	token, err := SignSessionGroups("assigned-user", groups)
	if err != nil {
		t.Fatalf("SignSessionGroups() error = %v", err)
	}

	claims, ok := tryParseSession(token)
	if !ok {
		t.Fatal("tryParseSession() did not recognize signed session")
	}
	if !reflect.DeepEqual(claims.Groups, groups) {
		t.Fatalf("session groups = %v, want %v", claims.Groups, groups)
	}
}
