package desktop

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/gen2brain/dlgs"
	"github.com/getlantern/systray"
	"github.com/zhquiz/go-zhquiz/shared"
)

// CreateSystray creates systray holder for browser
func CreateSystray(url string) chan bool {
	c := make(chan bool)

	go func() {
		systray.Run(func() {
			favicon, err := ioutil.ReadFile(filepath.Join(shared.ExecDir, "public", "favicon.ico"))
			if err != nil {
				log.Fatalln(err)
			}

			systray.SetIcon(favicon)
			systray.SetTitle("ZhQuiz")

			openBtn := systray.AddMenuItem("Open ZhQuiz", "Open ZhQuiz in web browser")
			go awaitOpenBtn(openBtn, url)

			closeBtn := systray.AddMenuItem("Quit", "Quit ZhQuiz")
			go func() {
				<-closeBtn.ClickedCh
				systray.Quit()
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
					panic(err)
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
		}, func() {
			c <- true
		})
	}()

	return c
}

func awaitOpenBtn(openBtn *systray.MenuItem, url string) {
	<-openBtn.ClickedCh
	openBrowser(url)
	awaitOpenBtn(openBtn, url)
}

func openBrowser(url string) {
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
