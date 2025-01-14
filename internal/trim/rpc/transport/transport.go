/*
 * Copyright 2025 CosPotato Lab.
 * Author: CosPotato Lin<i@0x233.cn>
 */

package transport

import (
	"context"

	"github.com/cospotato/fnos-acme/internal/trim/rpc/codes"
	"github.com/cospotato/fnos-acme/internal/trim/rpc/credentials"
)

type ClientTransport interface {
	Close(err error)
	RemoteAddr() string
	NewRequest(ctx context.Context, method string) (*ClientRequest, error)
	SetBackID(backId string)
}

type Request struct {
	id     string
	ctx    context.Context
	method string
}

type ResponseHeader struct {
	ReqId  string     `json:"reqid"`
	Result string     `json:"result"`
	ErrNo  codes.Code `json:"errno,omitempty"`
}

type Options struct {
	TransportCredentials credentials.TransportCredentials
}

type WriteOptions struct {
	Sign    bool
	Encrypt bool
}