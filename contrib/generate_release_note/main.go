// +build exclude

package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

func main() {
	args := os.Args
	if len(args) != 3 {
		fmt.Println("please add os.Args release version and if this is run as github action(true/false), example: go run main.go v7.0.0 false")
	}

	// default path  is for running on github action
	buildReportPath := "./artifacts/build_report"
	changelogPath := "./CHANGELOG.md"
	if args[2] == "false" {
		buildReportPath = "../../artifacts/build_report"
		changelogPath = "../../CHANGELOG.md"
	}

	buildReport, err := os.ReadFile(buildReportPath)
	if err != nil {
		fmt.Printf("file error: %s\n", err)
	}

	changelog, err := FindChangelog(changelogPath, args[1])
	if err != nil {
		fmt.Printf("cannot find changelog: %s\n", err)
	}

	note := strings.Builder{}
	note.WriteString(changelog)
	note.WriteString("```\n")
	note.Write(buildReport)
	note.WriteString("```\n")

	f, err := os.Create("./releasenote")
	if err != nil {
		fmt.Printf("cannot create a release note: %s\n", err)
	}
	defer f.Close()

	_, err = f.WriteString(note.String())
	if err != nil {
		fmt.Printf("cannot write to releasenote: %s\n", err)
	}
}

func FindChangelog(file string, version string) (string, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return "", errors.New("read changelog file failed.")
	}

	changelogs := string(data)
	i := strings.Index(changelogs, "["+version)
	if i == -1 {
		// -1 means not found
		return "", errors.New(fmt.Sprintf("cannot find version %s\n", version))
	}
	j := strings.Index(changelogs[i:], "##")
	if j == -1 {
		// -1 means not found
		return "", errors.New(fmt.Sprintf("cannot find the end of  %s's changelog \n", version))
	}

	return changelogs[i : i+j], nil
}
