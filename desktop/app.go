package desktop

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/getlantern/systray"
	"github.com/ncruces/zenity"
	"github.com/zhquiz/go-zhquiz/server/api"
	"github.com/zhquiz/go-zhquiz/shared"
	"github.com/zserge/lorca"
)

var url string

type systrayList struct {
	openButton    *systray.MenuItem
	browserButton *systray.MenuItem
	closeButton   *systray.MenuItem
}

func (s systrayList) init() {
	for {
		select {
		case <-s.openButton.ClickedCh:
			OpenURLInChromeApp()
		case <-s.browserButton.ClickedCh:
			OpenURLInDefaultBrowser(url)
		case <-s.closeButton.ClickedCh:
			systray.Quit()
		}
	}
}

var tray systrayList

// Start starts the app in Chrome App, if possible
func Start(res *api.Resource) {
	var ui *lorca.UI

	systray.Run(func() {
		favicon, err := ioutil.ReadFile(filepath.Join(shared.ExecDir, "public", "favicon.ico"))
		if err != nil {
			log.Fatalln(err)
		}

		systray.SetIcon(favicon)
		systray.SetTitle("ZhQuiz")

		url = fmt.Sprintf("http://localhost:%s", shared.Port())

		tray = systrayList{
			openButton:    systray.AddMenuItem("Open ZhQuiz", "Open ZhQuiz in Chrome App"),
			browserButton: systray.AddMenuItem("Open ZhQuiz in web browser", "Open ZhQuiz in web browser"),
			closeButton:   systray.AddMenuItem("Quit", "Quit ZhQuiz"),
		}

		go tray.init()

		attempts := 0

		terminateAppRunning := false
		terminateApp := func() {
			terminateAppRunning = true

			yes, e := zenity.Question(
				"ZhQuiz server is taking too long to start.\nDo you want to terminate the app?",
				zenity.Title("ZhQuiz server failed to start"),
				zenity.Icon(zenity.QuestionIcon),
				zenity.NoWrap(),
			)

			if e != nil {
				panic(e)
			}

			if yes {
				systray.Quit()
				return
			}

			terminateAppRunning = true
			attempts = 0
		}

		for {
			time.Sleep(1 * time.Second)
			_, err := http.Head(url)
			if err == nil {
				break
			}
			attempts++

			if !terminateAppRunning && attempts >= 5 {
				go terminateApp()
			}
		}

		systray.SetTooltip(fmt.Sprintf("Server running at %s", url))

		OpenURLInChromeApp()
	}, func() {
		if ui != nil {
			(*ui).Close()
		}
	})
}
