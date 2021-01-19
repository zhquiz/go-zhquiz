package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"time"

	"github.com/getlantern/systray"
	"github.com/zhquiz/go-zhquiz/desktop"
	"github.com/zhquiz/go-zhquiz/server"
	"github.com/zhquiz/go-zhquiz/server/api"
	"github.com/zhquiz/go-zhquiz/shared"
)

func main() {
	shared.Load()

	res := api.Prepare()
	defer res.Cleanup()

	if !shared.IsDebug() {
		server.Serve(&res)

		systray.Run(func() {
			favicon, err := ioutil.ReadFile(filepath.Join(shared.ExecDir, "public", "favicon.ico"))
			if err != nil {
				log.Fatalln(err)
			}

			systray.SetIcon(favicon)
			systray.SetTitle("ZhQuiz")

			url := fmt.Sprintf("http://localhost:%s", shared.Port())

			openChromeBtn := systray.AddMenuItem("Open ZhQuiz", "Open ZhQuiz in Chrome App")
			openDefaultBtn := systray.AddMenuItem("Open ZhQuiz in web browser", "Open ZhQuiz in web browser")
			closeBtn := systray.AddMenuItem("Quit", "Quit ZhQuiz")

			go func() {
				for {
					select {
					case <-openChromeBtn.ClickedCh:
						desktop.OpenURLInChromeApp(url+"/etabs.html", url)
					case <-openDefaultBtn.ClickedCh:
						desktop.OpenURLInDefaultBrowser(url)
					case <-closeBtn.ClickedCh:
						systray.Quit()
					}
				}
			}()

			attempts := 0

			terminateAppRunning := false
			terminateApp := func() {
				terminateAppRunning = true

				yes := desktop.MessageBox(
					"Server failed to start",
					"The server is taking too long to start. Do you want to terminate the app?",
				)

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

			if <-desktop.OpenURLInChromeApp(url+"/etabs.html", url) && runtime.GOOS != "darwin" {
				systray.Quit()
			}
		}, func() {
			res.Cleanup()
		})
	} else {
		server.Serve(&res)

		c := make(chan os.Signal)
		signal.Notify(c, os.Interrupt, os.Kill)

		<-c

		res.Cleanup()
	}
}
