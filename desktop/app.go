//+build !darwin

package desktop

import (
	"fmt"
	"net/http"
	"os"
	"time"

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

	<-OpenURLInChromeApp(url+"/etabs.html", url)

	res.Cleanup()
}
