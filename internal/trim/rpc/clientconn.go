/*
 * Copyright 2025 CosPotato Lab.
 * Author: CosPotato Lin<i@0x233.cn>
 */

package rpc

import (
	"context"

	"github.com/cospotato/fnos-acme/internal/trim/rpc/transport"
)

type ClientConnInterface interface {
	Invoke(ctx context.Context, method string, args, reply any, opts ...CallOption) error
}

type ClientConn struct {
	ctx    context.Context
	cancel context.CancelFunc
	t      transport.ClientTransport
	dopts  dialOptions

	target string
}

func NewClient(target string, opts ...DialOption) (conn *ClientConn, err error) {
	cc := &ClientConn{target: target}

	cc.ctx, cc.cancel = context.WithCancel(context.Background())

	for _, o := range opts {
		o.apply(&cc.dopts)
	}

	return cc, nil
}

func DialContext(ctx context.Context, target string, opts ...DialOption) (_ *ClientConn, err error) {
	cc, err := NewClient(target, opts...)
	if err != nil {
		return nil, err
	}

	// TODO: lazy dial websocket
	cc.t, err = transport.NewWebSocketClient(ctx, cc.target, cc.dopts.topts)
	if err != nil {
		return nil, err
	}

	return cc, nil
}

func Dial(target string, opts ...DialOption) (*ClientConn, error) {
	return DialContext(context.Background(), target, opts...)
}

func (cc *ClientConn) getTransport() (transport.ClientTransport, error) {
	return cc.t, nil
}

func (cc *ClientConn) SetBackID(backId string) {
	cc.t.SetBackID(backId)
}

func (cc *ClientConn) Close() error {
	cc.t.Close(nil)
	return nil
}
