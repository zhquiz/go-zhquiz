package api

import (
	"github.com/gin-gonic/gin"
	"github.com/zhquiz/go-server/shared"
)

func routerMedia(apiRouter *gin.RouterGroup) {
	r := apiRouter.Group("/media")

	r.POST("/upload", func(ctx *gin.Context) {
		userID := getUserID(ctx)
		if userID == "" {
			ctx.AbortWithStatus(401)
			return
		}

		file, err := ctx.FormFile("file")
		if err != nil {
			ctx.AbortWithError(400, err)
			return
		}

		ctx.SaveUploadedFile(file, shared.Paths().MediaPath())

		ctx.JSON(201, gin.H{
			"url": "/media/" + file.Filename,
		})
	})
}
