package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_getAppNameFromCaskInfo(t *testing.T) {

	tests := []struct {
		name string
		info []string
		want string
	}{
		{
			"empty",
			[]string{},
			"",
		},
		{
			"found",
			[]string{
				"kap: 2.2.0 (auto_updates)",
				"https://getkap.co/",
				"/usr/local/Caskroom/kap/2.0.0 (64B)",
				"From: https://github.com/Homebrew/homebrew-cask/blob/master/Casks/kap.rb",
				"==> Name",
				"Kap",
				"==> Artifacts",
				"Kap.app (App)",
			},
			"Kap.app",
		},
		{
			"found, no other info",
			[]string{
				"Example.app (App)",
			},
			"Example.app",
		},
		{
			"missing, other artifact",
			[]string{
				"mactex-20180417.pkg (Pkg)",
			},
			"",
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, getAppNameFromCaskInfo(tt.info))
		})
	}
}
