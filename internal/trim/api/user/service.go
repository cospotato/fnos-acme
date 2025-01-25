/*
 * Copyright 2025 CosPotato Lab.
 * Author: CosPotato Lin<i@0x233.cn>
 */

package user

import (
	"context"

	"github.com/cospotato/fnos-acme/internal/trim/rpc"
)

const (
	UserService_Login_FullMethodName     = "user.login"
	UserService_AuthToken_FullMethodName = "user.authToken"
	UserService_Active_FullMethodName    = "user.active"
)

type UserService interface {
	Login(ctx context.Context, in *LoginRequest, opts ...rpc.CallOption) (*LoginResponse, error)
	AuthToken(ctx context.Context, in *AuthTokenRequest, opts ...rpc.CallOption) (*AuthTokenResponse, error)
	Active(ctx context.Context, in *ActiveRequest, opts ...rpc.CallOption) (*ActiveResponse, error)
}

type userServiceClient struct {
	cc rpc.ClientConnInterface
}

func NewUserServiceClient(cc rpc.ClientConnInterface) UserService {
	return &userServiceClient{cc: cc}
}

func (c *userServiceClient) Login(ctx context.Context, in *LoginRequest, opts ...rpc.CallOption) (*LoginResponse, error) {
	out := new(LoginResponse)
	err := c.cc.Invoke(ctx, UserService_Login_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userServiceClient) AuthToken(ctx context.Context, in *AuthTokenRequest, opts ...rpc.CallOption) (*AuthTokenResponse, error) {
	out := new(AuthTokenResponse)
	err := c.cc.Invoke(ctx, UserService_AuthToken_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userServiceClient) Active(ctx context.Context, in *ActiveRequest, opts ...rpc.CallOption) (*ActiveResponse, error) {
	out := new(ActiveResponse)
	err := c.cc.Invoke(ctx, UserService_Active_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}
