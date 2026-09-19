package tui

import (
	"os/exec"
	"strings"
)

func copyText(s string) error {
	if s == "" {
		return nil
	}
	if _, err := exec.LookPath("wl-copy"); err == nil {
		cmd := exec.Command("wl-copy")
		cmd.Stdin = strings.NewReader(s)
		return cmd.Run()
	}
	cmd := exec.Command("xclip", "-selection", "clipboard")
	cmd.Stdin = strings.NewReader(s)
	return cmd.Run()
}
