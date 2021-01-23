package main

import (
	"os"
	"os/signal"

	"github.com/zhquiz/go-zhquiz/desktop"
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
		desktop.Start(&res)
	} else {
		c := make(chan os.Signal)
		signal.Notify(c, os.Interrupt, os.Kill)

		<-c
	}
}
