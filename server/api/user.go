package api

import (
	"fmt"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/zhquiz/go-server/server/db"
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
			return
		}

		qSel := strings.Split(query.Select, ",")
		sel := []string{}
		sMap := map[string]string{
			"level":                     "Meta",
			"levelMin":                  "Meta",
			"forvo":                     "Meta",
			"apiKey":                    "APIKey",
			"settings.quiz":             "Meta",
			"settings.level.whatToShow": "Meta",
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

		if r := resource.DB.Current.Select(sel).Where("ID = ?", userID).First(&dbUser); r.Error != nil {
			panic(r.Error)
		}

		out := gin.H{}
		outMap := map[string]func() interface{}{
			"level":                     func() interface{} { return dbUser.Meta.Level },
			"levelMin":                  func() interface{} { return dbUser.Meta.LevelMin },
			"forvo":                     func() interface{} { return *dbUser.Meta.Forvo },
			"apiKey":                    func() interface{} { return dbUser.APIKey },
			"settings.quiz":             func() interface{} { return dbUser.Meta.Settings.Quiz },
			"settings.level.whatToShow": func() interface{} { return dbUser.Meta.Settings.Level.WhatToShow },
		}

		for _, s := range qSel {
			v := outMap[s]
			if v != nil {
				out[s] = v()
			}
		}

		ctx.JSON(200, out)
	})

	r.PATCH("/", func(ctx *gin.Context) {
		userID := getUserID(ctx)
		if userID == "" {
			ctx.AbortWithStatus(401)
			return
		}

		var body struct {
			LevelMin *uint `json:"levelMin"`
			Level    *uint `json:"level"`
		}

		if e := ctx.ShouldBindJSON(&body); e != nil {
			ctx.AbortWithError(400, e)
			return
		}

		var dbUser db.User

		if r := resource.DB.Current.Where("ID = ?", userID).First(&dbUser); r.Error != nil {
			panic(r.Error)
		}

		if body.Level != nil {
			dbUser.Meta.Level = body.Level
		}

		if body.LevelMin != nil {
			dbUser.Meta.LevelMin = body.LevelMin
		}

		if r := resource.DB.Current.Save(&dbUser); r.Error != nil {
			panic(r.Error)
		}

		ctx.JSON(201, gin.H{
			"result": "updated",
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
