/*
 * Copyright 2025 CosPotato Lab.
 * Author: CosPotato Lin<i@0x233.cn>
 */

package util

import (
	"context"

	"github.com/cospotato/fnos-acme/internal/trim/rpc"
	"github.com/cospotato/fnos-acme/internal/trim/rpc/types"
)

const (
	UtilService_GetRSAPub_FullMethodName = "util.crypto.getRSAPub"
	UtilService_GetSI_FullMethodName     = "util.getSI"
)

type UtilsService interface {
	GetRSAPub(ctx context.Context, opts ...rpc.CallOption) (*GetRSAPubResponse, error)
	GetSI(ctx context.Context, opts ...rpc.CallOption) (*GetSIResponse, error)
}

type utilsServiceClient struct {
	cc rpc.ClientConnInterface
}

func NewUtilsServiceClient(cc rpc.ClientConnInterface) UtilsService {
	return &utilsServiceClient{cc}
}

func (c *utilsServiceClient) GetRSAPub(ctx context.Context, opts ...rpc.CallOption) (*GetRSAPubResponse, error) {
	out := new(GetRSAPubResponse)
	err := c.cc.Invoke(ctx, UtilService_GetRSAPub_FullMethodName, &types.Empty{}, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *utilsServiceClient) GetSI(ctx context.Context, opts ...rpc.CallOption) (*GetSIResponse, error) {
	out := new(GetSIResponse)
	err := c.cc.Invoke(ctx, UtilService_GetSI_FullMethodName, &types.Empty{}, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}
