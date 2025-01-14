/*
 * Copyright 2025 CosPotato Lab.
 * Author: CosPotato Lin<i@0x233.cn>
 */

package transport

import (
	"bytes"
	"io"

	"github.com/tidwall/sjson"
)

type ClientRequest struct {
	*Request

	ct   *webSocketClient
	data []byte
	err  error
	done chan struct{}
}

func (s *ClientRequest) Write(data []byte, opts *WriteOptions) (err error) {
	data, err = sjson.SetBytes(data, "reqid", s.id)
	if err != nil {
		return err
	}

	data, err = sjson.SetBytes(data, "req", s.method)
	if err != nil {
		return err
	}

	return s.ct.write(data, opts)
}

func (s *ClientRequest) Done() <-chan struct{} {
	return s.done
}

func (s *ClientRequest) Reader() io.Reader {
	return bytes.NewReader(s.data)
}

func (s *ClientRequest) Error() error {
	return s.err
}
