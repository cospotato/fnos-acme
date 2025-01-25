/*
 * Copyright 2024 CosPotato Lab.
 * Author: CosPotato Lin<i@0x233.cn>
 */

package trim

import (
	"context"
	"log/slog"
	"net/url"
	"runtime"
	"sync"
	"time"

	"github.com/cospotato/fnos-acme/internal/trim/api/remoteaccess"
	"github.com/cospotato/fnos-acme/internal/trim/api/user"
	"github.com/cospotato/fnos-acme/internal/trim/api/util"
	"github.com/cospotato/fnos-acme/internal/trim/rpc"
	"github.com/cospotato/fnos-acme/internal/trim/rpc/transport"
)

type Client struct {
	co        clientOpts
	connMu    sync.Mutex
	conn      *rpc.ClientConn
	connector func() (*rpc.ClientConn, error)
	creds     *tlsCreds

	si    string
	token string
}

func New(address, typ string, opts ...Opt) (*Client, error) {
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

	var co clientOpts

	for _, o := range opts {
		o(&co)
	}

	c := &Client{
		co:    co,
		creds: NewTLS(),
	}

	connector := func() (*rpc.ClientConn, error) {
		return rpc.DialContext(context.Background(), u.String(), rpc.WithTransportCredentials(c.creds), rpc.WithNotifyHandler(c.notifyHandler))
	}

	conn, err := connector()
	if err != nil {
		return nil, err
	}

	c.conn = conn
	c.connector = connector

	if err := c.preflight(context.Background()); err != nil {
		return c, err
	}

	go c.keepalive()

	return c, nil
}

func NewMainClient(address string, opts ...Opt) (*Client, error) {
	return New(address, "main", opts...)
}

func NewTimerClient(address string, opts ...Opt) (*Client, error) {
	return New(address, "timer", opts...)
}

func (c *Client) preflight(ctx context.Context) error {
	resp, err := c.Main().UtilService().GetRSAPub(ctx, rpc.SkipSign())
	if err != nil {
		return err
	}

	c.si = resp.SI

	if err := c.creds.SetPublicKey(resp.Pub); err != nil {
		return err
	}

	return c.login(ctx)
}

func (c *Client) authToken(ctx context.Context) error {
	req := &user.AuthTokenRequest{
		Token: c.token,
	}

	resp, err := c.Main().UserService().AuthToken(ctx, req, rpc.Encrypt(), rpc.Session(c.si))
	if err != nil {
		return err
	}

	c.conn.SetBackID(resp.BackId)

	return nil
}

func (c *Client) login(ctx context.Context) error {
	if c.token != "" {
		return c.authToken(ctx)
	}

	req := &user.LoginRequest{
		User:       c.co.username,
		Password:   c.co.password,
		DeviceType: "client-go",
		DeviceName: runtime.GOOS + "-Client",
		Stay:       true,
	}

	resp, err := c.Main().UserService().Login(ctx, req, rpc.Encrypt(), rpc.Session(c.si))
	if err != nil {
		return err
	}

	if err := c.creds.SetSecret(resp.Secret); err != nil {
		return err
	}

	c.conn.SetBackID(resp.BackId)
	c.token = resp.Token

	return nil
}

func (c *Client) Reconnect(ctx context.Context) error {
	c.connMu.Lock()
	defer c.connMu.Unlock()
	c.conn.Close()

	conn, err := c.connector()
	if err != nil {
		return err
	}

	c.conn = conn
	return c.preflight(ctx)
}

func (c *Client) notifyHandler(notify transport.Notify) {
	switch notify {
	case transport.TokenExpired, transport.PrivilegeChanged:
		slog.Debug("handle token expired/privilege changed")

		go func() {
			if err := c.Reconnect(context.Background()); err != nil {
				slog.Error("reconnect failed", "err", err)
			}
		}()
	default:
		slog.Error("unknown notify", "notify", notify)
	}
}

func (c *Client) keepalive() {
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		if _, err := c.Main().UserService().Active(ctx, &user.ActiveRequest{}); err != nil {
			slog.Error("keepalive failed", "err", err)

			if err := c.Reconnect(ctx); err != nil {
				slog.Error("reconnect failed", "err", err)
			}
		}
		cancel()

		time.Sleep(time.Minute)
	}
}

func (c *Client) Close() error {
	c.connMu.Lock()
	defer c.connMu.Unlock()

	if c.conn == nil {
		return nil
	}

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
	mc.c.connMu.Lock()
	defer mc.c.connMu.Unlock()
	return user.NewUserServiceClient(mc.c.conn)
}

func (mc *MainClient) UtilService() util.UtilsService {
	mc.c.connMu.Lock()
	defer mc.c.connMu.Unlock()
	return util.NewUtilsServiceClient(mc.c.conn)
}

func (mc *MainClient) RemoteAccessService() remoteaccess.RemoteAccessService {
	mc.c.connMu.Lock()
	defer mc.c.connMu.Unlock()
	return remoteaccess.NewRemoteAccessServiceClient(mc.c.conn)
}
