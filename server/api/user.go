package api

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zhquiz/go-zhquiz/server/db"
)

func routerUser(apiRouter *gin.RouterGroup) {
	r := apiRouter.Group("/user")

	r.GET("/", func(ctx *gin.Context) {
		var query struct {
			Select string `form:"select" binding:"required"`
		}

		if e := ctx.BindQuery(&query); e != nil {
			ctx.AbortWithError(400, e)
			return
		}

		qSel := strings.Split(query.Select, ",")
		sel := []string{}
		sMap := map[string]string{
			"level":                     "json_extract(meta, '$.level') [level]",
			"levelMin":                  "json_extract(meta, '$.levelMin') levelMin",
			"forvo":                     "json_extract(meta, '$.forvo') forvo",
			"settings.quiz":             "json_extract(meta, '$.settings.quiz') [settings.quiz]",
			"settings.level.whatToShow": "json_extract(meta, '$.settings.level.whatToShow') [settings.level.whatToShow]",
			"settings.sentence.min":     "json_extract(meta, '$.settings.sentence.min') [settings.sentence.min]",
			"settings.sentence.max":     "json_extract(meta, '$.settings.sentence.max') [settings.sentence.max]",
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

		getter := map[string]interface{}{}

		if r := resource.DB.Current.Model(&db.User{}).Select(strings.Join(sel, ",")).First(&getter); r.Error != nil {
			panic(r.Error)
		}

		out := map[string]interface{}{}
		for k, v := range getter {
			switch t := v.(type) {
			case string:
				var v0 interface{}
				if err := json.Unmarshal([]byte(t), &v0); err != nil {
					out[k] = v
				} else {
					out[k] = v0
				}
			default:
				out[k] = v
			}
		}

		ctx.JSON(200, out)
	})

	r.PATCH("/", func(ctx *gin.Context) {
		var body struct {
			LevelMin    *uint  `json:"levelMin"`
			Level       *uint  `json:"level"`
			SentenceMin *uint  `json:"sentenceMin"`
			SentenceMax *uint  `json:"sentenceMax"`
			WhatToShow  string `json:"settings.level.whatToShow"`
		}

		if e := ctx.BindJSON(&body); e != nil {
			ctx.AbortWithError(400, e)
			return
		}

		var dbUser db.User

		if r := resource.DB.Current.First(&dbUser); r.Error != nil {
			panic(r.Error)
		}

		if body.Level != nil {
			dbUser.Meta.Level = body.Level
		}

		if body.LevelMin != nil {
			dbUser.Meta.LevelMin = body.LevelMin
		}

		if body.SentenceMin != nil {
			dbUser.Meta.Settings.Sentence.Min = body.SentenceMin
		}

		if body.SentenceMax != nil {
			if *body.SentenceMax == 0 {
				dbUser.Meta.Settings.Sentence.Max = nil
			} else {
				dbUser.Meta.Settings.Sentence.Max = body.SentenceMax
			}
		}

		if body.WhatToShow != "" {
			dbUser.Meta.Settings.Level.WhatToShow = body.WhatToShow
		}

		if r := resource.DB.Current.Save(&dbUser); r.Error != nil {
			panic(r.Error)
		}

		ctx.JSON(201, gin.H{
			"result": "updated",
		})
	})
}
