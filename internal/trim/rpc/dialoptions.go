/*
 * Copyright 2025 CosPotato Lab.
 * Author: CosPotato Lin<i@0x233.cn>
 */

package rpc

import (
	"github.com/cospotato/fnos-acme/internal/trim/rpc/credentials"
	"github.com/cospotato/fnos-acme/internal/trim/rpc/transport"
)

type dialOptions struct {
	topts transport.Options
}

type DialOption interface {
	apply(*dialOptions)
}

type funcDialOption struct {
	f func(*dialOptions)
}

func (fdo *funcDialOption) apply(do *dialOptions) {
	fdo.f(do)
}

func newFuncDialOption(f func(*dialOptions)) *funcDialOption {
	return &funcDialOption{f: f}
}

func WithTransportCredentials(creds credentials.TransportCredentials) DialOption {
	return newFuncDialOption(func(do *dialOptions) {
		do.topts.TransportCredentials = creds
	})
}
