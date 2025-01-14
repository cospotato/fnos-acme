/*
 * Copyright 2025 CosPotato Lab.
 * Author: CosPotato Lin<i@0x233.cn>
 */

package main

import (
	"slices"
)

func domainsEqual(a, b []string) bool {
	for _, s := range a {
		if !slices.Contains(b, s) {
			return false
		}
	}

	for _, s := range b {
		if !slices.Contains(a, s) {
			return false
		}
	}

	return true
}

func merge(prevDomains, nextDomains []string) []string {
	for _, next := range nextDomains {
		if slices.Contains(prevDomains, next) {
			continue
		}

		prevDomains = append(prevDomains, next)
	}

	return prevDomains
}
