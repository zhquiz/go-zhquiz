package api

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func routerUser(apiRouter *gin.RouterGroup) {
	r := apiRouter.Group("/user")

	r.DELETE("/signOut", func(ctx *gin.Context) {
		session := sessions.Default(ctx)
		session.Clear()

		ctx.JSON(201, gin.H{
			"result": "signed out",
		})
	})
}
