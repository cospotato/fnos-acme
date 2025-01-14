/*
 * Copyright 2025 CosPotato Lab.
 * Author: CosPotato Lin<i@0x233.cn>
 */

package user

type LoginRequest struct {
	User       string `json:"user,omitempty"`
	Password   string `json:"password,omitempty"`
	DeviceType string `json:"deviceType,omitempty"`
	DeviceName string `json:"deviceName,omitempty"`
	Stay       bool   `json:"stay,omitempty"`
}

type LoginResponse struct {
	UID       int    `json:"uid,omitempty"`
	Admin     bool   `json:"admin,omitempty"`
	Token     string `json:"token,omitempty"`
	Secret    string `json:"secret,omitempty"`
	BackId    string `json:"backId,omitempty"`
	MachineId string `json:"machineId,omitempty"`
}

type AuthTokenRequest struct {
	Token string `json:"token,omitempty"`
	Main  bool   `json:"main,omitempty"`
}

type AuthTokenResponse struct {
	UID    string `json:"uid,omitempty"`
	Admin  bool   `json:"admin,omitempty"`
	BackId string `json:"backId,omitempty"`
}
