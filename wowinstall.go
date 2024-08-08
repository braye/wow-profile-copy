package main

import (
	"fmt"
	"github.com/pterm/pterm"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

type WowInstall struct {
	availableVersions []string
	installDirectory  string
}

// Finds all valid WTF configs (account, server, character) for a given WoW version
func (wow WowInstall) getWtfConfigurations(version string) []Wtf {
	var configurations []Wtf

	wtfPath := filepath.Join(wow.installDirectory, version, "WTF", "Account") // a fitting name

	// enumerate available accounts on this instance
	wtfFiles, err := os.ReadDir(wtfPath)
	if err != nil {
		log.Fatal(err)
	}

	// search all directories in WTF/Account
	for _, acct := range wtfFiles {
		if acct.IsDir() && acct.Name() != "SavedVariables" {
			accountPath := filepath.Join(wtfPath, acct.Name())
			serverFiles, err := os.ReadDir(accountPath) // enumerate available servers under each account
			if err != nil {
				log.Fatal(err)
			}
			for _, server := range serverFiles {
				if server.IsDir() && server.Name() != "SavedVariables" { // assume that any folder that isn't SavedVariables here is a realm
					serverPath := filepath.Join(accountPath, server.Name())
					characterFiles, err := os.ReadDir(serverPath)
					if err != nil {
						log.Fatal(err)
					}
					for _, character := range characterFiles { // any subdirectories of the server directories are characters, they have arbitrary names
						if character.IsDir() {
							finalWtf := Wtf{
								account:   acct.Name(),
								server:    server.Name(),
								character: character.Name(),
							}
							configurations = append(configurations, finalWtf)
						}
					}
				}
			}
		}
	}
	return configurations
}

// determines which WoW versions are available in a given WoW install directory (classic, retail, SoM, etc..)
func (wow *WowInstall) findAvailableVersions(dir string) {
	files, err := os.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		// if this directory contains a wow instance folder name, it's probably where WoW is installed
		_, matchesInstanceName := _wowInstanceFolderNames[file.Name()]
		if file.IsDir() && matchesInstanceName {
			wow.availableVersions = append(wow.availableVersions, file.Name())
		}
	}
}

// prompts the user to select a wow game version, and a WTF tuple to copy to/from
// wtf tuples are (account, server, character)
// isSource: whether we are selecting the source of the copy or the destination
func (wow WowInstall) selectWtf(isSource bool) CopyTarget {
	preposition := "to"
	if isSource {
		preposition = "from"
	}

	optionsHiddenText := "[Some options hidden, use arrow keys to reveal]"
	const selectHeight = 15
	var versions []string

	//
	// prompt for WoW version
	//

	for _, version := range wow.availableVersions {
		versions = append(versions, version)
	}

	defaultText := fmt.Sprintf("WoW Version to copy %s", preposition)
	// give the user an indication that they can scroll the selection window
	if len(versions) > selectHeight {
		defaultText = fmt.Sprintf("WoW Version to copy %s %s", preposition, optionsHiddenText)
	}

	wowVersion, _ := pterm.DefaultInteractiveSelect.
		WithOptions(versions).
		WithDefaultText(defaultText).
		WithMaxHeight(selectHeight).
		Show()
	pterm.Debug.Printfln("chose %s", wowVersion)

	// validate that the chosen wow version actually has configurations to copy from/to
	// wtf configs are only generated when you login to a character
	wtfConfigs := wow.getWtfConfigurations(wowVersion)
	if len(wtfConfigs) == 0 {
		pterm.Error.Printfln("No valid WTF configurations found in %s. Try logging into a character on this version of the client, first!", wowVersion)
		if runtime.GOOS == "windows" {
			// make windows users feel at home
			fmt.Println("Press Enter to continue...")
			fmt.Scanln()
		}
		os.Exit(1)
	}

	//
	// prompt for account
	//

	var accountOptions []string
	for _, wtf := range wtfConfigs {
		accountOptions = append(accountOptions, wtf.account)
	}
	accountOptions = deduplicateStringSlice(accountOptions)

	if len(accountOptions) > selectHeight {
		defaultText = fmt.Sprintf("Account to copy %s %s", preposition, optionsHiddenText)
	} else {
		defaultText = fmt.Sprintf("Account to copy %s", preposition)
	}

	chosenAccount, _ := pterm.DefaultInteractiveSelect.
		WithOptions(accountOptions).
		WithDefaultText(defaultText).
		WithMaxHeight(selectHeight).
		Show()
	pterm.Debug.Printfln("chose %s", chosenAccount)

	//
	// prompt for server
	//

	var serverOptions []string
	for _, wtf := range wtfConfigs {
		if wtf.account == chosenAccount {
			serverOptions = append(serverOptions, wtf.server)
		}
	}
	serverOptions = deduplicateStringSlice(serverOptions)

	if len(serverOptions) > selectHeight {
		defaultText = fmt.Sprintf("Server to copy %s %s", preposition, optionsHiddenText)
	} else {
		defaultText = fmt.Sprintf("Server to copy %s", preposition)
	}

	chosenServer, _ := pterm.DefaultInteractiveSelect.
		WithOptions(serverOptions).
		WithDefaultText(fmt.Sprintf("Server to copy %s", preposition)).
		WithMaxHeight(selectHeight).
		Show()
	pterm.Debug.Printfln("chose %s", chosenServer)

	//
	// prompt for character
	//

	var characterOptions []string
	for _, wtf := range wtfConfigs {
		if wtf.account == chosenAccount && wtf.server == chosenServer {
			characterOptions = append(characterOptions, wtf.character)
		}
	}

	if len(characterOptions) > selectHeight {
		defaultText = fmt.Sprintf("Character to copy %s %s", preposition, optionsHiddenText)
	} else {
		defaultText = fmt.Sprintf("Character to copy %s", preposition)
	}

	chosenCharacter, _ := pterm.DefaultInteractiveSelect.
		WithOptions(characterOptions).
		WithDefaultText(defaultText).
		WithMaxHeight(selectHeight).
		Show()
	pterm.Debug.Printfln("chose %s", chosenCharacter)

	return CopyTarget{
		wtf: Wtf{
			account:   chosenAccount,
			server:    chosenServer,
			character: chosenCharacter,
		},
		version: wowVersion,
	}
}
