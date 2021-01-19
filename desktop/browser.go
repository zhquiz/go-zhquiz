package desktop

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// OpenURLInDefaultBrowser opens specified URL in the default web browser
func OpenURLInDefaultBrowser(url string) {
	switch runtime.GOOS {
	case "linux":
		exec.Command("xdg-open", url).Run()
	case "darwin":
		exec.Command("open", url).Run()
	case "windows":
		r := strings.NewReplacer("&", "^&")
		exec.Command("cmd", "/c", "start", r.Replace(url)).Run()
	}
}

// OpenURLInChromeApp opens url in Chrome or Chromium windowed mode
func OpenURLInChromeApp(url string, fallbackURL string) chan bool {
	browser := LocateChrome()

	c := make(chan bool)

	if browser == "" {
		go PromptDownload()
		OpenURLInDefaultBrowser(fallbackURL)
		c <- false
		return c
	}

	go func() {
		cmd := exec.Command(browser, url, "--start-maximized", "--app="+url)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		cmd.Run()

		c <- true
	}()

	return c
}
