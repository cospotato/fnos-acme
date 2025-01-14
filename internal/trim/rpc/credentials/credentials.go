/*
 * Copyright 2025 CosPotato Lab.
 * Author: CosPotato Lin<i@0x233.cn>
 */

package credentials

type TransportCredentials interface {
	GetEncryptedRequest(data []byte) ([]byte, error)
	Sign(data []byte) ([]byte, error)
}
