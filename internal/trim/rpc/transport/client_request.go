/*
 * Copyright 2025 CosPotato Lab.
 * Author: CosPotato Lin<i@0x233.cn>
 */

package transport

import (
	"bytes"
	"fmt"
	"io"
)

type ClientRequest struct {
	*Request

	ct   *webSocketClient
	data []byte
	err  error
	done chan struct{}
}

func (s *ClientRequest) Write(data []byte, opts *WriteOptions) (err error) {
	hdr := []byte(fmt.Sprintf(`{"reqid":"%s","req":"%s"`, s.id, s.method))

	// just `{}`
	if len(data) == 2 {
		data = append(hdr, '}')
	} else {
		data = bytes.Join([][]byte{hdr, data[1:]}, []byte(","))
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
