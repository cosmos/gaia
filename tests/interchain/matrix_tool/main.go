//go:build version_tool
// +build version_tool

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"os"
	"os/exec"
	"path"
	"slices"
	"strconv"
	"strings"

	"github.com/google/go-github/v62/github"
	"github.com/tidwall/gjson"
	"golang.org/x/mod/semver"
)

const branchBuildSuffix = "-ref"

func GetPreviousMajorMinor(ctx context.Context, testVersion string) (previousVersions []string, upgradeName string, err error) {
	org, ok := os.LookupEnv("GITHUB_REPOSITORY_OWNER")
	if !ok {
		org = "cosmos"
	}
	client := github.NewClient(nil)
	releases, _, err := client.Repositories.ListReleases(ctx, org, "gaia", nil)
	if err != nil {
		err = fmt.Errorf("ListReleases failed: %w", err)
		return
	}
	upgradeName = semver.Major(testVersion)
	testMajor, err := strconv.Atoi(upgradeName[1:])
	if err != nil {
		err = fmt.Errorf("failed to parse major version: %w", err)
		return
	}
	semvers := make([]string, 0, len(releases))
	for _, release := range releases {
		semvers = append(semvers, release.GetTagName())
	}
	var previousMinor, previousRc bool
	slices.SortFunc(semvers,
		func(i, j string) int {
			// Sort in descending order
			return semver.Compare(j, i)
		})
	for _, v := range semvers {
		if !semver.IsValid(v) {
			continue
		}
		var major int
		major, err = strconv.Atoi(semver.Major(v)[1:])
		if err != nil {
			err = fmt.Errorf("failed to parse major version: %w", err)
			return
		}
		if major == testMajor && semver.Compare(v, testVersion) < 0 {
			if !previousRc && semver.Prerelease(v) != "" && semver.Prerelease(testVersion) != "" && (semver.MajorMinor(v) == semver.MajorMinor(testVersion) || semver.Prerelease(testVersion) == branchBuildSuffix) {
				previousRc = true
				previousVersions = append(previousVersions, v)
			} else if !previousMinor && semver.Prerelease(v) == "" {
				previousMinor = true
				previousVersions = append(previousVersions, v)
			}
			continue
		} else if major == testMajor-1 {
			previousVersions = append(previousVersions, v)
			return
		}
	}
	err = fmt.Errorf("failed to find previous major version")
	return
}

func GetSemverForBranch() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("os.Getwd failed: %w\n", err)
	}
	cmd := exec.Command("go", "mod", "edit", "-json")
	cmd.Dir = path.Join(cwd, "..", "..")
	mod, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("go mod edit -json failed: %w\n", err)
	}
	module := gjson.GetBytes(mod, "Module.Path").String()
	parts := strings.Split(module, "/")
	major := parts[len(parts)-1]
	return fmt.Sprintf("%s.999.0%s", major, branchBuildSuffix), nil
}

func GetTestList() ([]string, error) {
	retval := []string{}
	var stderr bytes.Buffer
	uniq := map[string]bool{}
	cmd := exec.Command("go", "test", "-list=.", "./...")
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("go test -list failed with %w : %s\n", err, stderr)
	}
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Test") && !uniq[line] {
			retval = append(retval, line)
			uniq[line] = true
		}
	}
	rand.Shuffle(len(retval), func(i, j int) {
		retval[i], retval[j] = retval[j], retval[i]
	})
	return retval, nil
}

func main() {
	ctx := context.Background()
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <version>\n", os.Args[0])
		return
	}
	if _, err := os.Stat("go.mod"); err != nil {
		fmt.Fprintf(os.Stderr, "go.mod not found: %v\nRun me from the root of the gaia repo!\n", err)
		return
	}
	testTag := os.Args[1]
	testVersion := testTag
	if !semver.IsValid(testVersion) {
		var err error
		testVersion, err = GetSemverForBranch()
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			return
		}
	}
	previous, upgradeName, err := GetPreviousMajorMinor(ctx, testVersion)
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		return
	}
	tests, err := GetTestList()
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		return
	}
	matrix := map[string][]string{
		// It needs to be versionOrBranch so it matches the docker image that was pushed
		"test_version":     {testTag},
		"previous_version": previous,
		"test_name":        tests,
		"upgrade_name":     {upgradeName},
	}
	marshaled, err := json.Marshal(matrix)
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		return
	}
	fmt.Println(string(marshaled))
}
