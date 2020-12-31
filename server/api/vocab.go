package api

import (
	"fmt"
	"math/rand"
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

		type Result struct {
			Simplified  string `json:"simplified"`
			Traditional string `json:"traditional"`
			Pinyin      string `json:"pinyin"`
			English     string `json:"english"`
		}

		preresult := []zh.Cedict{}

		if r := resource.Zh.Current.
			Model(&zh.Cedict{}).
			Joins("LEFT JOIN token ON token.entry = simplified").
			Where("Simplified = ? OR Traditional = ?", query.Entry, query.Entry).
			Group("cedict.ROWID").
			Order("token.frequency desc").
			Find(&preresult); r.Error != nil {
			panic(r.Error)
		}

		var result []Result

		for _, r := range preresult {
			result = append(result, Result{
				Simplified:  r.Simplified,
				Traditional: r.Traditional,
				Pinyin:      r.Pinyin,
				English:     r.English,
			})
		}

		if len(result) == 0 {
			result = make([]Result, 0)
		}

		ctx.AsciiJSON(200, gin.H{
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
			Level    string `form:"level"`
			LevelMin string `form:"levelMin"`
		}

		if e := ctx.ShouldBindQuery(&query); e != nil {
			ctx.AbortWithError(400, e)
			return
		}

		level := 60

		if query.Level != "" {
			v, e := strconv.Atoi(query.Level)
			if e != nil {
				ctx.AbortWithError(400, e)
				return
			}
			level = v
		}

		levelMin := 1

		if query.LevelMin != "" {
			v, e := strconv.Atoi(query.LevelMin)
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

		var items []struct {
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
			Find(&items); r.Error != nil {
			panic(r.Error)
		}

		if len(items) < 1 {
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
				Find(&items); r.Error != nil {
				panic(r.Error)
			}
		}

		if len(items) < 1 {
			ctx.AbortWithError(404, fmt.Errorf("no matched entries found"))
		}

		ctx.JSON(200, items[rand.Intn(len(items))])
	})
}
