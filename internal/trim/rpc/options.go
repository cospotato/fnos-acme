/*
 * Copyright 2025 CosPotato Lab.
 * Author: CosPotato Lin<i@0x233.cn>
 */

package rpc

type callInfo struct {
	encrypt  bool
	skipSign bool
	si       string
}

type CallOption interface {
	before(*callInfo) error
}

type EncryptCallOption struct{}

func Encrypt() EncryptCallOption {
	return EncryptCallOption{}
}

func (EncryptCallOption) before(ci *callInfo) error {
	ci.encrypt = true
	return nil
}

type SkipSignCallOption struct{}

func SkipSign() SkipSignCallOption {
	return SkipSignCallOption{}
}

func (SkipSignCallOption) before(ci *callInfo) error {
	ci.skipSign = true
	return nil
}

type SessionCallOption struct {
	si string
}

func Session(si string) SessionCallOption {
	return SessionCallOption{
		si: si,
	}
}

func (co SessionCallOption) before(ci *callInfo) error {
	ci.si = co.si
	return nil
}
