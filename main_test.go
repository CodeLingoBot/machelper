package main

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockCommand struct {
	mock.Mock
}

func (m *MockCommand) runCommand(cmd string, rest ...string) ([]string, error) {
	args := m.Called(cmd, rest)
	return args.Get(0).([]string), args.Error(1)
}

func TestGetApplications(t *testing.T) {
	tests := []struct {
		name        string
		shouldError bool
	}{
		{"Success", false},
		{"Fail", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := new(MockCommand)
			apps := []string{"app1", "app2"}
			if tt.shouldError {
				cmd.On("runCommand", "ls", []string{"/Applications"}).Return([]string{}, errors.New("some error"))
			} else {
				cmd.On("runCommand", "ls", []string{"/Applications"}).Return(apps, nil)
			}

			m := MacHelper{cm: cmd}
			res, err := m.getApplications()
			if tt.shouldError {
				require.Len(t, res, 0)
				require.Error(t, err)
			} else {
				require.ElementsMatch(t, res, apps)
				require.NoError(t, err)
			}

			cmd.AssertExpectations(t)
		})
	}

}
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
