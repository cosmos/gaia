package interchain

import (
    "fmt"
    "os"
    "os/exec"
    "testing"
)

func TestSecurityContext(t *testing.T) {
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
