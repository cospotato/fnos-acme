/*
 * Copyright 2025 CosPotato Lab.
 * Author: CosPotato Lin<i@0x233.cn>
 */

package main

import (
	"crypto/x509"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
)

const (
	issuerExt   = ".issuer.crt"
	certExt     = ".crt"
	keyExt      = ".key"
	pemExt      = ".pem"
	pfxExt      = ".pfx"
	resourceExt = ".json"
)

type cert struct {
	*x509.Certificate
	rawCert []byte
	rawKey  []byte
	name    string
}

func listCertificates(dataDir string) ([]cert, error) {
	matches, err := filepath.Glob(filepath.Join(dataDir, "certificates", "*.crt"))
	if err != nil {
		return nil, err
	}

	if len(matches) == 0 {
		return nil, nil
	}

	certs := make([]cert, 0)

	for _, filename := range matches {
		if strings.HasSuffix(filename, issuerExt) {
			continue
		}

		data, err := os.ReadFile(filename)
		if err != nil {
			return nil, err
		}

		pCert, err := certcrypto.ParsePEMCertificate(data)
		if err != nil {
			return nil, err
		}

		name, err := certcrypto.GetCertificateMainDomain(pCert)
		if err != nil {
			return nil, err
		}

		keyData, err := os.ReadFile(strings.ReplaceAll(filename, certExt, keyExt))
		if err != nil {
			slog.Error("get cert key failed", "err", err)
			continue
		}

		certs = append(certs, cert{
			Certificate: pCert,
			rawCert:     data,
			rawKey:      keyData,
			name:        name,
		})
	}

	return certs, nil
}

func saveCertificate(dataDir string, cert *certificate.Resource) error {
	certDir := filepath.Join(dataDir, "certificates")

	if _, err := os.Stat(certDir); errors.Is(err, os.ErrNotExist) {
		if err := os.MkdirAll(certDir, 0755); err != nil {
			return err
		}
	}

	domain := cert.Domain

	if err := os.WriteFile(filepath.Join(certDir, domain+certExt), cert.Certificate, 0600); err != nil {
		return err
	}

	if cert.IssuerCertificate != nil {
		if err := os.WriteFile(filepath.Join(certDir, domain+issuerExt), cert.IssuerCertificate, 0600); err != nil {
			return err
		}
	}

	if cert.PrivateKey != nil {
		if err := os.WriteFile(filepath.Join(certDir, domain+keyExt), cert.PrivateKey, 0600); err != nil {
			return err
		}
	}

	jsonBytes, err := json.MarshalIndent(cert, "", "\t")
	if err != nil {
		return err
	}

	if err = os.WriteFile(filepath.Join(certDir, domain+resourceExt), jsonBytes, 0600); err != nil {
		return err
	}

	slog.Info("saved certificate", "domain", cert.Domain)

	return nil
}
