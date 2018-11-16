package main

import (
	"errors"
	"fmt"
	"testing"

	"github.com/davecgh/go-spew/spew"
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

func TestAuditApplications(t *testing.T) {
	tests := []struct {
		name                 string
		lsApplicationsResult []string
		masListResult        []string
		getCasksResult       []string
		getCaskInfoResult    map[string][]string
		expectedResult       map[string][]string
		shouldError          bool
	}{
		{
			name:                 "success",
			lsApplicationsResult: []string{"App 1.app", "App 2.app", "App 3.app", "Safari.app"},
			masListResult:        []string{"1171820258 App 2 (1.2.3)"},
			getCasksResult:       []string{"app3"},
			getCaskInfoResult: map[string][]string{
				"app3": []string{"asdf", "artifacts", "App 3.app (App)"}},
			expectedResult: map[string][]string{
				"user": []string{"App 1.app"},
				"brew": []string{"App 3.app"},
				"mas":  []string{"App 2.app"}},
		},
		{
			name:           "bad mas data",
			masListResult:  []string{"foo"},
			getCasksResult: []string{"app3"},
			getCaskInfoResult: map[string][]string{
				"app3": []string{"asdf", "artifacts", "App 3.app (App)"}},
			lsApplicationsResult: nil,
			shouldError:          true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := new(MockCommand)
			m := MacHelper{cm: cmd}

			if len(tt.getCasksResult) > 0 {
				cmd.On("runCommand", "brew", []string{"cask", "list"}).Return(tt.getCasksResult, nil)
			}
			if len(tt.masListResult) > 0 {
				cmd.On("runCommand", "mas", []string{"list"}).Return(tt.masListResult, nil)
			}
			if len(tt.lsApplicationsResult) > 0 {
				cmd.On("runCommand", "ls", []string{"/Applications"}).Return(tt.lsApplicationsResult, nil)
			}

			for k, v := range tt.getCaskInfoResult {
				p := []string{"cask", "info"}
				p = append(p, k)
				cmd.On("runCommand", "brew", p).Return(v, nil)
			}

			res, err := m.AuditApplications()
			if tt.shouldError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedResult, res)
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, getAppNameFromCaskInfo(tt.info))
		})
	}
}

func Test_bytesWithNewLinesToStrings(t *testing.T) {

	tests := []struct {
		info []byte
		want []string
	}{
		{
			info: []byte(""),
			want: []string{""},
		},
		{
			info: []byte("foo\nbar\n"),
			want: []string{"foo", "bar"},
		},
		{
			info: []byte("foo\nbar"),
			want: []string{"foo", "bar"},
		},
	}
	for x, tt := range tests {
		t.Run(fmt.Sprintf("%v", x), func(t *testing.T) {
			require.Equal(t, tt.want, bytesWithNewLinesToStrings(tt.info))
		})
	}
}

func TestRunSystemCommand(t *testing.T) {
	sc := SystemCommands{}
	res, err := sc.runCommand("fake", "cmd")
	require.Len(t, res, 0)
	require.Error(t, err)
	res, err = sc.runCommand("ls")
	require.True(t, len(res) > 0)
	require.NoError(t, err)
	spew.Dump(sc.runCommand("ls"))
}
