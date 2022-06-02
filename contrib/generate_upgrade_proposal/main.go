package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"text/template"
)

type UpgradeJson struct {
	Upgrade_name    string
	Upgrade_version string
	Height          string
	Proposer        string
	Description     string
}

type ProposalInfo struct {
	Upgrade_name      string
	Gaia_version      string
	Height            string
	Proposer          string
	Linux_amd64_sha   string
	Linux_arm64_sha   string
	Darwin_amd64_sha  string
	Windows_amd64_sha string
	Description       string
}

func main() {
	 if len(os.Args) != 2 {
	 	fmt.Println("please indicate the path of release.json file")
	 	os.Exit(1)
	 }

	pi := GetProposalInfo(os.Args[1])
	err := FormProposal(pi, "./proposal")
	if err != nil {
		panic(err)
	}
}

// get proposal info from release.json and build_report
func GetProposalInfo(releaseJson  string) ProposalInfo {
	rdata, err := os.ReadFile(releaseJson)
	if err != nil {
		panic(err)
	}

	upgradeJson := UpgradeJson{}
	err = json.Unmarshal(rdata, &upgradeJson)
	if err != nil {
		panic(err)
	}

	report, err := GetBuildReport(upgradeJson.Upgrade_version)
	if err != nil {
		panic(err)
	}
	shaMap, err := ParseSha(string(report))

	return ProposalInfo{
		Upgrade_name:      upgradeJson.Upgrade_name,
		Gaia_version:      upgradeJson.Upgrade_version,
		Height:            upgradeJson.Height,
		Proposer:          upgradeJson.Proposer,
		Linux_amd64_sha:   shaMap["linux_amd64"],
		Linux_arm64_sha:   shaMap["linux_arm64"],
		Darwin_amd64_sha:  shaMap["darwin_amd64"],
		Windows_amd64_sha: shaMap["windows_amd64"],
		Description:       upgradeJson.Description,
	}
}

// generate proposal from template
func FormProposal(pi ProposalInfo, saveAt string) error {
	t, err := template.New("proposal generation").Parse(`
gaiad tx gov submit-proposal software-upgrade {{ .Upgrade_name}} \
--title  {{ .Upgrade_name}} \
--deposit 50000stake \
--upgrade-height {{ .Height}} \
--upgrade-info '{"binaries":{"linux/amd64":"https://github.com/cosmos/gaia/releases/download/{{ .Gaia_version}}/gaiad-{{ .Gaia_version}}-linux-amd64?checksum=sha256:{{ .Linux_amd64_sha}}","linux/arm64":"https://github.com/cosmos/gaia/releases/download/{{ .Gaia_version}}/gaiad-{{ .Gaia_version}}-linux-arm64?checksum=sha256:{{ .Linux_arm64_sha}}","darwin/amd64":"https://github.com/cosmos/gaia/releases/download/{{ .Gaia_version}}/gaiad-{{ .Gaia_version}}-darwin-amd64?checksum=sha256:{{ .Darwin_amd64_sha}}","windows/amd64":"https://github.com/cosmos/gaia/releases/download/{{ .Gaia_version}}/gaiad-{{ .Gaia_version}}-windows-amd64.exe?checksum=sha256:{{ .Windows_amd64_sha}}"}}' \
--description {{ .Description}} \
--from  {{ .Proposer }} \
--gas auto \
--generate-only
`)

	if err != nil {
		return err
	}

	file, err := os.Create(saveAt)
	if err != nil {
		return err
	}
	err = t.Execute(file, pi)
	if err != nil {
		return err
	}

	return nil
}

func GetBuildReport(version string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://github.com/cosmos/gaia/releases/download/%s/build_report", version))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	return string(b), err
}


// get sha info from build_report file
func ParseSha(content string) (map[string]string, error) {
	shaMap := make(map[string]string)
	file, err := os.Open(content)
	if err != nil {
		return shaMap, err
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	osArcs := []string{"darwin-amd64", "linux-amd64", "linux-arm64", "windows-amd64"}

	for scanner.Scan() {
		text := scanner.Text()
		for _, osArc := range osArcs {
			if strings.Contains(text, osArc) {
				shaMap[osArc] = strings.Split(text, " ")[1]
			}
		}
	}

	return shaMap, nil
}
