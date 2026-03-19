package installer

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

type Runtime string

const (
	Docker Runtime = "docker"
	Podman Runtime = "podman"
)

type InstallResult struct {
	Runtime Runtime
	Success bool
	Output  string
	Err     error
}

func Install(rt Runtime, lines chan<- string) error {
	goos := runtime.GOOS

	switch rt {
	case Docker:
		return installDocker(goos, lines)
	case Podman:
		return installPodman(goos, lines)
	default:
		return fmt.Errorf("unknown runtime: %s", rt)
	}
}

func installDocker(goos string, lines chan<- string) error {
	switch goos {
	case "linux":
		send(lines, "Downloading Docker install script from get.docker.com...")
		cmd := exec.Command("sh", "-c", "curl -fsSL https://get.docker.com | sh")
		return runStreamed(cmd, lines)
	case "darwin":
		send(lines, "macOS detected — Docker Desktop must be installed manually.")
		send(lines, "Download: https://www.docker.com/products/docker-desktop/")
		return fmt.Errorf("manual install required on macOS")
	default:
		return fmt.Errorf("unsupported OS: %s", goos)
	}
}

func installPodman(goos string, lines chan<- string) error {
	switch goos {
	case "linux":
		distro := detectDistro()
		send(lines, fmt.Sprintf("Detected Linux distro: %s", distro))

		var cmd *exec.Cmd
		switch distro {
		case "debian", "ubuntu":
			send(lines, "Installing Podman via apt...")
			cmd = exec.Command(
				"sh", "-c",
				"apt-get update -qq && apt-get install -y podman",
			)
		case "fedora":
			send(lines, "Installing Podman via dnf...")
			cmd = exec.Command("sh", "-c", "dnf install -y podman")
		case "rhel", "centos", "rocky", "almalinux":
			send(lines, "Installing Podman via dnf...")
			cmd = exec.Command("sh", "-c", "dnf install -y podman")
		case "arch":
			send(lines, "Installing Podman via pacman...")
			cmd = exec.Command("sh", "-c", "pacman -Sy --noconfirm podman")
		case "opensuse":
			send(lines, "Installing Podman via zypper...")
			cmd = exec.Command("sh", "-c", "zypper install -y podman")
		default:
			// Fallback: official podman install script
			send(lines, "Using official Podman install script...")
			cmd = exec.Command(
				"sh", "-c",
				"curl -fsSL https://podman.io/install.sh | sh",
			)
		}
		return runStreamed(cmd, lines)
	case "darwin":
		send(lines, "Installing Podman via Homebrew...")
		cmd := exec.Command("sh", "-c", "brew install podman && podman machine init && podman machine start")
		return runStreamed(cmd, lines)
	default:
		return fmt.Errorf("unsupported OS: %s", goos)
	}
}

func PostInstallSetup(rt Runtime, lines chan<- string) error {
	if rt != Podman {
		return nil
	}
	send(lines, "Enabling rootless Podman socket...")
	cmds := []string{
		"systemctl --user enable --now podman.socket",
		"loginctl enable-linger $USER",
	}
	for _, c := range cmds {
		if err := runStreamed(exec.Command("sh", "-c", c), lines); err != nil {
			send(lines, fmt.Sprintf("Warning: %v (non-fatal)", err))
		}
	}
	return nil
}

func detectDistro() string {
	if runtime.GOOS != "linux" {
		return runtime.GOOS
	}
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
	case strings.Contains(content, "rhel") || strings.Contains(content, "red hat"):
		return "rhel"
	case strings.Contains(content, "centos"):
		return "centos"
	case strings.Contains(content, "rocky"):
		return "rocky"
	case strings.Contains(content, "almalinux"):
		return "almalinux"
	case strings.Contains(content, "arch"):
		return "arch"
	case strings.Contains(content, "opensuse"):
		return "opensuse"
	default:
		return "unknown"
	}
}

func runStreamed(cmd *exec.Cmd, lines chan<- string) error {
	cmd.Stdout = &chanWriter{ch: lines}
	cmd.Stderr = &chanWriter{ch: lines}
	return cmd.Run()
}

func send(ch chan<- string, msg string) {
	if ch != nil {
		ch <- msg
	}
}

type chanWriter struct {
	ch  chan<- string
	buf strings.Builder
}

func (w *chanWriter) Write(p []byte) (int, error) {
	w.buf.Write(p)
	for {
		s := w.buf.String()
		idx := strings.IndexByte(s, '\n')
		if idx < 0 {
			break
		}
		line := strings.TrimRight(s[:idx], "\r")
		if line != "" {
			w.ch <- line
		}
		w.buf.Reset()
		w.buf.WriteString(s[idx+1:])
	}
	return len(p), nil
}
