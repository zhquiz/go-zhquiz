package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/joho/godotenv"
	"github.com/zhquiz/go-server/desktop"
	"github.com/zhquiz/go-server/server"
	"github.com/zhquiz/go-server/server/api"
	"github.com/zhquiz/go-server/shared"
)

func main() {
	p := shared.Paths()
	godotenv.Load(p.Dotenv())

	res := api.Prepare()
	defer res.Cleanup()

	if shared.IsDesktop() {
		w := desktop.OpenInWindowedChrome(fmt.Sprintf("http://localhost:%s", shared.Port()))
		defer w.Close()

		server.Serve(&res)

		<-w.Done()
	} else {
		server.Serve(&res)

		c := make(chan os.Signal)
		signal.Notify(c, os.Interrupt)

		<-c
	}
}
