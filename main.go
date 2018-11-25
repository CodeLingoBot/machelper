package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strings"
	"sync"
)

var defaultApps = map[string]struct{}{
	"App Store.app":          struct{}{},
	"Automator.app":          struct{}{},
	"Books.app":              struct{}{},
	"Calculator.app":         struct{}{},
	"Calendar.app":           struct{}{},
	"Chess.app":              struct{}{},
	"Contacts.app":           struct{}{},
	"Dashboard.app":          struct{}{},
	"Dictionary.app":         struct{}{},
	"FaceTime.app":           struct{}{},
	"Font Book.app":          struct{}{},
	"Home.app":               struct{}{},
	"Image Capture.app":      struct{}{},
	"iTunes.app":             struct{}{},
	"Launchpad.app":          struct{}{},
	"Mail.app":               struct{}{},
	"Maps.app":               struct{}{},
	"Messages.app":           struct{}{},
	"Mission Control.app":    struct{}{},
	"News.app":               struct{}{},
	"Notes.app":              struct{}{},
	"Photo Booth.app":        struct{}{},
	"Photos.app":             struct{}{},
	"Preview.app":            struct{}{},
	"QuickTime Player.app":   struct{}{},
	"Reminders.app":          struct{}{},
	"Safari.app":             struct{}{},
	"Siri.app":               struct{}{},
	"Stickies.app":           struct{}{},
	"Stocks.app":             struct{}{},
	"System Preferences.app": struct{}{},
	"TextEdit.app":           struct{}{},
	"Time Machine.app":       struct{}{},
	"Utilities":              struct{}{},
	"Voice Memos.app":        struct{}{},
	"WebStorm.app":           struct{}{},
}

func main() {
	a := MacHelper{
		cm: &SystemCommands{},
	}
	audit, _ := a.AuditApplications()
	auditJSON, _ := json.Marshal(audit)
	fmt.Println(string(auditJSON))
}

//BrewAudit holds audit results
type BrewAudit struct {
	Names []string `json:"names,omitempty"`
}

//BrewCaskAudit holds audit results
type BrewCaskAudit struct {
	Names     []string `json:"names,omitempty"`
	Artifacts []string `json:"artifacts,omitempty"`
}

//AppStoreAudit holds audit results
type AppStoreAudit struct {
	Artifacts []string `json:"artifacts,omitempty"`
}

//UserInstalledAudit holds audit results
type UserInstalledAudit struct {
	Artifacts []string `json:"artifacts,omitempty"`
}

//ApplicationAudit holds audit results
type ApplicationAudit struct {
	Brew          BrewAudit          `json:"brew,omitempty"`
	BrewCask      BrewCaskAudit      `json:"brew_cask,omitempty"`
	AppStore      AppStoreAudit      `json:"app_store,omitempty"`
	UserInstalled UserInstalledAudit `json:"user_installed,omitempty"`
}

// AuditApplications gives a breakdown of applications based on their source,
// namely: (user, brew cask, mac app store)
func (b *MacHelper) AuditApplications() (*ApplicationAudit, error) {
	audit := ApplicationAudit{}
	casks, err := b.getCasks()
	if err != nil {
		return nil, err
	}
	audit.BrewCask.Names = casks

	brews, err := b.getBrews()
	if err != nil {
		return nil, err
	}
	audit.Brew.Names = brews

	var wg sync.WaitGroup

	caskInfo := make(map[string][]string)
	for _, cask := range casks {
		wg.Add(1)
		go func(cask string) {
			defer wg.Done()
			info, err := b.getCaskInfo(cask)
			if err != nil {
				log.Println(err)
			}
			// spew.Dump(info)
			caskInfo[cask] = info
		}(cask)
	}
	wg.Wait()

	masFoundApps, err := b.getMacAppStoreApplications()
	if err != nil {
		return nil, err
	}

	allApps, err := b.getApplications()
	if err != nil {
		return nil, err
	}
	for _, app := range allApps {
		if _, ok := defaultApps[app]; !ok {
			//user installed app
			foundSource := false
			//TODO: make sure brew is installed
			for _, caskInfo := range caskInfo {
				appName := getAppNameFromCaskInfo(caskInfo)
				if app == appName {
					audit.BrewCask.Artifacts = append(audit.BrewCask.Artifacts, app)
					foundSource = true
					break
				}
			}
			//TODO: make sure mas is installed
			for _, masApp := range masFoundApps {
				if app == masApp {
					audit.AppStore.Artifacts = append(audit.AppStore.Artifacts, app)
					foundSource = true
					break
				}
			}

			if !foundSource {
				audit.UserInstalled.Artifacts = append(audit.UserInstalled.Artifacts, app)
				log.Printf("couldn't find match for app=%v \n", app)
			}
		}

	}

	return &audit, nil
}

//MacHelper represents the main helper
type MacHelper struct {
	cm CommandManager
}

//CommandManager is used for running commands
type CommandManager interface {
	runCommand(cmd string, args ...string) ([]string, error)
}

// SystemCommands is the impl
type SystemCommands struct{}

func (c *SystemCommands) runCommand(cmd string, args ...string) ([]string, error) {
	bytes, err := exec.Command(cmd, args...).Output()
	if err != nil {
		return nil, err
	}
	return bytesWithNewLinesToStrings(bytes), err
}

func (b *MacHelper) getApplications() ([]string, error) {
	return b.cm.runCommand("ls", "/Applications")
}
func (b *MacHelper) getMacAppStoreApplications() ([]string, error) {
	apps, err := b.cm.runCommand("mas", "list")
	if err != nil {
		return nil, err
	}
	re := regexp.MustCompile("([0-9]{3,20}) ([^(]+)")

	var executableNames []string
	for _, app := range apps {
		//matches looks like:
		// ([][]string) (len=1 cap=10) {
		// 	([]string) (len=3 cap=3) {
		// 	 (string) (len=22) "1171820258 Highland 2 ",
		// 	 (string) (len=10) "1171820258",
		// 	 (string) (len=11) "Highland 2 "
		// 	}
		//    }
		matches := re.FindAllStringSubmatch(app, -1)
		if len(matches) != 1 || len(matches[0]) != 3 {
			return nil, fmt.Errorf("regex parse error, src='%v', got='%v'", app, matches)
		}
		appName := strings.TrimRight(matches[0][2], " ")
		executableNames = append(executableNames, fmt.Sprintf("%s.app", appName))
	}
	return executableNames, nil
}
func (b *MacHelper) getCasks() ([]string, error) {
	return b.cm.runCommand("brew", "cask", "list")
}
func (b *MacHelper) getBrews() ([]string, error) {
	return b.cm.runCommand("brew", "list")
}
func (b *MacHelper) getCaskInfo(name string) ([]string, error) {
	return b.cm.runCommand("brew", "cask", "info", name)
}

func bytesWithNewLinesToStrings(bytes []byte) []string {
	return strings.Split(strings.TrimSuffix(string(bytes[:]), "\n"), "\n")
}

func getAppNameFromCaskInfo(info []string) string {

	const AppArtifactSuffix = " (App)"
	for _, line := range info {
		if strings.Contains(line, AppArtifactSuffix) {
			return strings.Split(line, AppArtifactSuffix)[0]
		}
	}
	return ""
}
