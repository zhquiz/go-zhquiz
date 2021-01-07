package api

import (
	"io/ioutil"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/zhquiz/go-zhquiz/shared"
	"gopkg.in/yaml.v2"
)

func routerLibrary(apiRouter *gin.RouterGroup) {
	r := apiRouter.Group("/library")

	r.GET("/q", func(ctx *gin.Context) {
		var query struct {
			Q string `form:"q" binding:"required"`
		}

		if e := ctx.ShouldBindQuery(&query); e != nil {
			ctx.AbortWithError(400, e)
			return
		}

		type Result struct {
			Entry string `json:"entry"`
		}
		result := make([]Result, 0)

		if r := resource.Zh.Current.Raw(`
			SELECT Entry FROM token_q WHERE token_q MATCH ?
			`, query.Q).Find(&result); r.Error != nil {
			panic(r.Error)
		}

		if len(result) > 0 {
			var entries []string
			for _, r := range result {
				entries = append(entries, r.Entry)
			}
			result = []Result{}

			if r := resource.Zh.Current.Raw(`
				SELECT simplified Entry FROM vocab WHERE simplified IN ? OR traditional IN ? GROUP BY simplified
				`, entries, entries).Find(&result); r.Error != nil {
				panic(r.Error)
			}
		}

		ctx.JSON(200, gin.H{
			"result": result,
		})
	})

	r.GET("/library.json", func(ctx *gin.Context) {
		type TSub struct {
			Title   string   `json:"title"`
			Entries []string `json:"entries"`
		}

		type T struct {
			Title    string `json:"title"`
			Children []TSub `json:"children"`
		}

		t := make([]T, 0)

		b, err := ioutil.ReadFile(filepath.Join(shared.ExecDir, "assets", "library.yaml"))
		if err != nil {
			panic(err)
		}

		if err := yaml.Unmarshal(b, &t); err != nil {
			panic(err)
		}

		ctx.JSON(200, t)
	})
}
