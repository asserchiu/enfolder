package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	defaultConfigName = "enfolder_rule.json"
)

type EnfolderRule struct {
	FolderName string   `json:"folder_name"`
	Keywords   []string `json:"keywords"`
}

// ValidationError represents a folder name validation error
type ValidationError struct {
	FolderName string
	Reason     string
}

// Error implements the error interface for ValidationError
func (e *ValidationError) Error() string {
	return fmt.Sprintf("invalid folder name '%s': %s", e.FolderName, e.Reason)
}

// ValidateFolderName checks if a folder name is valid across Windows, macOS, and Linux
// Returns ValidationError if the name is invalid, nil otherwise
func ValidateFolderName(name string) error {
	// Check if empty
	if name == "" {
		return &ValidationError{
			FolderName: name,
			Reason:     "folder name cannot be empty",
		}
	}

	// Check for Unix-specific restrictions first
	// Cannot be "." or ".."
	if name == "." || name == ".." {
		return &ValidationError{
			FolderName: name,
			Reason:     "folder name cannot be '.' or '..'",
		}
	}

	// Check length (max 255 characters for most filesystems)
	if len(name) > 255 {
		return &ValidationError{
			FolderName: name,
			Reason:     "folder name exceeds 255 characters",
		}
	}

	// Check for invalid characters (Windows is most restrictive)
	// Invalid: < > : " / \ | ? * and control characters (0x00-0x1F)
	invalidChars := regexp.MustCompile(`[<>:"/\\|?*\x00-\x1F]`)
	if invalidChars.MatchString(name) {
		return &ValidationError{
			FolderName: name,
			Reason:     `folder name contains invalid characters (< > : " / \ | ? * or control characters)`,
		}
	}

	// Check if name ends with space or period (Windows restriction)
	if strings.HasSuffix(name, " ") || strings.HasSuffix(name, ".") {
		return &ValidationError{
			FolderName: name,
			Reason:     "folder name cannot end with space or period",
		}
	}

	// Check for reserved names on Windows (case-insensitive)
	// CON, PRN, AUX, NUL, COM1-COM9, LPT1-LPT9
	reservedNames := map[string]bool{
		"CON": true, "PRN": true, "AUX": true, "NUL": true,
		"COM1": true, "COM2": true, "COM3": true, "COM4": true, "COM5": true,
		"COM6": true, "COM7": true, "COM8": true, "COM9": true,
		"LPT1": true, "LPT2": true, "LPT3": true, "LPT4": true, "LPT5": true,
		"LPT6": true, "LPT7": true, "LPT8": true, "LPT9": true,
	}

	// Check base name without extension for reserved names
	upperName := strings.ToUpper(name)
	if reservedNames[upperName] {
		return &ValidationError{
			FolderName: name,
			Reason:     fmt.Sprintf("folder name '%s' is a reserved name on Windows", name),
		}
	}

	// Also check if the name before the first dot is reserved (e.g., "CON.txt")
	if dotIndex := strings.Index(name, "."); dotIndex > 0 {
		baseName := strings.ToUpper(name[:dotIndex])
		if reservedNames[baseName] {
			return &ValidationError{
				FolderName: name,
				Reason:     fmt.Sprintf("folder name '%s' starts with a reserved name on Windows", name),
			}
		}
	}

	return nil
}

func main() {
	var (
		err error

		// TODO: auto detect the execution running from command line interface
		cliMode = flag.Bool("cli", false, "Command Line Interface mode: do not pause execution at the end of execution")
		cfgFile = flag.String("config", defaultConfigName, "config file")
	)

	// parse command-line flags
	flag.Parse()

	// set log output to stdout, instead of stderr
	log.SetOutput(os.Stdout)

	log.Printf("===== enfolder begin =====")
	defer func() {
		if err := recover(); err != nil {
			// No need to print second time, just recover from the panic
			//log.Printf("recover: %v", err)
		}

		log.Printf("===== enfolder end =====")

		if !*cliMode {
			log.Printf("Press ENTER to exit ... (use -cli to skip)")
			bufio.NewReader(os.Stdin).ReadByte()
		}
	}()

	// read rule file
	content, err := ioutil.ReadFile(*cfgFile)
	if err != nil {
		log.Panicf("ioutil.ReadFile err: %v", err)
	}

	// parse rule file
	var rules []EnfolderRule
	err = json.Unmarshal(content, &rules)
	if err != nil {
		log.Panicf("json.Unmarshal err: %v", err)
	}

	// validate all folder names in the config
	var validationErrors []*ValidationError
	for _, rule := range rules {
		if err := ValidateFolderName(rule.FolderName); err != nil {
			// Try to assert the error as ValidationError
			if ve, ok := err.(*ValidationError); ok {
				validationErrors = append(validationErrors, ve)
			} else {
				// Fallback for unexpected error types
				log.Panicf("Unexpected error during validation: %v", err)
			}
		}
	}

	if len(validationErrors) > 0 {
		log.Printf("ERROR: Found %d invalid folder name(s) in the config file:", len(validationErrors))
		log.Printf("-------------------------------------------------------")
		for i, ve := range validationErrors {
			log.Printf("%d. Folder name: '%s'", i+1, ve.FolderName)
			log.Printf("   Reason: %s", ve.Reason)
			log.Printf("")
		}
		log.Printf("-------------------------------------------------------")
		log.Printf("Please fix the invalid folder names in '%s' and run again.", *cfgFile)
		log.Panicf("Config validation failed")
	}
	log.Printf("Config validation passed: all %d folder name(s) are valid", len(rules))

	// get all filename in the working directory
	fileNames, err := filepath.Glob("*")
	if err != nil {
		log.Panicf("filepath.Glob err: %v", err)
	}

	// for all files
	for _, fileName := range fileNames {
		// get destination folder name
		destination := GetDestinationFolderName(fileName, rules)
		if destination == "" {
			// no destination, do nothing and go to next file/folder
			continue
		}

		// create folder before moving file
		err = os.Mkdir(destination, os.ModePerm)
		if err != nil && !os.IsExist(err) {
			log.Panicf("os.Mkdir err: %v", err)
		}

		// move file into the folder
		log.Printf("* Move `%s` into `%s`", fileName, destination)
		err = os.Rename(fileName, destination+string(os.PathSeparator)+fileName)
		if err != nil {
			log.Panicf("os.Rename err: %v", err)
		}
	}
}

func GetDestinationFolderName(fileName string, rules []EnfolderRule) (destinationFolderName string) {
	if fileName == "" {
		return ""
	}

	// check with each rule (destination folder)
	for _, rule := range rules {
		// ignore the folder (or file) that has the same name with the destination folder name
		if fileName == rule.FolderName {
			return ""
		}

		// check with all keywords of the rule
		for _, keyword := range rule.Keywords {
			// check if the filename contains the keyword
			if keyword != "" && strings.Contains(strings.ToLower(fileName), strings.ToLower(keyword)) {
				return rule.FolderName
			}
		}
	}

	return ""
}
