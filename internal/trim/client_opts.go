/*
 * Copyright 2025 CosPotato Lab.
 * Author: CosPotato Lin<i@0x233.cn>
 */

package trim

type clientOpts struct {
	username string
	password string
}

type Opt func(*clientOpts) error

func WithLogin(username, password string) Opt {
	return func(opts *clientOpts) error {
		opts.username = username
		opts.password = password
		return nil
	}
}
