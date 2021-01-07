package api

import (
	"os/exec"

	"github.com/gin-gonic/gin"
	"github.com/zhquiz/go-zhquiz/shared"
)

func routerChinese(apiRouter *gin.RouterGroup) {
	r := apiRouter.Group("/chinese")

	r.GET("/jieba", func(ctx *gin.Context) {
		var query struct {
			Q string `form:"q" binding:"required"`
		}

		if e := ctx.ShouldBindQuery(&query); e != nil {
			ctx.AbortWithError(400, e)
			return
		}

		ctx.JSON(200, gin.H{
			"result": cutChinese(query.Q),
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

func cutChinese(s string) []string {
	out := make([]string, 0)
	func(ch <-chan string) {
		for word := range ch {
			out = append(out, word)
		}
	}(jieba.CutAll(s))

	return out
}
