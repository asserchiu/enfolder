package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const (
	defaultConfigName = "enfolder_rule.json"
)

type EnfolderRule struct {
	FolderName string   `json:"folder_name"`
	Keywords   []string `json:"keywords"`
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
