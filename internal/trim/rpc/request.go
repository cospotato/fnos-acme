/*
 * Copyright 2025 CosPotato Lab.
 * Author: CosPotato Lin<i@0x233.cn>
 */

package rpc

import (
	"context"
	"encoding/json"

	"github.com/cospotato/fnos-acme/internal/trim/rpc/transport"
	"github.com/tidwall/sjson"
)

type ClientRequest interface {
	Context() context.Context
	SendMsg(m any) error
	RecvMsg(m any) error
}

func newClientRequest(ctx context.Context, cc *ClientConn, method string, opts ...CallOption) (_ ClientRequest, err error) {
	ci := &callInfo{}

	for _, o := range opts {
		if err := o.before(ci); err != nil {
			return nil, err
		}
	}

	cs := &clientRequest{
		cc:       cc,
		ctx:      ctx,
		method:   method,
		opts:     opts,
		callInfo: ci,
	}

	return cs, nil
}

type clientRequest struct {
	method   string
	opts     []CallOption
	callInfo *callInfo
	cc       *ClientConn
	r        *transport.ClientRequest

	ctx context.Context
}

func (cs *clientRequest) Context() context.Context {
	return cs.ctx
}

func (cs *clientRequest) SendMsg(m any) error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}

	opts := &transport.WriteOptions{}

	if cs.callInfo.encrypt {
		opts.Encrypt = true
	}

	if !cs.callInfo.skipSign {
		opts.Sign = true
	}

	if cs.callInfo.si != "" {
		data, err = sjson.SetBytes(data, "si", cs.callInfo.si)
		if err != nil {
			return err
		}
	}

	t, err := cs.cc.getTransport()
	if err != nil {
		return err
	}

	cs.r, err = t.NewRequest(cs.ctx, cs.method)
	if err != nil {
		return err
	}

	return cs.r.Write(data, opts)
}

func (cs *clientRequest) RecvMsg(m any) error {
	<-cs.r.Done()

	if err := cs.r.Error(); err != nil {
		return err
	}

	return json.NewDecoder(cs.r.Reader()).Decode(m)
}
