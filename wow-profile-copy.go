package main

import (
	"fmt"
	"os"
	"io"
	"log"
	"runtime"
	"regexp"
	"path/filepath"
	"github.com/pterm/pterm"
	"strings"
	// "github.com/pterm/pterm/putils" 
)

type WowInstall struct {
	availableVersions []string
	installDirectory string
}

type Wtf struct {
	account string
	server string
	character string
}

type CopyTarget struct {
	wtf Wtf
	version string
}

// smelly?
var _wowInstanceFolderNames = map[string]string{
		"_classic_": "WoTLK Classic",
		"_classic_ptr_": "WoTLK Classic PTR",
		"_classic_beta_": "WoTLK Classic Beta",
		"_retail_": "Retail",
	}

var _probableWowInstallLocations = map[string]string{
		"darwin": "/Applications/World of Warcraft",
		"windows": "C:\\Program Files (x86)\\World of Warcraft",
	}


//
//
// WoWInstall methods
//
//

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
								account: acct.Name(),
								server: server.Name(),
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

// prompts the user to select a WTF tuple to copy to/from
// isSource: whether we are selecting the source of the copy or the destination
func (wow WowInstall) selectWtf(isSource bool) CopyTarget {
	var preposition = "to"
	if isSource {
		preposition = "from"
	} 

	var versions []string

	for _, version := range wow.availableVersions {
		versions = append(versions, version)
	}

	wowVersion, _ := pterm.DefaultInteractiveSelect.
		WithOptions(versions).
		WithDefaultText(fmt.Sprintf("WoW Version to copy %s", preposition)).
		Show()
	pterm.Debug.Printfln("chose %s", wowVersion)

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

	var accountOptions []string
	for _, wtf := range wtfConfigs {
		accountOptions = append(accountOptions, wtf.account)
	}
	accountOptions = deduplicateStringSlice(accountOptions)

	chosenAccount, _ := pterm.DefaultInteractiveSelect.
		WithOptions(accountOptions).
		WithDefaultText(fmt.Sprintf("Account to copy %s", preposition)).
		Show()
	pterm.Debug.Printfln("chose %s", chosenAccount)

	var serverOptions []string
	for _, wtf := range wtfConfigs {
		if wtf.account == chosenAccount {
			serverOptions = append(serverOptions, wtf.server)
		}
	}
	serverOptions = deduplicateStringSlice(serverOptions)

	chosenServer, _ := pterm.DefaultInteractiveSelect.
		WithOptions(serverOptions).
		WithDefaultText(fmt.Sprintf("Server to copy %s", preposition)).
		Show()
	pterm.Debug.Printfln("chose %s", chosenServer)

	var characterOptions []string
	for _, wtf := range wtfConfigs {
		if wtf.account == chosenAccount && wtf.server == chosenServer {
			characterOptions = append(characterOptions, wtf.character)
		}
	}

	chosenCharacter, _ := pterm.DefaultInteractiveSelect.
		WithOptions(characterOptions).
		WithDefaultText(fmt.Sprintf("Character to copy %s", preposition)).
		Show()
	pterm.Debug.Printfln("chose %s", chosenCharacter)

	return CopyTarget{
		wtf: Wtf{
			account: chosenAccount,
			server: chosenServer,
			character: chosenCharacter,
		},
		version: wowVersion,
	}
}

//
//
// helper functions
//
//

// determines if a given path appears to contain a WoW install
func isWowInstallDirectory(dir string) bool {
	var isInstallDir = false

	files, err := os.ReadDir(dir)
	if err != nil {
		// directory probably doesn't exist
		return false
	}

	for _, file := range files {
		// if this directory contains a wow instance folder name, it's probably where WoW is installed
		_, matchesInstanceName := _wowInstanceFolderNames[file.Name()]
		if file.IsDir() && matchesInstanceName {
			isInstallDir = true
			break
		}
	}

	return isInstallDir
}

