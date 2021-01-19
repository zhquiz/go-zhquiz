//+build !darwin

package desktop

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/getlantern/systray"
	"github.com/zhquiz/go-zhquiz/server/api"
	"github.com/zhquiz/go-zhquiz/shared"
)

// Start starts the app in Chrome App, if possible
func Start(res *api.Resource) {
	url := fmt.Sprintf("http://localhost:%s", shared.Port())

	attempts := 0

	terminateAppRunning := false
	terminateApp := func() {
		terminateAppRunning = true

		yes := MessageBox(
			"Server failed to start",
			"The server is taking too long to start. Do you want to terminate the app?",
		)

		if yes {
			res.Cleanup()
			os.Exit(0)
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

	ui := OpenURLInChromeApp(url+"/etabs.html", url)

	if ui != nil {
		defer (*ui).Close()

		sigc := make(chan os.Signal)
		signal.Notify(sigc, os.Interrupt)
		select {
		case <-sigc:
		case <-(*ui).Done():
		}
	} else {
		systray.Run(func() {
			favicon, err := ioutil.ReadFile(filepath.Join(shared.ExecDir, "public", "favicon.ico"))
			if err != nil {
				log.Fatalln(err)
			}

			systray.SetIcon(favicon)
			systray.SetTitle("ZhQuiz")

			url := fmt.Sprintf("http://localhost:%s", shared.Port())
			systray.SetTooltip(fmt.Sprintf("Server running at %s", url))

			openDefaultBtn := systray.AddMenuItem("Open ZhQuiz in web browser", "Open ZhQuiz in web browser")
			closeBtn := systray.AddMenuItem("Quit", "Quit ZhQuiz")

			go func() {
				for {
					select {
					case <-openDefaultBtn.ClickedCh:
						OpenURLInDefaultBrowser(url)
					case <-closeBtn.ClickedCh:
						systray.Quit()
					}
				}
			}()
		}, func() {})
	}

	res.Cleanup()
}
