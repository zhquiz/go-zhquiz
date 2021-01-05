package api

import (
	"os/exec"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/wangbin/jiebago"
	"github.com/zhquiz/go-zhquiz/shared"
)

func routerChinese(apiRouter *gin.RouterGroup) {
	var jieba jiebago.Segmenter
	jieba.LoadDictionary(filepath.Join(shared.ExecDir, "assets", "dict.txt"))

	r := apiRouter.Group("/chinese")

	r.GET("/jieba", func(ctx *gin.Context) {
		var query struct {
			Q string `form:"q" binding:"required"`
		}

		if e := ctx.ShouldBindQuery(&query); e != nil {
			ctx.AbortWithError(400, e)
			return
		}

		out := make([]string, 0)
		func(ch <-chan string) {
			for word := range ch {
				out = append(out, word)
			}
		}(jieba.CutAll(query.Q))

		ctx.JSON(200, gin.H{
			"result": out,
		})
	})

	speakCmd := shared.SpeakFn()
	if speakCmd != "" {
		r.POST("/speak", func(ctx *gin.Context) {
			var query struct {
				Q string `form:"q" binding:"required"`
			}

			if e := ctx.ShouldBindQuery(&query); e != nil {
				ctx.AbortWithError(400, e)
				return
			}

			cmd := exec.Command(speakCmd, query.Q)

			if e := cmd.Start(); e != nil {
				panic(e)
			}

			ctx.JSON(200, gin.H{
				"result": "success",
			})
		})
	}
}
