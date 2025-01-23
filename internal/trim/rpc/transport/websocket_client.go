/*
 * Copyright 2025 CosPotato Lab.
 * Author: CosPotato Lin<i@0x233.cn>
 */

package transport

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/cospotato/fnos-acme/internal/trim/rpc/codes"
	"github.com/gorilla/websocket"
)

type webSocketClient struct {
	ctx     context.Context
	cancel  context.CancelFunc
	address string
	conn    *websocket.Conn
	topts   Options

	keepalivePong chan any
	keepaliveDone chan any

	mu             sync.Mutex
	backID         string
	nextID         uint16
	activeRequests map[string]*ClientRequest

	logger *slog.Logger
}

func dial(ctx context.Context, addr string) (*websocket.Conn, error) {
	dialer := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 45 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	conn, _, err := dialer.DialContext(ctx, addr, nil)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func NewWebSocketClient(ctx context.Context, addr string, opts Options) (_ ClientTransport, err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		if err != nil {
			cancel()
		}
	}()

	conn, err := dial(ctx, addr)
	if err != nil {
		return nil, err
	}

	defer func(conn *websocket.Conn) {
		if err != nil {
			conn.Close()
		}
	}(conn)

	t := &webSocketClient{
		ctx:            ctx,
		cancel:         cancel,
		address:        addr,
		conn:           conn,
		topts:          opts,
		backID:         "0000000000000000",
		nextID:         1,
		activeRequests: make(map[string]*ClientRequest),
		logger:         slog.Default().WithGroup(fmt.Sprintf("[client-transport %s]", addr)),
		keepalivePong:  make(chan any),
		keepaliveDone:  make(chan any),
	}

	go t.reader()
	defer func() {
		if err != nil {
			t.Close(err)
		}
	}()

	go t.keepalive()

	return t, nil
}

func (t *webSocketClient) Close(err error) {
	if err != nil {
		t.logger.Error("closing with error", "err", err)
	}

	t.cancel()

	<-t.keepaliveDone

	if err := t.conn.Close(); err != nil {
		t.logger.Error("closing websocket failed", "err", err)
	}
}

func (t *webSocketClient) RemoteAddr() string {
	return t.address
}

func (t *webSocketClient) genReqId() string {
	t.mu.Lock()
	defer t.mu.Unlock()

	id := t.nextID
	t.nextID += 1

	return fmt.Sprintf("%08x%s%04x", time.Now().Unix(), t.backID, id)
}

func (t *webSocketClient) newRequest(ctx context.Context, method string) *ClientRequest {
	s := &ClientRequest{
		Request: &Request{
			method: method,
		},
		ct:   t,
		done: make(chan struct{}),
	}

	s.ctx = ctx

	return s
}

func (t *webSocketClient) NewRequest(ctx context.Context, method string) (*ClientRequest, error) {
	r := t.newRequest(ctx, method)
	r.id = t.genReqId()

	t.mu.Lock()
	t.activeRequests[r.id] = r
	t.mu.Unlock()

	return r, nil
}

func (t *webSocketClient) SetBackID(backId string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.backID = backId
}

func (t *webSocketClient) write(data []byte, opts *WriteOptions) (err error) {
	if t.topts.TransportCredentials != nil {
		switch {
		case opts.Encrypt:
			data, err = t.topts.TransportCredentials.GetEncryptedRequest(data)
		case opts.Sign:
			data, err = t.topts.TransportCredentials.Sign(data)
		}
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	t.logger.Debug("write", "data", string(data))

	return t.conn.WriteMessage(websocket.TextMessage, data)
}

func (t *webSocketClient) handlePing() {
	t.keepalivePong <- struct{}{}
}

func (t *webSocketClient) handleResponse(data []byte) error {
	t.logger.Debug("handle response", "data", string(data))

	var hdr ResponseHeader
	if err := json.Unmarshal(data, &hdr); err != nil {
		return err
	}

	t.mu.Lock()
	req, ok := t.activeRequests[hdr.ReqId]
	delete(t.activeRequests, hdr.ReqId)
	t.mu.Unlock()

	if !ok {
		t.logger.Error("Unrecognized reqid", "reqid", hdr.ReqId)
		return nil
	}

	if hdr.ErrNo == codes.OK {
		req.data = data
	} else {
		req.err = errors.New(hdr.ErrNo.String())
	}

	close(req.done)

	return nil
}

func (t *webSocketClient) handleTaskInfo(taskInfo string) {
	if t.logger != nil {
		t.logger.Info("received task info", "taskInfo", taskInfo)
	}
}

func (t *webSocketClient) handleNotify(notify Notify) {
	if t.topts.NotifyHandler != nil {
		t.topts.NotifyHandler(notify)
	}
}

func (t *webSocketClient) reader() {
	var errClose error
	defer func() {
		if errClose != nil {
			t.Close(errClose)
		}
	}()

	for {
		typ, r, err := t.conn.NextReader()
		if err != nil {
			errClose = err
			return
		}

		if typ != websocket.TextMessage {
			t.logger.Error("received invalid message type", "type", typ)
			continue
		}

		data, err := io.ReadAll(r)
		if err != nil {
			t.logger.Error("read message failed", "err", err)
			continue
		}

		var msg Message
		if err := json.Unmarshal(data, &msg); err != nil {
			t.logger.Error("unmarshal message failed", "err", err)
			continue
		}

		switch {
		case msg.Res == "pong":
			t.handlePing()
		case msg.TaskInfo != "":
			t.handleTaskInfo(msg.TaskInfo)
		case msg.Notify != "":
			t.handleNotify(msg.Notify)
		case msg.DeviceNotify != "":
			t.handleNotify(msg.DeviceNotify)
		case msg.SysNotify != "":
			t.handleNotify(msg.SysNotify)
		default:
			if err := t.handleResponse(data); err != nil {
				t.logger.Error("handle response failed", "err", err)
			}
		}
	}
}

func (t *webSocketClient) keepalive() {
	var err error
	defer func() {
		close(t.keepaliveDone)
		if err != nil {
			t.Close(err)
		}
	}()

	type request struct {
		Req string `json:"req"`
	}

	ticker := time.NewTicker(15 * time.Second)

	for {
		select {
		case <-ticker.C:
			t.mu.Lock()
			if err := t.conn.WriteJSON(&request{Req: "ping"}); err != nil {
				t.logger.Error("send keepalive failed", "err", err)
				t.mu.Unlock()
				continue
			}
			t.mu.Unlock()

			timer := time.NewTimer(time.Minute)

			select {
			case <-t.keepalivePong:
				timer.Stop()
			case <-timer.C:
				err = fmt.Errorf("wait server pong timeout after 1 minute")
				return
			}
		case <-t.ctx.Done():
			ticker.Stop()
			return
		}
	}
}
