package desktop

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/gen2brain/dlgs"
	"github.com/zhquiz/go-server/shared"
)

// OpenURL opens url in web browser windowed mode
func OpenURL(url string) chan bool {
	browser := shared.Browser()

	if browser == "" {
		browser = LocateChrome()
	}

	if browser == "" {
		PromptDownload()
		log.Fatal(fmt.Errorf("cannot open outside a web browser"))
	}

	c := make(chan bool)

	go func() {
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
				c <- false
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

		cmd := exec.Command(browser, url, "--start-maximized")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		cmd.Run()

		c <- true
	}()

	return c
}
