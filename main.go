package main

import (
	"fmt"
	"github.com/pterm/pterm"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	// "github.com/pterm/pterm/putils"
)

type Wtf struct {
	account   string
	server    string
	character string
}

type CopyTarget struct {
	wtf     Wtf
	version string
}

var _probableWowInstallLocations = map[string]string{
	"darwin":  "/Applications/World of Warcraft",
	"windows": "C:\\Program Files (x86)\\World of Warcraft",
}

func main() {
	var wow WowInstall

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	pterm.DefaultHeader.WithFullWidth().Println("wow-profile-copy 0.3.0")

	//
	// prompts
	//

	_probableWowInstallLocations["linux"] = fmt.Sprintf("%s/.var/app/com.usebottles.bottles/data/bottles/bottles/WoW/drive_c/Program Files (x86)/World of Warcraft", userHomeDir)

	// this will crash when not on linux, macOS, or windows
	// if you're trying to run wow on BSD or plan9, you can probably fix this yourself
	installLocation := _probableWowInstallLocations[runtime.GOOS]
	base := "/"

	dirOk := isWowInstallDirectory(installLocation)
	if !dirOk {
		if runtime.GOOS == "windows" {
			baseInput, _ := pterm.DefaultInteractiveTextInput.
				WithDefaultText("Which drive is WoW located on? e.g. C, D").
				Show()
			base = fmt.Sprintf("%s:\\", string(baseInput[0]))
		}
		installLocation, _ = promptForWowDirectory(base)
	}

	pterm.Success.Printfln("Found WoW install. Location: %s", installLocation)

	dirConfirm, _ := pterm.DefaultInteractiveConfirm.
		WithDefaultText("Is this directory correct?").
		WithDefaultValue(true).
		Show()
	if !dirConfirm {
		installLocation, _ = promptForWowDirectory(base)
	}

	wow.installDirectory = installLocation
	wow.findAvailableVersions(installLocation)

	pterm.DefaultHeader.Printfln("WoW Install Directory: %s", wow.installDirectory)

	pterm.Info.Println("First, pick the Version, Account, Server, and Character to copy configuration data from.")
	srcConfig := wow.selectWtf(true)
	pterm.Info.Println("Next, pick the Version, Account, Server, and Character to apply that configuration data to.")
	dstConfig := wow.selectWtf(false)

	pterm.Info.Printfln(
		"Source: { Version: %s, Account: %s, Server: %s, Character: %s }",
		srcConfig.version,
		srcConfig.wtf.account,
		srcConfig.wtf.server,
		srcConfig.wtf.character)
	pterm.Info.Printfln(
		"Destination: { Version: %s, Account :%s, Server: %s, Character: %s }",
		dstConfig.version,
		dstConfig.wtf.account,
		dstConfig.wtf.server,
		dstConfig.wtf.character)

	pterm.DefaultHeader.
		WithFullWidth().
		WithBackgroundStyle(pterm.NewStyle(pterm.BgRed)).
		WithTextStyle(pterm.NewStyle(pterm.FgBlack)).
		Println("Take a backup of the relevant WTF folder(s) - This operation can cause data loss!")

	confirmation, _ := pterm.DefaultInteractiveConfirm.
		WithTextStyle(&pterm.ThemeDefault.ErrorMessageStyle).
		WithDefaultText(fmt.Sprintf("Overwrite %s-%s's Keybindings, Macros, and SavedVariables?", dstConfig.wtf.character, dstConfig.wtf.server)).
		Show()
	if !confirmation {
		os.Exit(1)
	}

	//
	// copying process
	//

	hasProblems := false
	srcWtfAccountPath := filepath.Join(wow.installDirectory, srcConfig.version, "WTF", "Account", srcConfig.wtf.account)
	dstWtfAccountPath := filepath.Join(wow.installDirectory, dstConfig.version, "WTF", "Account", dstConfig.wtf.account)
	svFileRegex := regexp.MustCompile(`.*\.lua$`)

	//
	// account-level
	//

	if srcConfig.wtf.account == dstConfig.wtf.account {
		pterm.Warning.Println("Skipping account-level copying - accounts are the same.")
	} else {
		// client configuration
		accountFilesToCopy := [4]string{"bindings-cache.wtf", "config-cache.wtf", "macros-cache.txt", "edit-mode-cache-account.txt"}
	
		for _, file := range accountFilesToCopy {
			src := filepath.Join(srcWtfAccountPath, file)
			dst := filepath.Join(dstWtfAccountPath, file)
			_, err := copyFile(src, dst)
			if err != nil {
				pterm.Warning.Printfln("Not copying %s: %s", file, err)
				hasProblems = true
			} else {
				pterm.Info.Printfln("Copied %s", src)
			}
		}

		// saved variables
	
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
					pterm.Warning.Printfln("Not copying %s: %s", file, err)
					hasProblems = true
				} else {
					pterm.Info.Printfln("Copied %s", src)
				}
			}
		}	
	}

	//
	// character-level client configuration
	//

	srcWtfCharacterPath := filepath.Join(srcWtfAccountPath, srcConfig.wtf.server, srcConfig.wtf.character)
	dstWtfCharacterPath := filepath.Join(dstWtfAccountPath, dstConfig.wtf.server, dstConfig.wtf.character)

	characterFilesToCopy := [5]string{"AddOns.txt", "config-cache.wtf", "layout-local.txt", "macros-cache.txt", "edit-mode-cache-character.txt"}

	for _, file := range characterFilesToCopy {
		src := filepath.Join(srcWtfCharacterPath, file)
		dst := filepath.Join(dstWtfCharacterPath, file)
		_, err := copyFile(src, dst)
		if err != nil {
			pterm.Warning.Printfln("Not copying %s: %s", file, err)
			hasProblems = true
		} else {
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
				pterm.Warning.Printfln("Not copying %s: %s", file, err)
				hasProblems = true
			} else {
				pterm.Info.Printfln("Copied %s", src)
			}
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
	if !hasProblems {
		pterm.Success.Println("Profile copying completed without issues!")
	} else {
		pterm.Warning.Println("Profile copying completed with warnings. See above.")
	}

	if runtime.GOOS == "windows" {
		fmt.Println("Press Enter to continue...")
		fmt.Scanln()
	}
}
