package api

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/zhquiz/go-server/server/db"
	"gorm.io/gorm"
)

func routerUser(apiRouter *gin.RouterGroup) {
	r := apiRouter.Group("/user")

	r.GET("/", func(ctx *gin.Context) {
		session := sessions.Default(ctx)
		userID := session.Get("userID").(string)
		if userID == "" {
			ctx.AbortWithStatus(401)
		}

		var query struct {
			Select string `form:"select" binding:"required"`
		}

		if e := ctx.ShouldBindQuery(&query); e != nil {
			ctx.AbortWithError(400, e)
		}

		qSel := strings.Split(query.Select, ",")
		sel := []string{}
		sMap := map[string]string{
			"level":    "Meta",
			"levelMin": "Meta",
			"forvo":    "Meta",
			"apiKey":   "APIKey",
		}

		for _, s := range qSel {
			v := sMap[s]
			if v != "" {
				sel = append(sel, v)
			}
		}

		if len(sel) == 0 {
			ctx.AbortWithError(400, fmt.Errorf("nothing to select"))
		}

		var dbUser db.User

		if r := resource.DB.Current.Select(sel).First(&dbUser, userID); errors.Is(r.Error, gorm.ErrRecordNotFound) {
			panic(r.Error)
		}

		ctx.JSON(200, gin.H{
			"level":    dbUser.Meta.Level,
			"levelMin": dbUser.Meta.LevelMin,
			"forvo":    dbUser.Meta.Forvo,
			"apiKey":   dbUser.APIKey,
		})
	})

	r.DELETE("/signOut", func(ctx *gin.Context) {
		session := sessions.Default(ctx)
		session.Clear()

		ctx.JSON(201, gin.H{
			"result": "signed out",
		})
	})
}
