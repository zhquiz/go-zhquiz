package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/tebeka/atexit"
	"github.com/zhquiz/go-zhquiz/desktop"
	"github.com/zhquiz/go-zhquiz/server"
	"github.com/zhquiz/go-zhquiz/server/api"
	"github.com/zhquiz/go-zhquiz/shared"
)

func main() {
	shared.Load()

	res := api.Prepare()
	defer res.Cleanup()

	if shared.IsDesktop() {
		cmd := desktop.OpenURL(fmt.Sprintf("http://localhost:%s/random", shared.Port()))

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
