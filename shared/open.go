package shared

import (
	"os/exec"
	"runtime"
	"strings"
)

// OpenURL opens specified URL in the default web browser
func OpenURL(u string) {
	switch runtime.GOOS {
	case "linux":
		exec.Command("xdg-open", u).Run()
	case "darwin":
		exec.Command("open", u).Run()
	case "windows":
		r := strings.NewReplacer("&", "^&")
		exec.Command("cmd", "/c", "start", r.Replace(u)).Run()
	}
}
