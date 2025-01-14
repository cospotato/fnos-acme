/*
 * Copyright 2025 CosPotato Lab.
 * Author: CosPotato Lin<i@0x233.cn>
 */

package util

type GetRSAPubResponse struct {
	Pub string `json:"pub,omitempty"`
	SI  string `json:"si,omitempty"`
}

type GetSIResponse struct {
	SI string `json:"si,omitempty"`
}
