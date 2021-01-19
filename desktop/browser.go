package desktop

import (
	"log"
	"os/exec"
	"runtime"
	"strings"

	"github.com/zserge/lorca"
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
func OpenURLInChromeApp(url string, fallbackURL string) *lorca.UI {
	browser := lorca.LocateChrome()

	if browser == "" {
		go func() {
			yes := MessageBox(
				"Chrome not found",
				"No Chrome/Chromium installation was found. Would you like to download and install it now?",
			)

			if yes {
				OpenURLInDefaultBrowser("https://www.google.com/chrome/")
			}
		}()
		OpenURLInDefaultBrowser(fallbackURL)
		return nil
	}

	ui, err := lorca.New(url, "", 1024, 768)
	if err != nil {
		log.Fatal(err)
	}
	ui.SetBounds(lorca.Bounds{
		WindowState: lorca.WindowStateMaximized,
	})

	// A simple way to know when UI is ready (uses body.onload event in JS)
	ui.Bind("openExternal", func(url string) {
		OpenURLInDefaultBrowser(url)
	})

	return &ui
}
