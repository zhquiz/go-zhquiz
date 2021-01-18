package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/gen2brain/dlgs"
	"github.com/getlantern/systray"
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

			openBtn := systray.AddMenuItem("Open ZhQuiz", "Open ZhQuiz in web browser")
			closeBtn := systray.AddMenuItem("Quit", "Quit ZhQuiz")

			go func() {
				for {
					select {
					case <-openBtn.ClickedCh:
						url := fmt.Sprintf("http://localhost:%s", shared.Port())

						switch runtime.GOOS {
						case "linux":
							exec.Command("xdg-open", url).Run()
						case "darwin":
							exec.Command("open", url).Run()
						case "windows":
							r := strings.NewReplacer("&", "^&")
							exec.Command("cmd", "/c", "start", r.Replace(url)).Run()
						}
					case <-closeBtn.ClickedCh:
						systray.Quit()
					}
				}
			}()

			attempts := 0

			terminateAppRunning := false
			terminateApp := func() {
				terminateAppRunning = true
				yes, err := dlgs.Question(
					"Server failed to start",
					"The server is taking too long to start. Do you want to terminate the app?",
					false,
				)

				if err != nil {
					log.Fatalln(err)
				}

				if yes {
					systray.Quit()
					return
				}

				terminateAppRunning = true
				attempts = 0
			}

			url := fmt.Sprintf("http://localhost:%s", shared.Port())

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
