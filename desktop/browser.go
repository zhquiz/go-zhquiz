package desktop

import (
	"log"
	"os/exec"
	"runtime"
	"strings"

	"github.com/ncruces/zenity"
	"github.com/zserge/lorca"
)

var ui *lorca.UI

// OpenURLInDefaultBrowser opens specified URL in the default web browser
func OpenURLInDefaultBrowser(u string) {
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

// initWebview opens url in Chrome or Chromium windowed mode
func initWebview() chan bool {
	c := make(chan bool)

	if lorca.LocateChrome() == "" {
		go func() {
			yes, e := zenity.Question(
				"No Chrome/Chromium installation was found.\nWould you like to download and install it now?",
				zenity.Title("Chrome not found"),
				zenity.Icon(zenity.QuestionIcon),
				zenity.NoWrap(),
			)

			if e != nil {
				panic(e)
			}

			if yes {
				OpenURLInDefaultBrowser("https://www.google.com/chrome/")
			}
		}()

		zenity.Notify("ZhQuiz server is running in systray. Click the systray to activate or shutdown.")

		OpenURLInDefaultBrowser(url)

		c <- false
		return c
	}

	u, err := lorca.New(url+"/etabs.html", "", 1024, 768)
	if err != nil {
		log.Fatal(err)
	}

	ui = &u

	u.SetBounds(lorca.Bounds{
		WindowState: lorca.WindowStateMaximized,
	})

	u.Bind("openExternal", func(url string) {
		OpenURLInDefaultBrowser(url)
	})

	go func() {
		defer func() {
			(*ui).Close()
			ui = nil
		}()

		<-(*ui).Done()

		if tray.openButton != nil {
			tray.openButton.Enable()
			zenity.Notify("ZhQuiz server is still running. Click the systray to reactivate or shutdown.")
		}

		c <- true
	}()

	tray.openButton.Disable()

	return c
}
