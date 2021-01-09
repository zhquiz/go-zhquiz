package api

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func getUserID(ctx *gin.Context) string {
	session := sessions.Default(ctx)
	k := session.Get("userID")

	if k == nil {
		return ""
	}

	return k.(string)
}
