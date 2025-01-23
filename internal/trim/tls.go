/*
 * Copyright 2025 CosPotato Lab.
 * Author: CosPotato Lin<i@0x233.cn>
 */

package trim

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	mathrand "math/rand"
	"time"

	"github.com/cospotato/fnos-acme/internal/trim/rpc/credentials"
)

func parsePublicKey(pubKeyStr string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pubKeyStr))
	if block == nil || block.Type != "PUBLIC KEY" {
		return nil, errors.New("failed to decode PEM block containing public key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("not an RSA public key")
	}

	return rsaPub, nil
}

func hmacSha256(secret []byte, data []byte) []byte {
	h := hmac.New(sha256.New, secret)
	h.Write(data)

	return []byte(base64.StdEncoding.EncodeToString(h.Sum(nil)))
}

func pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padtext...)
}

func randomString(n int) string {
	const letters = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	r := mathrand.New(mathrand.NewSource(time.Now().UnixNano()))

	result := make([]byte, n)
	for i := range result {
		result[i] = letters[r.Intn(len(letters))]
	}

	return string(result)
}

type tlsCreds struct {
	aesKey    []byte
	aesIv     []byte
	publicKey *rsa.PublicKey
	secret    []byte
}

// TLS_RSA_WITH_AES_256_CBC_SHA256
func NewTLS() *tlsCreds {
	tc := &tlsCreds{
		aesKey: []byte(randomString(32)),
		aesIv:  make([]byte, aes.BlockSize),
	}

	_, _ = io.ReadFull(rand.Reader, tc.aesIv)

	return tc
}

var _ credentials.TransportCredentials = (*tlsCreds)(nil)

func (tc *tlsCreds) encryptRSA(data []byte) (string, error) {
	ciphertext, err := rsa.EncryptPKCS1v15(rand.Reader, tc.publicKey, data)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt data: %v", err)
	}

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (tc *tlsCreds) encryptAES(data []byte) (string, error) {
	block, err := aes.NewCipher(tc.aesKey)
	if err != nil {
		return "", fmt.Errorf("failed to create AES cipher: %v", err)
	}

	mode := cipher.NewCBCEncrypter(block, tc.aesIv)
	paddedData := pad(data, aes.BlockSize)

	ciphertext := make([]byte, len(paddedData))
	mode.CryptBlocks(ciphertext, paddedData)

	encryptedData := base64.StdEncoding.EncodeToString(ciphertext)

	return encryptedData, nil
}

func (tc *tlsCreds) decryptAES(data string) ([]byte, error) {
	block, err := aes.NewCipher(tc.aesKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %v", err)
	}

	mode := cipher.NewCBCDecrypter(block, tc.aesIv)

	ciphertext, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 data: %v", err)
	}

	decryptedData := make([]byte, len(ciphertext))
	mode.CryptBlocks(decryptedData, ciphertext)

	return decryptedData, nil
}

func (tc *tlsCreds) getEncodeParams(data []byte) ([]byte, error) {
	rsa, err := tc.encryptRSA(tc.aesKey)
	if err != nil {
		return nil, err
	}

	aes, err := tc.encryptAES(data)
	if err != nil {
		return nil, err
	}

	return json.Marshal(map[string]string{
		"req": "encrypted",
		"iv":  base64.StdEncoding.EncodeToString(tc.aesIv),
		"rsa": rsa,
		"aes": aes,
	})
}

func (tc *tlsCreds) setSecret(secret string) error {
	data, err := tc.decryptAES(secret)
	if err != nil {
		return err
	}

	tc.secret = bytes.Clone(data[:16])

	return nil
}

func (tc *tlsCreds) SetPublicKey(data string) (err error) {
	tc.publicKey, err = parsePublicKey(data)
	return
}

func (tc *tlsCreds) SetSecret(data string) error {
	return tc.setSecret(data)
}

func (tc *tlsCreds) GetEncryptedRequest(data []byte) ([]byte, error) {
	return tc.getEncodeParams(data)
}

func (tc *tlsCreds) Sign(data []byte) ([]byte, error) {
	return append(hmacSha256(tc.secret, data), data...), nil
}
