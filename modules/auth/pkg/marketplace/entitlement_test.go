package marketplace

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/marketplacemetering"
	"github.com/aws/smithy-go"
)

type fakeClient struct {
	err   error
	calls int
}

func (f *fakeClient) RegisterUsage(context.Context, *marketplacemetering.RegisterUsageInput, ...func(*marketplacemetering.Options)) (*marketplacemetering.RegisterUsageOutput, error) {
	f.calls++
	return &marketplacemetering.RegisterUsageOutput{}, f.err
}

func TestRegisterUsageSuccess(t *testing.T) {
	client := &fakeClient{}
	if err := register(context.Background(), client, "product-code", 1); err != nil {
		t.Fatalf("register() error = %v", err)
	}
	if client.calls != 1 {
		t.Fatalf("RegisterUsage calls = %d, want 1", client.calls)
	}
}

func TestRegisterUsageRejectsUnentitledBuyer(t *testing.T) {
	client := &fakeClient{err: &smithy.GenericAPIError{Code: "CustomerNotEntitledException", Message: "not subscribed"}}
	err := register(context.Background(), client, "product-code", 1)
	if err == nil {
		t.Fatal("register() succeeded for an unentitled buyer")
	}
	if !errors.Is(err, client.err) {
		t.Fatalf("register() error = %v, want wrapped Marketplace error", err)
	}
	if client.calls != 1 {
		t.Fatalf("RegisterUsage calls = %d, want 1", client.calls)
	}
}
