package api

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/zhquiz/go-server/server/db"
	"gorm.io/gorm"
)

func routerUser(apiRouter *gin.RouterGroup) {
	r := apiRouter.Group("/user")

	r.GET("/", func(ctx *gin.Context) {
		userID := getUserID(ctx)
		if userID == "" {
			ctx.AbortWithStatus(401)
			return
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
			return
		}

		var dbUser db.User

		if r := resource.DB.Current.Select(sel).Where("ID = ?", userID).First(&dbUser); errors.Is(r.Error, gorm.ErrRecordNotFound) {
			panic(r.Error)
		}

		out := gin.H{}
		outMap := map[string]func() interface{}{
			"level":    func() interface{} { return dbUser.Meta.Level },
			"levelMin": func() interface{} { return dbUser.Meta.LevelMin },
			"forvo":    func() interface{} { return *dbUser.Meta.Forvo },
			"apiKey":   func() interface{} { return dbUser.APIKey },
		}

		for _, s := range qSel {
			v := outMap[s]
			if v != nil {
				out[s] = v()
			}
		}

		log.Println(out)

		ctx.JSON(200, out)
	})

	r.DELETE("/signOut", func(ctx *gin.Context) {
		session := sessions.Default(ctx)
		session.Clear()

		ctx.JSON(201, gin.H{
			"result": "signed out",
		})
	})
}