func promptForWowDirectory(dir string) (wowDir string, err error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}

	var fileChoices = []string{".. (go back)"}

	for _, file := range files {
		fileChoices = append(fileChoices, file.Name())
	}

	selectedFile, _ := pterm.DefaultInteractiveSelect.
		WithOptions(fileChoices).
		WithDefaultText("Select a WoW Install directory").
		WithMaxHeight(15).
		Show()
	var fullSelectedPath string
	if selectedFile == ".. (go back)" {
		fullSelectedPath = filepath.Clean(filepath.Join(dir, ".."))
	} else {
		fullSelectedPath = filepath.Join(dir, selectedFile)
	}
	isWowDir := isWowInstallDirectory(fullSelectedPath)
	if !isWowDir {
		return promptForWowDirectory(fullSelectedPath)
	} else {
		return fullSelectedPath, nil
	}
}

// deduplicates slices by throwing them into a map 
// not mine, credit to @kylewbanks
func deduplicateStringSlice(input []string) []string {
	u := make([]string, 0, len(input))
	m := make(map[string]bool)
	for _, val := range input {
		if _, ok := m[val]; !ok {
			m[val] = true
			u = append(u, val)
		}
	}
	return u
}

// small wrapper around os and io to copy files from source to destination
func copyFile(src string, dest string) (bytes int64, err error) {
	srcFileHandle, err := os.Open(src)
	if err != nil {
		return -1, err
	}
	defer srcFileHandle.Close()

	dstFileHandle, err := os.Create(dest)
	if err != nil {
		return -1, err
	}
	defer dstFileHandle.Close()

	bytes, err = io.Copy(dstFileHandle, srcFileHandle)
	return bytes, err
}

