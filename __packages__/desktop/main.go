package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zhquiz/zhquiz-desktop/server"
	"github.com/zhquiz/zhquiz-desktop/server/api"
	"github.com/zhquiz/zhquiz-desktop/shared"
	"github.com/zserge/lorca"
)

func main() {
	shared.Load()

	res := api.Prepare()
	defer res.Cleanup()

	server.Serve(&res)

	if !shared.IsDebug() {
		title := "ZhQuiz - Hanzi, Vocab and Sentences quizzing"
		u := fmt.Sprintf("http://localhost:%d", shared.Port())

		if lorca.LocateChrome() != "" {
			ui, _ := lorca.New("data:text/html,"+url.PathEscape(fmt.Sprintf(`
			<html>
				<head><title>%s</title></head>
			</html>
			`, title)), "", 1024, 768)
			defer ui.Close()
			ui.SetBounds(lorca.Bounds{
				WindowState: lorca.WindowStateMaximized,
			})

			for {
				time.Sleep(1 * time.Second)
				_, err := http.Head(u)
				if err == nil {
					break
				}
			}

			ui.Load(u + "/etabs.html")
			<-ui.Done()
		} else {
			fallback(title, u+"/etabs.html")
		}
	} else {
		c := make(chan os.Signal, 2)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)

		<-c
	}
}
