/*
 * Copyright 2024 CosPotato Lab.
 * Author: CosPotato Lin<i@0x233.cn>
 */

package trim

import (
	"context"
	"net/url"
	"runtime"

	"github.com/cospotato/fnos-acme/internal/trim/api/remoteaccess"
	"github.com/cospotato/fnos-acme/internal/trim/api/user"
	"github.com/cospotato/fnos-acme/internal/trim/api/util"
	"github.com/cospotato/fnos-acme/internal/trim/rpc"
)

type Client struct {
	conn  *rpc.ClientConn
	creds *tlsCreds

	si    string
	token string
}

func New(address, typ string) (*Client, error) {
	u, err := url.Parse(address)
	if err != nil {
		return nil, err
	}

	if u.Scheme == "https" {
		u.Scheme = "wss"
	} else {
		u.Scheme = "ws"
	}

	u.Path = "/websocket"
	q := u.Query()
	q.Add("type", typ)
	u.RawQuery = q.Encode()

	creds := NewTLS()

	conn, err := rpc.DialContext(context.Background(), u.String(), rpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, err
	}

	c := &Client{
		conn:  conn,
		creds: creds,
	}

	if err := c.preflight(); err != nil {
		return c, err
	}

	return c, nil
}

func NewMainClient(address string) (*Client, error) {
	return New(address, "main")
}

func NewTimerClient(address string) (*Client, error) {
	return New(address, "timer")
}

func (c *Client) preflight() error {
	resp, err := c.Main().UtilService().GetRSAPub(context.TODO(), rpc.SkipSign())
	if err != nil {
		return err
	}

	c.si = resp.SI

	if err := c.creds.SetPublicKey(resp.Pub); err != nil {
		return err
	}

	return nil
}

func (c *Client) Login(ctx context.Context, req *user.LoginRequest) (*user.LoginResponse, error) {
	req.DeviceType = "client-go"
	req.DeviceName = runtime.GOOS + "-Client"
	req.Stay = true

	resp, err := c.Main().UserService().Login(ctx, req, rpc.Encrypt(), rpc.Session(c.si))
	if err != nil {
		return nil, err
	}

	if err := c.creds.SetSecret(resp.Secret); err != nil {
		return resp, err
	}

	c.conn.SetBackID(resp.BackId)
	c.token = resp.Token

	return resp, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) Main() *MainClient {
	return &MainClient{
		c: c,
	}
}

type MainClient struct {
	c *Client
}

func (mc *MainClient) UserService() user.UserService {
	return user.NewUserServiceClient(mc.c.conn)
}

func (mc *MainClient) UtilService() util.UtilsService {
	return util.NewUtilsServiceClient(mc.c.conn)
}

func (mc *MainClient) RemoteAccessService() remoteaccess.RemoteAccessService {
	return remoteaccess.NewRemoteAccessServiceClient(mc.c.conn)
}
