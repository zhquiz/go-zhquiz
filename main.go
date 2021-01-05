package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/tebeka/atexit"
	"github.com/zhquiz/go-server/desktop"
	"github.com/zhquiz/go-server/server"
	"github.com/zhquiz/go-server/server/api"
	"github.com/zhquiz/go-server/shared"
)

func main() {
	shared.Load()

	res := api.Prepare()
	defer res.Cleanup()

	if shared.IsDesktop() {
		cmd := desktop.OpenURL(fmt.Sprintf("http://localhost:%s", shared.Port()))

		server.Serve(&res)

		<-cmd
	} else {
		server.Serve(&res)

		c := make(chan os.Signal)
		signal.Notify(c, os.Interrupt)

		<-c
	}

	atexit.Exit(0)
}
