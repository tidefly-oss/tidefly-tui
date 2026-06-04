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

	distroDebian    = "debian"
	distroUbuntu    = "ubuntu"
	distroFedora    = "fedora"
	distroRHEL      = "rhel"
	distroCentOS    = "centos"
	distroRocky     = "rocky"
	distroAlmaLinux = "almalinux"
	distroArch      = "arch"
)

func stepInstallCrowdSec(_ SetupConfig, envFile string) error {
	if runtime.GOOS != OSLinux {
		return fmt.Errorf("CrowdSec install is only supported on Linux")
	}

	ctx := context.Background()
	distro := detectLinuxDistro()

	repoScript := "curl -s https://install.crowdsec.net | sh"
	if err := runCmd(ctx, "sh", "-c", repoScript); err != nil {
		return fmt.Errorf("%s repo setup: %w", pkgCrowdSec, err)
	}

	switch distro {
	case distroDebian, distroUbuntu:
		_ = runCmd(ctx, "apt-get", "update", "-qq")
		if err := runCmd(ctx, "apt-get", "install", "-y", "-qq", pkgCrowdSec); err != nil {
			return fmt.Errorf("%s install: %w", pkgCrowdSec, err)
		}
	case distroFedora, distroRHEL, distroCentOS, distroRocky, distroAlmaLinux:
		if err := runCmd(ctx, "dnf", "install", "-y", pkgCrowdSec); err != nil {
			return fmt.Errorf("%s install: %w", pkgCrowdSec, err)
		}
	default:
		return fmt.Errorf("unsupported distro for %s: %s", pkgCrowdSec, distro)
	}

	if err := runCmd(ctx, "systemctl", "enable", "--now", pkgCrowdSec); err != nil {
		return fmt.Errorf("%s service: %w", pkgCrowdSec, err)
	}

	collections := []string{
		"crowdsecurity/linux",
		"crowdsecurity/caddy",
		"crowdsecurity/http-cve",
		"crowdsecurity/whitelist-good-actors",
	}
	for _, col := range collections {
		_ = runCmd(ctx, "cscli", "collections", "install", col)
	}

	// Delete existing bouncer if present (idempotent reinstall)
	_ = runCmd(ctx, "cscli", "bouncers", "delete", "caddy-bouncer")

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

func stepInstallFail2ban() error {
	if runtime.GOOS != OSLinux {
		return fmt.Errorf("Fail2ban install is only supported on Linux")
	}

	ctx := context.Background()
	distro := detectLinuxDistro()

	var installCmd []string
	switch distro {
	case distroDebian, distroUbuntu:
		_ = runCmd(ctx, "apt-get", "update", "-qq")
		installCmd = []string{"apt-get", "install", "-y", "-qq", pkgFail2ban}
	case distroFedora, distroRHEL, distroCentOS, distroRocky, distroAlmaLinux:
		installCmd = []string{"dnf", "install", "-y", pkgFail2ban}
	case distroArch:
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
	case strings.Contains(content, distroUbuntu):
		return distroUbuntu
	case strings.Contains(content, distroDebian):
		return distroDebian
	case strings.Contains(content, distroFedora):
		return distroFedora
	case strings.Contains(content, distroRHEL), strings.Contains(content, "red hat"):
		return distroRHEL
	case strings.Contains(content, distroCentOS):
		return distroCentOS
	case strings.Contains(content, distroRocky):
		return distroRocky
	case strings.Contains(content, distroAlmaLinux):
		return distroAlmaLinux
	case strings.Contains(content, distroArch):
		return distroArch
	default:
		return "unknown"
	}
}
