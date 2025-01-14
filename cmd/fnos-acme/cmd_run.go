/*
 * Copyright 2025 CosPotato Lab.
 * Author: CosPotato Lin<i@0x233.cn>
 */

package main

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/cospotato/fnos-acme/internal/trim"
	"github.com/cospotato/fnos-acme/internal/trim/api/remoteaccess"
	"github.com/cospotato/fnos-acme/internal/trim/api/user"
	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
	"github.com/urfave/cli/v3"
)

const (
	accountJson = "accounts.json"
)

const (
	flgCheckInterval = "check-interval"
	flgRenewDays     = "renew-days"
)

const (
	defaultKeyType = certcrypto.RSA2048
)

func commandRun() *cli.Command {
	return &cli.Command{
		Name:   "run",
		Action: run,
		Flags: []cli.Flag{
			&cli.DurationFlag{
				Name:    flgCheckInterval,
				Value:   time.Hour,
				Usage:   "cert check interval",
				Sources: cli.EnvVars("CHECK_INTERVAL"),
			},
			&cli.IntFlag{
				Name:    flgRenewDays,
				Value:   3,
				Usage:   "renew days",
				Sources: cli.EnvVars("RENEW_DAYS"),
			},
		},
	}
}

func flagCheck(c *cli.Command) error {
	if len(c.StringSlice(flgDomains)) == 0 {
		return fmt.Errorf("must specific DOMAINS")
	}

	if c.String(flgEmail) == "" {
		return fmt.Errorf("must specific EMAIL")
	}

	if c.String(flgFnosAddress) == "" {
		return fmt.Errorf("must specific FNOS_ADDRESS")
	}

	if c.String(flgFnosUsername) == "" {
		return fmt.Errorf("must specific FNOS_USERNAME")
	}

	if c.String(flgFnosPassword) == "" {
		return fmt.Errorf("must specific FNOS_PASSWORD")
	}

	if c.String(flgDnsProvider) == "" {
		return fmt.Errorf("must specific DNS_PROVIDER")
	}

	return nil
}

func run(ctx context.Context, c *cli.Command) error {
	if err := flagCheck(c); err != nil {
		return err
	}

	// create data dir
	if _, err := os.Stat(c.String(flgDataDir)); errors.Is(err, os.ErrNotExist) {
		if err := os.MkdirAll(c.String(flgDataDir), 0755); err != nil {
			return err
		}
	}

	client, err := trim.NewMainClient(c.String(flgFnosAddress))
	if err != nil {
		return err
	}

	defer client.Close()

	if _, err := client.Login(context.Background(), &user.LoginRequest{
		User:     c.String(flgFnosUsername),
		Password: c.String(flgFnosPassword),
	}); err != nil {
		return err
	}

	// do checkAndUpdate immediately at starting up
	if err := checkAndUpdate(ctx, c, client); err != nil {
		return err
	}

	ticker := time.NewTicker(c.Duration(flgCheckInterval))

	for {
		select {
		case <-ticker.C:
			if err := checkAndUpdate(ctx, c, client); err != nil {
				return err
			}
		case <-ctx.Done():
			return nil
		}
	}
}

func ensureCert(trimClient *trim.Client, cert cert, edit bool) error {
	certList, err := trimClient.Main().RemoteAccessService().GetCertList(context.TODO())
	if err != nil {
		return err
	}

	var remoteCert *remoteaccess.Cert

	for i := range certList.Data {
		if certList.Data[i].Domain == cert.name {
			remoteCert = &certList.Data[i]
			break
		}
	}

	if remoteCert == nil {
		resp, err := trimClient.Main().RemoteAccessService().UploadCert(context.TODO(), &remoteaccess.UploadCertRequest{
			Data: remoteaccess.CertRequestData{
				Desc:              cert.name,
				PrivateKeyBase64:  base64.StdEncoding.EncodeToString(cert.rawKey),
				CertificateBase64: base64.StdEncoding.EncodeToString(cert.rawCert),
			},
		})
		if err != nil {
			return err
		}

		if !resp.Data {
			return errors.New("upload cert return false")
		}
	}

	if edit {
		resp, err := trimClient.Main().RemoteAccessService().ReplaceCert(context.TODO(), &remoteaccess.ReplaceCertRequest{
			Data: remoteaccess.CertRequestData{
				ID:                remoteCert.ID,
				Desc:              cert.name,
				PrivateKeyBase64:  base64.StdEncoding.EncodeToString(cert.rawKey),
				CertificateBase64: base64.StdEncoding.EncodeToString(cert.rawCert),
			},
		})
		if err != nil {
			return err
		}

		if !resp.Data {
			return errors.New("upload cert return false")
		}
	}

	return nil
}

func obtainAndUpload(dataDir string, domains []string, legoClient *lego.Client, trimClient *trim.Client) error {
	request := certificate.ObtainRequest{
		Domains: domains,
		Bundle:  true,
	}

	certResource, err := legoClient.Certificate.Obtain(request)
	if err != nil {
		return err
	}

	if err := saveCertificate(dataDir, certResource); err != nil {
		return err
	}

	return ensureCert(trimClient, cert{
		name:    certResource.Domain,
		rawCert: certResource.Certificate,
		rawKey:  certResource.PrivateKey,
	}, true)
}

func checkAndUpdate(ctx context.Context, c *cli.Command, trimClient *trim.Client) error {
	account, err := setupAccount(c.String(flgDataDir), c.String(flgEmail))
	if err != nil {
		return err
	}

	legoClient, err := newClient(ctx, account, defaultKeyType, c.String(flgDnsProvider), 30*time.Second, c.StringSlice(flgDnsResolvers))
	if err != nil {
		return err
	}

	if account.Registration == nil {
		// register account
		// !!! ACCEPTED TOS BY DEFAULT
		reg, err := legoClient.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
		if err != nil {
			return err
		}

		account.Registration = reg
		if err := saveAccount(c.String(flgDataDir), account); err != nil {
			return err
		}
	}

	certs, err := listCertificates(c.String(flgDataDir))
	if err != nil {
		return err
	}

	if len(certs) == 0 {
		return obtainAndUpload(c.String(flgDataDir), c.StringSlice(flgDomains), legoClient, trimClient)
	}

	for _, cert := range certs {
		ok := func() bool {
			if cert.name != c.StringSlice(flgDomains)[0] {
				return false
			}

			if domainsEqual(c.StringSlice(flgDomains), certcrypto.ExtractDomains(cert.Certificate)) {
				return false
			}

			if cert.NotAfter.After(time.Now().AddDate(0, 0, -1*int(c.Int(flgRenewDays)))) {
				return false
			}

			return true
		}

		if !ok() {
			if err := obtainAndUpload(c.String(flgDataDir), c.StringSlice(flgDomains), legoClient, trimClient); err != nil {
				return err
			}
		}

		if err := ensureCert(trimClient, cert, false); err != nil {
			return err
		}
	}

	return nil
}
