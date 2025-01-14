/*
 * Copyright 2024 CosPotato Lab.
 * Author: CosPotato Lin<i@0x233.cn>
 */

package rpc

import (
	"fmt"

	"github.com/cospotato/fnos-acme/internal/trim/rpc/codes"
)

type Error struct {
	c codes.Code
	m string
}

func (e Error) Error() string {
	return fmt.Sprintf("rpc error: code = %d desc = %s", e.c, e.m)
}

func New(c codes.Code, msg string) *Error {
	return &Error{c: c, m: msg}
}

func Newf(c codes.Code, format string, a ...any) *Error {
	return &Error{c: c, m: fmt.Sprintf(format, a...)}
}
