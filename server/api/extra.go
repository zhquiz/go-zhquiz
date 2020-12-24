package api

import (
	"github.com/gin-gonic/gin"
)

type tExtraRouter struct {
	Router *gin.RouterGroup
}

func (r tExtraRouter) init() {
	r.getQ()
}

func (r tExtraRouter) getQ() {
	r.Router.GET("/q", func(ctx *gin.Context) {
		var query struct {
			Select  string `binding:"required"`
			Sort    *string
			Page    *string `binding:"number"`
			PerPage *string `binding:"number"`
		}

		if e := ctx.ShouldBindQuery(&query); e != nil {
			panic(e)
		}

		ctx.JSON(200, gin.H{
			"result": r.jieba.CutAll(query.Q),
		})
	})
}
