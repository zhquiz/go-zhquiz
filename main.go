package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/webview/webview"
	"github.com/zhquiz/go-zhquiz/server"
	"github.com/zhquiz/go-zhquiz/server/api"
	"github.com/zhquiz/go-zhquiz/shared"
)

func main() {
	shared.Load()

	res := api.Prepare()
	defer res.Cleanup()

	server.Serve(&res)

	if !shared.IsDebug() {
		url := fmt.Sprintf("http://localhost:%d", shared.Port())

		for {
			time.Sleep(1 * time.Second)
			_, err := http.Head(url)
			if err == nil {
				break
			}
		}

		w := webview.New(true)
		defer w.Destroy()

		w.SetSize(1024, 768, webview.HintNone)
		w.SetTitle("ZhQuiz")
		w.Navigate(url + "/etabs.html")
		w.Run()
	} else {
		c := make(chan os.Signal, 2)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)

		<-c
	}
}
