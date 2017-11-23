package main

import (
	"bufio"
	"encoding/json"
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
	KeyWords   []string `json:"key_words"`
}

func main() {
	var err error

	log.SetOutput(os.Stdout)

	log.Printf("===== enfolder begin =====")
	defer log.Printf("===== enfolder end =====")

	// read rule file
	content, err := ioutil.ReadFile(defaultConfigName)
	if err != nil {
		log.Fatalf("ioutil.ReadFile err: %v", err)
	}

	// parse rule file
	var rules []EnfolderRule
	err = json.Unmarshal(content, &rules)
	if err != nil {
		log.Fatalf("json.Unmarshal err: %v", err)
	}

	// get all filename in the working directory
	files, err := filepath.Glob("*")
	if err != nil {
		log.Fatalf("filepath.Glob err: %v", err)
	}

	// for all files
	for _, f := range files {
		// check with each rule (destination folder)
		for _, rule := range rules {

			// ignore the file (folder) that has the same name with the destination folder name
			if f == rule.FolderName {
				continue
			}

			// check with all keywords of the rule
			for _, kw := range rule.KeyWords {

				// check if the filename contains the keyword
				if !strings.Contains(strings.ToLower(f), strings.ToLower(kw)) {
					// didn't matched, check next keyword
					continue
				}

				// create folder before moving file
				err = os.Mkdir(rule.FolderName, os.ModePerm)
				if err != nil && !os.IsExist(err) {
					log.Fatalf("os.Mkdir err: %v", err)
				}

				// move file into the folder
				log.Printf("* `%s` <- `%s`", rule.FolderName, f)
				err = os.Rename(f, rule.FolderName+string(os.PathSeparator)+f)
				if err != nil {
					log.Fatalf("os.Rename err: %v", err)
				}
			}
		}
	}

	log.Printf("Press ENTER to exit ...")

	bio := bufio.NewReader(os.Stdin)
	_, _, err = bio.ReadLine()
}