func main() {
	var wow WowInstall

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	_probableWowInstallLocations["linux"] = fmt.Sprintf("%s/.var/app/com.usebottles.bottles/data/bottles/bottles/WoW/drive_c/Program Files (x86)/World of Warcraft", userHomeDir)

	// this will crash when not on linux, macOS, or windows
	// if you're trying to run wow on BSD or plan9, you can probably fix this yourself
	installLocation := _probableWowInstallLocations[runtime.GOOS]

	dirOk := isWowInstallDirectory(installLocation);
	if !dirOk {
		base := "/"
		if runtime.GOOS == "windows" {
			baseInput, _ := pterm.DefaultInteractiveTextInput.
				WithDefaultText("Which drive is WoW located on? e.g. C, D").
				Show()
			base = fmt.Sprintf("%s:\\", string(baseInput[0]))
		}
		installLocation, _ = promptForWowDirectory(base);
	}

	wow.installDirectory = installLocation
	wow.findAvailableVersions(installLocation)

	pterm.DefaultHeader.Printfln("WoW Install Directory: %s", wow.installDirectory)

	pterm.Info.Println("First, pick the Version, Account, Server, and Character to copy configuration data from.")
	srcConfig := wow.selectWtf(true)
	pterm.Info.Println("Next, pick the Version, Account, Server, and Character to apply that configuration data to.")
	dstConfig := wow.selectWtf(false)

	pterm.Info.Printfln("Source: { Version: %s, Account: %s, Server: %s, Character: %s }", _wowInstanceFolderNames[srcConfig.version], srcConfig.wtf.account, srcConfig.wtf.server, srcConfig.wtf.character)
	pterm.Info.Printfln("Destination: { Version: %s, Account :%s, Server: %s, Character: %s }", _wowInstanceFolderNames[dstConfig.version], dstConfig.wtf.account, dstConfig.wtf.server, dstConfig.wtf.character)

	confirmation, _ := pterm.DefaultInteractiveConfirm.
		WithTextStyle(&pterm.ThemeDefault.WarningMessageStyle).
		WithDefaultText(fmt.Sprintf("Overwrite %s-%s's Keybindings, Macros, and SavedVariables?\nThis can cause data loss - make a backup if unsure!", dstConfig.wtf.character, dstConfig.wtf.server)).
		Show()
	if !confirmation {
		os.Exit(1)
	}

	//
	// account-level client configuration
	//

	srcWtfAccountPath := filepath.Join(wow.installDirectory, srcConfig.version, "WTF", "Account", srcConfig.wtf.account)
	dstWtfAccountPath := filepath.Join(wow.installDirectory, dstConfig.version, "WTF", "Account", dstConfig.wtf.account)

	accountFilesToCopy := [3]string{"bindings-cache.wtf", "config-cache.wtf", "macros-cache.txt"}

	for _, file := range accountFilesToCopy {
		src := filepath.Join(srcWtfAccountPath, file)
		dst := filepath.Join(dstWtfAccountPath, file)
		_, err := copyFile(src, dst)
		if err != nil {
			log.Fatal(err)
		}
		pterm.Info.Printfln("Copied %s", src)
	}

	//
	// character-level client configuration
	//

	srcWtfCharacterPath := filepath.Join(srcWtfAccountPath, srcConfig.wtf.server, srcConfig.wtf.character)
	dstWtfCharacterPath := filepath.Join(dstWtfAccountPath, dstConfig.wtf.server, dstConfig.wtf.character)

	characterFilesToCopy := [4]string{"AddOns.txt", "config-cache.wtf", "layout-local.txt", "macros-cache.txt"}

	for _, file := range characterFilesToCopy {
		src := filepath.Join(srcWtfCharacterPath, file)
		dst := filepath.Join(dstWtfCharacterPath, file)
		_, err := copyFile(src, dst)
		if err != nil {
			log.Fatal(err)
		}
		pterm.Info.Printfln("Copied %s", src)
	}

	//
	// account-level saved variables
	//

	svFileRegex := regexp.MustCompile(`.*\.lua$`)

	accountSavedVariablesFiles, err := os.ReadDir(filepath.Join(srcWtfAccountPath, "SavedVariables"))
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range accountSavedVariablesFiles {
		if svFileRegex.MatchString(file.Name()) {
			src := filepath.Join(srcWtfAccountPath, "SavedVariables", file.Name())
			dst := filepath.Join(dstWtfAccountPath, "SavedVariables", file.Name())
			_, err := copyFile(src, dst)
			if err != nil {
				log.Fatal(err)
			}
			pterm.Info.Printfln("Copied %s", src)
		}
	}

	//
	// character-level saved variables
	//

	charSavedVariablesFiles, err := os.ReadDir(filepath.Join(srcWtfCharacterPath, "SavedVariables"))
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range charSavedVariablesFiles {
		if svFileRegex.MatchString(file.Name()) {
			src := filepath.Join(srcWtfCharacterPath, "SavedVariables", file.Name())
			dst := filepath.Join(dstWtfCharacterPath, "SavedVariables", file.Name())
			_, err := copyFile(src, dst)
			if err != nil {
				log.Fatal(err)
			}
			pterm.Info.Printfln("Copied %s", src)
		}
	}

	//
	// clean up
	//
	dstAccountCache := filepath.Join(dstWtfAccountPath, "cache.md5")
	err = os.Remove(dstAccountCache)
	if err != nil {
		if !strings.Contains(err.Error(), "no such file or directory") {
			log.Fatal(err)
		}
	}

	pterm.Info.Printfln("Removed %s", dstAccountCache)

	dstCharacterCache := filepath.Join(dstWtfCharacterPath, "cache.md5")
	err = os.Remove(dstCharacterCache)
	if err != nil {
		if !strings.Contains(err.Error(), "no such file or directory") {
			log.Fatal(err)
		}
	}

	pterm.Info.Printfln("Removed %s", dstCharacterCache)
	pterm.Success.Println("All files copied successfully!")

	if runtime.GOOS == "windows" {
		fmt.Println("Press Enter to continue...")
		fmt.Scanln()
	}
}