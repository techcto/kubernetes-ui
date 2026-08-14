package oidc

import (
	"reflect"
	"testing"
)

func TestKubernetesGroupsPreservesOwner(t *testing.T) {
	got := kubernetesGroups([]string{"admin", "owner"})
	want := []string{"owner"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("kubernetesGroups() = %v, want %v", got, want)
	}
}

func TestKubernetesGroupsPreservesNamespaceGroups(t *testing.T) {
	got := kubernetesGroups([]string{
		"spacemade:namespace:osirus:view",
		"spacemade:namespace:techcto:admin",
	})
	want := []string{
		"spacemade:namespace:osirus:view",
		"spacemade:namespace:techcto:admin",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("kubernetesGroups() = %v, want %v", got, want)
	}
}

func TestKubernetesGroupsRejectsUnmanagedRoles(t *testing.T) {
	if got := kubernetesGroups([]string{"offline_access"}); len(got) != 0 {
		t.Fatalf("kubernetesGroups() = %v, want no groups", got)
	}
}
