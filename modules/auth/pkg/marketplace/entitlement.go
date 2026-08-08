// Copyright 2026 The Kubernetes Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package marketplace

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/marketplacemetering"
	"github.com/aws/smithy-go"
)

const maxAttempts = 5

type registerUsageAPI interface {
	RegisterUsage(context.Context, *marketplacemetering.RegisterUsageInput, ...func(*marketplacemetering.Options)) (*marketplacemetering.RegisterUsageOutput, error)
}

// Register verifies the buyer's entitlement and starts Marketplace metering.
// An empty product code deliberately disables the integration for local and
// non-Marketplace builds. Marketplace release charts always set it.
func Register(ctx context.Context, productCode string, publicKeyVersion int32) error {
	if productCode == "" {
		return nil
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("load AWS configuration: %w", err)
	}

	return register(ctx, marketplacemetering.NewFromConfig(cfg), productCode, publicKeyVersion)
}

func register(ctx context.Context, client registerUsageAPI, productCode string, publicKeyVersion int32) error {
	input := &marketplacemetering.RegisterUsageInput{
		ProductCode:      aws.String(productCode),
		PublicKeyVersion: aws.Int32(publicKeyVersion),
	}

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		_, err := client.RegisterUsage(ctx, input)
		if err == nil {
			return nil
		}

		var apiErr smithy.APIError
		if !errors.As(err, &apiErr) || apiErr.ErrorCode() != "ThrottlingException" || attempt == maxAttempts {
			return fmt.Errorf("register AWS Marketplace usage: %w", err)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Duration(attempt) * time.Second):
		}
	}

	return nil
}
