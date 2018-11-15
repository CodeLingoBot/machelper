package main

import (
	"log"
	"os/exec"
	"strings"

	"github.com/davecgh/go-spew/spew"
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
	a := LocalBrew{}
	run(&a)
}

func run(b BrewManager) error {
	casks, err := b.getCasks()
	if err != nil {
		return err
	}
	spew.Dump(casks)

	caskInfo := make(map[string][]string)
	for _, cask := range casks {
		info, err := b.getCaskInfo(cask)
		if err != nil {
			return err
		}
		spew.Dump(info)
		caskInfo[cask] = info
	}
	spew.Dump(caskInfo)

	var userApps []string
	var brewApps []string
	allApps, err := b.getApplications()
	if err != nil {
		return err
	}
	for _, app := range allApps {
		if _, ok := defaultApps[app]; !ok {
			//user installed app
			foundSource := false
			for _, caskInfo := range caskInfo {
				appName := getAppNameFromCaskInfo(caskInfo)
				if app == appName {
					brewApps = append(brewApps, app)
					foundSource = true
					continue
				}
			}

			if !foundSource {
				userApps = append(userApps, app)
			}
		}

		log.Printf("couldn't find match for app=%v \n", app)
	}

	spew.Dump(userApps)
	spew.Dump(brewApps)
	return nil
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

//BrewManager is the interface
type BrewManager interface {
	getCasks() ([]string, error)
	getCaskInfo(name string) ([]string, error)
	getApplications() ([]string, error)
}

//LocalBrew represents a local implementation of brew
type LocalBrew struct{}

func (b *LocalBrew) getApplications() ([]string, error) {
	bytes, err := exec.Command("ls", "/Applications").Output()
	if err != nil {
		return nil, err
	}
	return bytesWithNewLinesToStrings(bytes), err
}
func (b *LocalBrew) getCasks() ([]string, error) {
	// cmdName := "brew"
	// cmdArgs := []string{"cask", "list"}
	bytes, err := exec.Command("brew", "cask", "list").Output()
	if err != nil {
		return nil, err
	}
	return bytesWithNewLinesToStrings(bytes), nil
}
func (b *LocalBrew) getCaskInfo(name string) ([]string, error) {
	bytes, err := exec.Command("brew", "cask", "info", name).Output()
	if err != nil {
		return nil, err
	}
	return bytesWithNewLinesToStrings(bytes), nil
}

func bytesWithNewLinesToStrings(bytes []byte) []string {
	return strings.Split(strings.TrimSuffix(string(bytes[:]), "\n"), "\n")
}
