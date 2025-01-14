/*
 * Copyright 2025 CosPotato Lab.
 * Author: CosPotato Lin<i@0x233.cn>
 */

package main

import (
	"context"
	"time"

	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/providers/dns"
	"github.com/go-acme/lego/v4/registration"
)

const (
	userAgent = "fnos-acme/0.1.0"
)

func setupDNSChallenge(client *lego.Client, providerName string, wait time.Duration, resolvers []string) error {
	provider, err := dns.NewDNSChallengeProviderByName(providerName)
	if err != nil {
		return err
	}

	return client.Challenge.SetDNS01Provider(provider,
		dns01.CondOption(len(resolvers) > 0, dns01.AddRecursiveNameservers(dns01.ParseNameservers(resolvers))),
		dns01.CondOption(wait > 0, dns01.PropagationWait(wait, true)),
	)
}

func newClient(
	ctx context.Context,
	acc registration.User,
	keyType certcrypto.KeyType,
	providerName string,
	wait time.Duration,
	resolvers []string,
) (*lego.Client, error) {
	config := lego.NewConfig(acc)
	config.Certificate = lego.CertificateConfig{
		KeyType:             keyType,
		Timeout:             30 * time.Second,
		OverallRequestLimit: certificate.DefaultOverallRequestLimit,
	}

	config.UserAgent = userAgent

	client, err := lego.NewClient(config)
	if err != nil {
		return nil, err
	}

	if err := setupDNSChallenge(client, providerName, wait, resolvers); err != nil {
		return nil, err
	}

	return client, nil
}
