package analyzer

import (
	"testing"
)

func TestGetExtByShebang(t *testing.T) {
	tests := []struct {
		shebang string
		want    string
	}{
		{"#!/bin/sh", "sh"},
		{"#!/usr/bin/env python3", "py"},
		{"#!/usr/bin/bash", "bash"},
		{"#!/usr/bin/perl", "pl"},
		{"#!/usr/bin/ruby", "rb"},
		{"#!/usr/bin/env node", "js"},
		{"#!/usr/bin/php", "php"},
		{"#!/usr/bin/env python", "py"},
		{"#!/usr/bin/env zsh", "zsh"},
		{"#!/usr/bin/lua", "lua"},
		{"#!/usr/bin/env groovy", "groovy"},
		{"#!/usr/bin/env ksh", "ksh"},
		{"#!/usr/bin/fish", "fish"},
		{"#!/usr/bin/env awk", "awk"},
	}

	for _, tt := range tests {
		t.Run(tt.shebang, func(t *testing.T) {
			got, _ := GetExtByShebang(tt.shebang)

			if tt.want != got {
				t.Errorf("Expected %s got %s", tt.want, got)
				t.Fail()
			}
		})
	}
}
