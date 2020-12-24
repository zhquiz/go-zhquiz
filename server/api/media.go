package api

import (
	"github.com/gin-gonic/gin"
	"github.com/zhquiz/go-server/shared"
)

type tMediaRouter struct {
	Router *gin.RouterGroup
}

func (r tMediaRouter) init() {
	r.doUpload()
}

func (r tMediaRouter) doUpload() {
	r.Router.POST("/upload", func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			panic(err)
		}

		c.SaveUploadedFile(file, shared.Paths().MediaPath())

		c.JSON(201, gin.H{
			"url": "/media/" + file.Filename,
		})
	})
}
