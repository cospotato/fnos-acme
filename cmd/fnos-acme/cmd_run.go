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
	"log/slog"
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
	flgCheckInterval        = "check-interval"
	flgRenewDays            = "renew-days"
	flgTermsOfServiceAgreed = "tos-agreed"
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
			&cli.BoolFlag{
				Name:    flgTermsOfServiceAgreed,
				Value:   false,
				Usage:   "agree the acme term of service",
				Sources: cli.EnvVars("ACME_TERM_OF_SERVICE_AGREED"),
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
		slog.Error("flag check failed", "err", err)
		return err
	}

	// create data dir
	if _, err := os.Stat(c.String(flgDataDir)); errors.Is(err, os.ErrNotExist) {
		if err := os.MkdirAll(c.String(flgDataDir), 0755); err != nil {
			slog.Error("create data dir failed", "err", err)
			return err
		}
	}

	// login fnos
	client, err := trim.NewMainClient(c.String(flgFnosAddress))
	if err != nil {
		slog.Error("create fnos client failed", "err", err)
		return err
	}

	defer client.Close()

	if _, err := client.Login(context.Background(), &user.LoginRequest{
		User:     c.String(flgFnosUsername),
		Password: c.String(flgFnosPassword),
	}); err != nil {
		slog.Error("login fnos failed", "err", err)
		return err
	}

	slog.Info("login fnos success")

	// login acme
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
		reg, err := legoClient.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: c.Bool(flgTermsOfServiceAgreed)})
		if err != nil {
			return err
		}

		account.Registration = reg
		if err := saveAccount(c.String(flgDataDir), account); err != nil {
			return err
		}

		slog.Info("registered acme account", "email", c.String(flgEmail))
	}

	// do checkAndUpdate immediately at starting up
	if err := checkAndUpdate(ctx, c, client, legoClient); err != nil {
		slog.Error("check certificate and update failed, wait next sync", "err", err, "nextSyncTime", time.Now().Add(c.Duration(flgCheckInterval)))
	}

	ticker := time.NewTicker(c.Duration(flgCheckInterval))

	for {
		select {
		case <-ticker.C:
			if err := checkAndUpdate(ctx, c, client, legoClient); err != nil {
				slog.Error("check certificate and update failed", "err", err)
			}

			slog.Info("wait next sync", "nextSyncTime", time.Now().Add(c.Duration(flgCheckInterval)))
		case <-ctx.Done():
			return nil
		}
	}
}

func ensureCert(ctx context.Context, trimClient *trim.Client, cert cert, edit bool) error {
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
		slog.Info("certificate in fnos not exists, upload it")

		resp, err := trimClient.Main().RemoteAccessService().UploadCert(ctx, &remoteaccess.UploadCertRequest{
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
		slog.Info("certificate in fnos out of date, replace it")

		resp, err := trimClient.Main().RemoteAccessService().ReplaceCert(ctx, &remoteaccess.ReplaceCertRequest{
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

func obtainAndUpload(ctx context.Context, dataDir string, domains []string, legoClient *lego.Client, trimClient *trim.Client) error {
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

	return ensureCert(ctx, trimClient, cert{
		name:    certResource.Domain,
		rawCert: certResource.Certificate,
		rawKey:  certResource.PrivateKey,
	}, true)
}

func checkAndUpdate(ctx context.Context, c *cli.Command, trimClient *trim.Client, legoClient *lego.Client) error {
	slog.Info("start check certificate")

	certs, err := listCertificates(c.String(flgDataDir))
	if err != nil {
		return err
	}

	if len(certs) == 0 {
		slog.Info("no certificate found, obtain one")
		return obtainAndUpload(ctx, c.String(flgDataDir), c.StringSlice(flgDomains), legoClient, trimClient)
	}

	for _, cert := range certs {
		ok := func() bool {
			if cert.name != c.StringSlice(flgDomains)[0] {
				return false
			}

			if !domainsEqual(c.StringSlice(flgDomains), certcrypto.ExtractDomains(cert.Certificate)) {
				return false
			}

			if time.Now().AddDate(0, 0, -1*int(c.Int(flgRenewDays))).After(cert.NotAfter) {
				return false
			}

			return true
		}

		if !ok() {
			slog.Info("certificate found, but out of date, obtain one and upload", "domain", cert.name)
			if err := obtainAndUpload(ctx, c.String(flgDataDir), c.StringSlice(flgDomains), legoClient, trimClient); err != nil {
				return err
			}
		}

		if err := ensureCert(ctx, trimClient, cert, false); err != nil {
			return err
		}
	}

	slog.Info("certificate is ready")

	return nil
}
