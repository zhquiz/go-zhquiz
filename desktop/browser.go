package desktop

import (
	"log"
	"regexp"

	"github.com/ncruces/zenity"
	"github.com/zhquiz/go-zhquiz/shared"
	"github.com/zserge/lorca"
)

var ui *lorca.UI

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
				shared.OpenURL("https://www.google.com/chrome/")
			}
		}()

		zenity.Notify("ZhQuiz server is running in systray. Click the systray to activate or shutdown.")

		shared.OpenURL(url)

		c <- false
		return c
	}

	u, err := lorca.New(url+"/etabs.html", "", 1024, 768, "--start-maximized")
	if err != nil {
		log.Fatal(err)
	}

	ui = &u

	u.SetBounds(lorca.Bounds{
		WindowState: lorca.WindowStateMaximized,
	})

	u.Bind("openExternal", func(url string) {
		if regexp.MustCompile("^https?://").MatchString(url) {
			shared.OpenURL(url)
		}
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
