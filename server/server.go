package server

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"

	"github.com/zhquiz/go-zhquiz/server/api"
	"github.com/zhquiz/go-zhquiz/server/rand"
	"github.com/zhquiz/go-zhquiz/shared"
)

// Serve starts the server.
// Runs `go func` by default.
func Serve(res *api.Resource) *gin.Engine {
	if !shared.IsDebug() {
		gin.SetMode(gin.ReleaseMode)
	}

	app := gin.New()

	app.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		ps := strings.SplitN(param.Path, "?", 2)
		path := ps[0]
		if len(ps) > 1 {
			q, e := url.QueryUnescape(ps[1])
			if e != nil {
				path += "?" + ps[1]
			} else {
				path += "?" + q
			}
		}

		out := []string{"[" + param.TimeStamp.Format(time.RFC3339) + "]"}
		out = append(out, param.Method)
		out = append(out, strconv.Itoa(param.StatusCode))
		out = append(out, param.Latency.String())
		out = append(out, path)

		if param.ErrorMessage != "" {
			out = append(out, param.ErrorMessage)
		}

		out = append(out, "\n")

		return strings.Join(out, " ")
	}))
	app.Use(gin.Recovery())

	app.Use(func(c *gin.Context) {
		b, _ := ioutil.ReadAll(c.Request.Body)

		if len(b) > 0 {
			c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(b))

			gin.DefaultWriter.Write([]byte(c.Request.Method + " " + c.Request.URL.Path + " body: "))
			gin.DefaultWriter.Write(b)
			gin.DefaultWriter.Write([]byte("\n"))
		}
		c.Next()
	})

	serverOptions := api.Options{
		Token: "",
	}

	if !shared.IsDebug() {
		token, _ := rand.GenerateRandomString(64)
		serverOptions.Token = token
	}

	app.Use(func(c *gin.Context) {
		if c.Request.Method == "GET" {
			c.SetCookie("csrf_token", serverOptions.Token, 2592000, "/", "localhost", false, true)

			static.Serve("/", static.LocalFile(filepath.Join(shared.ExecDir, "public"), true))(c)
			return
		}
		c.Next()
	})

	res.Register(app, &serverOptions)

	port := shared.Port()
	fmt.Printf("Server running at http://localhost:%d\n", port)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: app,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	return app
}
