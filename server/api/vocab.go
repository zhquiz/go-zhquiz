package api

import (
	"database/sql"
	"errors"
	"strconv"
	"time"

	"github.com/gin-contrib/cache"
	"github.com/gin-gonic/gin"
	"github.com/zhquiz/go-server/server/db"
	"github.com/zhquiz/go-server/server/zh"
)

func routerVocab(apiRouter *gin.RouterGroup) {
	r := apiRouter.Group("/vocab")

	r.GET("/", cache.CachePage(persist, time.Hour, func(ctx *gin.Context) {
		var query struct {
			Entry string `form:"entry" binding:"required"`
		}

		if e := ctx.ShouldBindQuery(&query); e != nil {
			ctx.AbortWithError(400, e)
			return
		}

		result := []zh.Cedict{}

		if r := resource.Zh.Current.
			Model(&zh.Cedict{}).
			Joins("LEFT JOIN token ON token.entry = simplified").
			Where("Simplified = ? OR Traditional = ?", query.Entry, query.Entry).
			Group("cedict.ROWID").
			Order("token.frequency desc").
			Find(&result); r.Error != nil {
			panic(r.Error)
		}

		if len(result) == 0 {
			result = make([]zh.Cedict, 0)
		}

		ctx.JSON(200, gin.H{
			"result": result,
		})
	}))

	r.GET("/random", func(ctx *gin.Context) {
		userID := getUserID(ctx)
		if userID == "" {
			ctx.AbortWithStatus(401)
			return
		}

		var query struct {
			Level    *string
			LevelMin *string
		}

		if e := ctx.ShouldBindQuery(&query); e != nil {
			ctx.AbortWithError(400, e)
			return
		}

		level := 60

		if query.Level != nil {
			v, e := strconv.Atoi(*query.Level)
			if e != nil {
				ctx.AbortWithError(400, e)
				return
			}
			level = v
		}

		levelMin := 1

		if query.LevelMin != nil {
			v, e := strconv.Atoi(*query.LevelMin)
			if e != nil {
				ctx.AbortWithError(400, e)
				return
			}
			levelMin = v
		}

		var existing []db.Quiz
		if r := resource.DB.Current.
			Where("user_id = ? AND [type] = 'vocab' AND srs_level IS NOT NULL AND next_review IS NOT NULL", userID).
			Find(&existing); r.Error != nil {
			panic(r.Error)
		}

		var entries []interface{}
		for _, it := range existing {
			entries = append(entries, it.Entry)
		}

		cond := map[string]interface{}{
			"entries":  entries,
			"levelMin": levelMin,
			"level":    level,
		}

		sqlString := "cedict.english IS NOT NULL AND vocab_level >= @levelMin AND vocab_level <= @level"

		if len(entries) > 0 {
			sqlString = "entry NOT IN @entries AND " + sqlString
		}

		var out struct {
			Result  string `json:"result"`
			English string `json:"english"`
			Level   int    `json:"level"`
		}

		if r := resource.Zh.Current.
			Model(&zh.Token{}).
			Select("entry AS Result", "cedict.english AS English", "vocab_level AS Level").
			Joins("LEFT JOIN cedict ON cedict.simplified = entry").
			Where(sqlString, cond).
			Group("entry").
			Order("RANDOM()").
			First(&out); r.Error != nil {
			if errors.Is(r.Error, sql.ErrNoRows) {
				sqlString := "cedict.english IS NOT NULL"

				if len(entries) > 0 {
					sqlString = "entry NOT IN @entries AND " + sqlString
				}

				if r := resource.Zh.Current.
					Model(&zh.Token{}).
					Select("entry AS Result", "cedict.english AS English", "vocab_level AS Level").
					Joins("LEFT JOIN cedict ON cedict.simplified = entry").
					Where(sqlString, cond).
					Group("entry").
					Order("RANDOM()").
					First(&out); r.Error != nil {
					if errors.Is(r.Error, sql.ErrNoRows) {
						ctx.AbortWithStatus(404)
					} else {
						panic(r.Error)
					}
					return
				}
			} else {
				panic(r.Error)
			}
		}

		ctx.JSON(200, out)
	})
}
