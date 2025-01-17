/*
 * Copyright 2025 CosPotato Lab.
 * Author: CosPotato Lin<i@0x233.cn>
 */

package main

import (
	"bytes"
	"crypto"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"os"
	"path/filepath"

	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/registration"
)

type account struct {
	Email        string                 `json:"email"`
	Registration *registration.Resource `json:"registration"`
	RawKey       string                 `json:"privateKey"`
	key          crypto.PrivateKey
}

func (a *account) GetEmail() string {
	return a.Email
}

func (a *account) GetRegistration() *registration.Resource {
	return a.Registration
}

func (a *account) GetPrivateKey() crypto.PrivateKey {
	return a.key
}

func createAccount(email string) (*account, error) {
	privateKey, err := certcrypto.GeneratePrivateKey(defaultKeyType)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)

	pemKey := certcrypto.PEMBlock(privateKey)
	err = pem.Encode(buf, pemKey)
	if err != nil {
		return nil, err
	}

	return &account{
		Email:  email,
		RawKey: buf.String(),
		key:    privateKey,
	}, nil
}

func setupAccount(dataDir, email string) (*account, error) {
	_, err := os.Stat(filepath.Join(dataDir, accountJson))
	if errors.Is(err, os.ErrNotExist) {
		return createAccount(email)
	}

	f, err := os.Open(filepath.Join(dataDir, accountJson))
	if err != nil {
		return nil, err
	}

	accounts := make(map[string]*account)

	if err := json.NewDecoder(f).Decode(&accounts); err != nil {
		return nil, err
	}

	acc, ok := accounts[email]
	if !ok {
		return createAccount(email)
	}

	if acc.RawKey != "" {
		keyBlock, _ := pem.Decode([]byte(acc.RawKey))

		acc.key, err = x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
		if err != nil {
			return nil, err
		}
	}

	return acc, nil
}

func loadAccount(dataDir string) (map[string]*account, error) {
	f, err := os.Open(filepath.Join(dataDir, accountJson))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return make(map[string]*account), nil
		}

		return nil, err
	}

	defer f.Close()

	accounts := make(map[string]*account)

	if err := json.NewDecoder(f).Decode(&accounts); err != nil {
		return nil, err
	}

	return accounts, nil
}

func saveAccount(dataDir string, acc *account) error {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return err
	}

	accounts, err := loadAccount(dataDir)
	if err != nil {
		return err
	}

	accounts[acc.Email] = acc

	f, err := os.OpenFile(filepath.Join(dataDir, accountJson), os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}

	return json.NewEncoder(f).Encode(accounts)
}
