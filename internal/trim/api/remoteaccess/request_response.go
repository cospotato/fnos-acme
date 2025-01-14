/*
 * Copyright 2025 CosPotato Lab.
 * Author: CosPotato Lin<i@0x233.cn>
 */

package remoteaccess

type Cert struct {
	ID          int    `json:"id"`
	Domain      string `json:"domain"`
	EncryptType string `json:"encryptType"`
	IssuedBy    string `json:"issuedBy"`
	San         string `json:"SAN"`
	Status      string `json:"status"`
	ValidFrom   int64  `json:"validFrom"`
	ValidTo     int64  `json:"validTo"`
	Desc        string `json:"desc"`
	IsDefault   int    `json:"isDefault"`
	Source      string `json:"source"`
}

type GetCertListResponse struct {
	Data []Cert `json:"data"`
}

type CertRequestData struct {
	ID                      int    `json:"id"`
	Desc                    string `json:"desc"`
	PrivateKeyBase64        string `json:"privateKeyBase64"`
	CertificateBase64       string `json:"certificateBase64"`
	IssuerCertificateBase64 string `json:"issuerCertificateBase64"`
	IsDefault               int    `json:"isDefault"`
}

type UploadCertRequest struct {
	Data CertRequestData `json:"data"`
}

type UploadCertResponse struct {
	Data bool `json:"data"`
}

type ReplaceCertRequest struct {
	Data CertRequestData `json:"data"`
}

type ReplaceCertResponse struct {
	Data bool `json:"data"`
}
