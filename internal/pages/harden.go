package pages

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

const (
	pkgFail2ban = "fail2ban"
	pkgCrowdSec = "crowdsec"
)

// stepInstallCrowdSec installs CrowdSec, adds standard collections,
// registers the Caddy bouncer and writes the API key to the env file.
func stepInstallCrowdSec(_ SetupConfig, envFile string) error {
	if runtime.GOOS != OSLinux {
		return fmt.Errorf("CrowdSec install is only supported on Linux")
	}

	ctx := context.Background()

	installScript := "curl -s https://install.crowdsec.net | sh"
	if err := runCmd(ctx, "sh", "-c", installScript); err != nil {
		return fmt.Errorf("%s install: %w", pkgCrowdSec, err)
	}

	_ = runCmd(ctx, "systemctl", "enable", "--now", pkgCrowdSec)

	collections := []string{
		"crowdsecurity/linux",
		"crowdsecurity/caddy",
		"crowdsecurity/http-cve",
		"crowdsecurity/whitelist-good-actors",
	}
	for _, col := range collections {
		if err := runCmd(ctx, "cscli", "collections", "install", col); err != nil {
			_ = err // non-fatal
		}
	}

	out, err := runCmdOutput(ctx, "cscli", "bouncers", "add", "caddy-bouncer", "--output", "raw")
	if err != nil {
		return fmt.Errorf("crowdsec bouncer registration: %w", err)
	}

	apiKey := strings.TrimSpace(out)
	if apiKey == "" {
		return fmt.Errorf("crowdsec: empty bouncer API key")
	}

	vars := map[string]string{
		"CROWDSEC_API_KEY": apiKey,
		"CROWDSEC_ENABLED": "true",
	}
	return writeEnvVars(envFile, vars)
}

// stepInstallFail2ban installs Fail2ban with default SSH jail.
func stepInstallFail2ban() error {
	if runtime.GOOS != OSLinux {
		return fmt.Errorf("Fail2ban install is only supported on Linux")
	}

	ctx := context.Background()
	distro := detectLinuxDistro()

	var installCmd []string
	switch distro {
	case "debian", "ubuntu":
		_ = runCmd(ctx, "apt-get", "update", "-qq")
		installCmd = []string{"apt-get", "install", "-y", "-qq", pkgFail2ban}
	case "fedora", "rhel", "centos", "rocky", "almalinux":
		installCmd = []string{"dnf", "install", "-y", pkgFail2ban}
	case "arch":
		installCmd = []string{"pacman", "-Sy", "--noconfirm", pkgFail2ban}
	default:
		return fmt.Errorf("unsupported distro for %s: %s", pkgFail2ban, distro)
	}

	if err := runCmd(ctx, installCmd[0], installCmd[1:]...); err != nil {
		return fmt.Errorf("%s install: %w", pkgFail2ban, err)
	}

	jailLocal := `[DEFAULT]
bantime  = 1h
findtime = 10m
maxretry = 5
backend  = systemd

[sshd]
enabled = true
port    = ssh
logpath = %(sshd_log)s
maxretry = 3
`
	if err := os.WriteFile("/etc/fail2ban/jail.local", []byte(jailLocal), 0o644); err != nil {
		return fmt.Errorf("%s jail.local: %w", pkgFail2ban, err)
	}

	if err := runCmd(ctx, "systemctl", "enable", "--now", pkgFail2ban); err != nil {
		return fmt.Errorf("%s start: %w", pkgFail2ban, err)
	}

	return nil
}

// ── helpers ───────────────────────────────────────────────────────────────────

func runCmd(ctx context.Context, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runCmdOutput(ctx context.Context, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	out, err := cmd.Output()
	return string(out), err
}

func detectLinuxDistro() string {
	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return "unknown"
	}
	content := strings.ToLower(string(data))
	switch {
	case strings.Contains(content, "ubuntu"):
		return "ubuntu"
	case strings.Contains(content, "debian"):
		return "debian"
	case strings.Contains(content, "fedora"):
		return "fedora"
	case strings.Contains(content, "rhel"), strings.Contains(content, "red hat"):
		return "rhel"
	case strings.Contains(content, "centos"):
		return "centos"
	case strings.Contains(content, "rocky"):
		return "rocky"
	case strings.Contains(content, "almalinux"):
		return "almalinux"
	case strings.Contains(content, "arch"):
		return "arch"
	default:
		return "unknown"
	}
}
