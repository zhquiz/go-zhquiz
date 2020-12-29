package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"

	"github.com/zhquiz/go-server/server/api"
	"github.com/zhquiz/go-server/shared"
)

// Serve starts the server.
// Runs `go func` by default.
func Serve(res *api.Resource) *gin.Engine {
	app := gin.Default()

	p := shared.Paths()
	app.Use(func(c *gin.Context) {
		if c.Request.Method == "GET" {
			if strings.HasPrefix(c.Request.URL.Path, "/docs/") || c.Request.URL.Path == "/docs" {
				static.Serve("/docs", static.LocalFile(filepath.Join(p.Dir, "docs"), true))(c)
				return
			}

			if strings.HasPrefix(c.Request.URL.Path, "/media/") {
				static.Serve("/media", static.LocalFile(p.MediaPath(), false))(c)
				return
			}

			static.Serve("/", static.LocalFile(filepath.Join(p.Dir, "public"), true))(c)
			return
		}
		c.Next()
	})

	if _, err := os.Stat(filepath.Join(p.Dir, "public")); os.IsNotExist(err) {
		app.GET("/", func(c *gin.Context) {
			c.Redirect(http.StatusTemporaryRedirect, "/docs")
		})
	} else {
		app.NoRoute(func(ctx *gin.Context) {
			method := ctx.Request.Method
			if method == "GET" {
				ctx.File(filepath.Join(p.Dir, "public", "index.html"))
			} else {
				ctx.Next()
			}
		})
	}

	res.Register(app)

	port := shared.Port()
	fmt.Printf("Server running at http://localhost:%s\n", port)
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: app,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	return app
}
