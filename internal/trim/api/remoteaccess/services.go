/*
 * Copyright 2025 CosPotato Lab.
 * Author: CosPotato Lin<i@0x233.cn>
 */

package remoteaccess

import (
	"context"

	"github.com/cospotato/fnos-acme/internal/trim/rpc"
	"github.com/cospotato/fnos-acme/internal/trim/rpc/types"
)

const (
	RemoteAccessService_UploadCert_FullMethodName  = "appcgi.netsvr.cert.upload"
	RemoteAccessService_ReplaceCert_FullMethodName = "appcgi.netsvr.cert.replace"
	RemoteAccessService_GetCertList_FullMethodName = "appcgi.netsvr.cert.list"
)

type RemoteAccessService interface {
	UploadCert(ctx context.Context, in *UploadCertRequest, opts ...rpc.CallOption) (*UploadCertResponse, error)
	ReplaceCert(ctx context.Context, in *ReplaceCertRequest, opts ...rpc.CallOption) (*ReplaceCertResponse, error)
	GetCertList(ctx context.Context, opts ...rpc.CallOption) (*GetCertListResponse, error)
}

type remoteAccessServiceClient struct {
	cc *rpc.ClientConn
}

func NewRemoteAccessServiceClient(cc *rpc.ClientConn) RemoteAccessService {
	return &remoteAccessServiceClient{
		cc: cc,
	}
}

func (c *remoteAccessServiceClient) UploadCert(ctx context.Context, in *UploadCertRequest, opts ...rpc.CallOption) (*UploadCertResponse, error) {
	out := new(UploadCertResponse)
	err := c.cc.Invoke(ctx, RemoteAccessService_UploadCert_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *remoteAccessServiceClient) ReplaceCert(ctx context.Context, in *ReplaceCertRequest, opts ...rpc.CallOption) (*ReplaceCertResponse, error) {
	out := new(ReplaceCertResponse)
	err := c.cc.Invoke(ctx, RemoteAccessService_ReplaceCert_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *remoteAccessServiceClient) GetCertList(ctx context.Context, opts ...rpc.CallOption) (*GetCertListResponse, error) {
	out := new(GetCertListResponse)
	err := c.cc.Invoke(ctx, RemoteAccessService_GetCertList_FullMethodName, &types.Empty{}, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}
