/*
 * Copyright 2025 CosPotato Lab.
 * Author: CosPotato Lin<i@0x233.cn>
 */

package transport

type Notify string

const (
	FileFav          Notify = "fileFav"
	Liveupdate       Notify = "liveupdate"
	PrivilegeChanged Notify = "privilegeChanged"
	TokenExpired     Notify = "tokenExpired"
)
