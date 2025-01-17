package main

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"image/color"
	"io"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2/theme"
	fatihColor "github.com/fatih/color"
)

func getSHA1Hash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha1.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func loadIgnoreList(filepath string) ([]string, error) {
	var ignoreList []string

	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, &ignoreList); err != nil {
		return nil, err
	}

	return ignoreList, nil
}

func contains(slice []string, val string) bool {
	for _, item := range slice {
		// fmt.Printf("Comparing %q to %q\n", item, val)
		if item == val {
			return true
		}
	}
	return false
}

func checkForContent(directory string) error {
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		printInfo(fatihColor.FgYellow, "%s directory not found\n", directory)
		return fmt.Errorf("%s directory not found", directory)
	}

	logOutput := func(s string) {
		if !guiEnabled {
			printInfo(fatihColor.FgYellow, s+"\n")
		} else {
			addText(theme.PrimaryColorNamed(theme.ColorYellow), s)
		}
	}

	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check directories that are exactly 8 characters long, potential titleID
		if info.IsDir() && len(info.Name()) == 8 {
			titleID := strings.ToLower(info.Name())
			titleData, ok := titles.Titles[titleID]
			if ok {
				// Process known titles as before
				if guiEnabled {
					addHeader(titleData.TitleName)
				}
				printHeader(titleData.TitleName)
			}

			// Check and potentially process $c subdirectory
			subDirDLC := filepath.Join(path, "$c")
			subInfoDLC, err := os.Stat(subDirDLC)
			if err == nil && subInfoDLC.IsDir() {
				if ok { // Process content if titleID is known
					err = processDLCContent(subDirDLC, titleData, titleID, directory)
					if err != nil {
						return err
					}
				} else {
					logOutput(fmt.Sprintf("DLC content found in unrecognized directory: %s\n", subDirDLC))
				}
			}

			// Check and potentially process $u subdirectory
			subDirUpdates := filepath.Join(path, "$u")
			subInfoUpdates, err := os.Stat(subDirUpdates)
			if err == nil && subInfoUpdates.IsDir() {
				if ok { // Process updates if titleID is known
					err = processUpdates(subDirUpdates, titleData, titleID, directory)
					if err != nil {
						return err
					}
				} else {
					if guiEnabled {
					}
					logOutput(fmt.Sprintf("Updates found in unrecognized directory: %s\n", subDirUpdates))
				}
			}

			if !ok {
				return filepath.SkipDir // Skip further processing in unrecognized directories
			}
		}
		return nil
	})

	return err
}

func processDLCContent(subDirDLC string, titleData TitleData, titleID string, directory string) error {
	subContents, err := os.ReadDir(subDirDLC)
	if err != nil {
		return err
	}

	for _, subContent := range subContents {
		subContentPath := filepath.Join(subDirDLC, subContent.Name())
		if !subContent.IsDir() {
			continue
		}

		subDirContents, err := os.ReadDir(subContentPath)
		if err != nil {
			return err
		}

		hasContentMetaXbx := false
		for _, dlcFiles := range subDirContents {
			if strings.Contains(strings.ToLower(dlcFiles.Name()), "contentmeta.xbx") && !dlcFiles.IsDir() {
				hasContentMetaXbx = true
				break
			}
		}

		if !hasContentMetaXbx {
			continue
		}

		contentID := strings.ToLower(subContent.Name())
		if !contains(titleData.ContentIDs, contentID) {
			if guiEnabled {
				addText(theme.ErrorColor(), "Unknown content found at: %s", subContentPath)
			}
			printInfo(fatihColor.FgRed, "Unknown content found at: %s\n", subContentPath)
			continue
		}

		archivedName := ""
		for _, archived := range titleData.Archived {
			for archivedID, name := range archived {
				if archivedID == contentID {
					archivedName = name
					break
				}
			}
			if archivedName != "" {
				break
			}
		}

		subContentPath = strings.TrimPrefix(subContentPath, directory+"/")
		if archivedName != "" {
			if guiEnabled {
				addText(theme.PrimaryColorNamed(theme.ColorGreen), "Content is known and archived %s", archivedName)
			}
			printInfo(fatihColor.FgGreen, "Content is known and archived %s\n", archivedName)

		} else {
			if guiEnabled {
				addText(theme.ErrorColor(), "%s has unarchived content found at: %s", titleData.TitleName, subContentPath)
			}
			printInfo(fatihColor.FgYellow, "%s has unarchived content found at: %s\n", titleData.TitleName, subContentPath)

		}
	}

	return nil
}

func processUpdates(subDirUpdates string, titleData TitleData, titleID string, directory string) error {
	files, err := os.ReadDir(subDirUpdates)
	if err != nil {
		return err
	}

	knownUpdateFound := false
	for _, f := range files {
		if filepath.Ext(f.Name()) != ".xbe" {
			continue
		}

		filePath := filepath.Join(subDirUpdates, f.Name())
		fileHash, err := getSHA1Hash(filePath)
		if err != nil {
			if guiEnabled {
				addText(theme.ErrorColor(), "Error calculating hash for file: %s, error: %s", f.Name(), err.Error())
			}
			printInfo(fatihColor.FgRed, "Error calculating hash for file: %s, error: %s\n", f.Name(), err.Error())

			continue
		}

		for _, knownUpdate := range titleData.TitleUpdatesKnown {
			for knownHash, name := range knownUpdate {
				if knownHash == fileHash {
					if guiEnabled {
						addHeader("File Info")
						addText(theme.PrimaryColorNamed(theme.ColorGreen), "Known and Archived Title update found for %s (%s) (%s)", titleData.TitleName, titleID, name)
						filePath = strings.TrimPrefix(filePath, directory+"/")
						addText(theme.PrimaryColorNamed(theme.ColorGreen), "Path: %s", filePath)
						addText(theme.PrimaryColorNamed(theme.ColorGreen), "SHA1: %s", fileHash)
						addText(color.Transparent, separator)
					}
					printHeader("File Info")
					printInfo(fatihColor.FgGreen, "Known and Archive Title update found for %s (%s) (%s)\n", titleData.TitleName, titleID, name)
					filePath = strings.TrimPrefix(filePath, directory+"/")
					printInfo(fatihColor.FgGreen, "Path: %s\n", filePath)
					printInfo(fatihColor.FgGreen, "SHA1: %s\n", fileHash)
					fmt.Println(separator)

					knownUpdateFound = true
					break
				}
			}
			if knownUpdateFound {
				break
			}
		}

		if !knownUpdateFound {
			if guiEnabled {
				addHeader("File Info")
				addText(theme.ErrorColor(), "Unknown Title Update found for %s (%s)", titleData.TitleName, titleID)
				filePath = strings.TrimPrefix(filePath, directory+"/")
				addText(theme.ErrorColor(), "Path: %s", filePath)
				addText(theme.ErrorColor(), "SHA1: %s", fileHash)
			}
			printHeader("File Info")
			printInfo(fatihColor.FgRed, "Unknown Title Update found for %s (%s)\n", titleData.TitleName, titleID)
			filePath = strings.TrimPrefix(filePath, directory+"/")
			printInfo(fatihColor.FgRed, "Path: %s\n", filePath)
			printInfo(fatihColor.FgRed, "SHA1: %s\n", fileHash)

		}
	}

	return nil
}
