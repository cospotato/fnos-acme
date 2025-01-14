/*
 * Copyright 2025 CosPotato Lab.
 * Author: CosPotato Lin<i@0x233.cn>
 */

package rpc

import "context"

func (cc *ClientConn) Invoke(ctx context.Context, method string, req, reply any, opts ...CallOption) error {
	return invoke(ctx, method, req, reply, cc, opts...)
}

func invoke(ctx context.Context, method string, req, reply any, cc *ClientConn, opts ...CallOption) error {
	cs, err := newClientRequest(ctx, cc, method, opts...)
	if err != nil {
		return err
	}
	if err := cs.SendMsg(req); err != nil {
		return err
	}
	return cs.RecvMsg(reply)
}
