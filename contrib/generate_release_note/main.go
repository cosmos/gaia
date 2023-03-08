//go:build exclude
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
	if len(args) != 4 {
		fmt.Println("please add os.Args release version, build_report path and CHANGELOG.md path, example: go run main.go v7.0.0 ../../artifacts/build_report ../../CHANGELOG.md")
	}

	buildReportPath := args[2]
	changelogPath := args[3]

	buildReport, err := os.ReadFile(buildReportPath)
	if err != nil {
		fmt.Printf("file error: %s\n", err)
	}

	changelog, err := FindChangelog(changelogPath, args[1])
	if err != nil {
		fmt.Printf("cannot find changelog: %s\n", err)
	}

	note := strings.Builder{}
	note.WriteString(fmt.Sprintf("#Gaia %s Release Notes\n", args[1]))
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

func FindChangelog(file, version string) (string, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return "", errors.New("read changelog file failed")
	}

	changelogs := string(data)
	i := strings.Index(changelogs, "["+version)
	if i == -1 {
		// -1 means not found
		return "", fmt.Errorf("cannot find version %s", version)
	}
	j := strings.Index(changelogs[i:], "##")
	if j == -1 {
		// -1 means not found
		return "", fmt.Errorf("cannot find the end of  %s's changelog", version)
	}

	return changelogs[i : i+j], nil
}
