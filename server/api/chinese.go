package api

import (
	"time"

	"github.com/gin-contrib/cache"
	"github.com/gin-gonic/gin"
	"github.com/tebeka/atexit"
	"github.com/yanyiwu/gojieba"
)

func routerChinese(apiRouter *gin.RouterGroup) {
	jieba := gojieba.NewJieba()
	atexit.Register(jieba.Free)

	r := apiRouter.Group("/chinese")

	r.GET("/jieba", cache.CachePage(persist, time.Hour, func(ctx *gin.Context) {
		var query struct {
			Q string `form:"q" binding:"required"`
		}

		if e := ctx.ShouldBindQuery(&query); e != nil {
			ctx.AbortWithError(400, e)
			return
		}

		ctx.JSON(200, gin.H{
			"result": jieba.CutAll(query.Q),
		})
	}))
}
