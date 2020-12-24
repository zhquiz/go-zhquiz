package api

import (
	"github.com/gin-gonic/gin"
	"github.com/yanyiwu/gojieba"
)

type tChineseRouter struct {
	Router *gin.RouterGroup
	jieba  *gojieba.Jieba
}

func (r tChineseRouter) init() {
	r.jieba = gojieba.NewJieba()
	// defer r.jieba.Free()

	r.getJieba()
}

func (r tChineseRouter) getJieba() {
	r.Router.GET("/jieba", func(ctx *gin.Context) {
		var query struct {
			Q string
		}

		if e := ctx.ShouldBindQuery(&query); e != nil {
			panic(e)
		}

		ctx.JSON(200, gin.H{
			"result": r.jieba.CutAll(query.Q),
		})
	})
}
