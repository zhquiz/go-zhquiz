package api

import (
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/gin-contrib/cache"
	"github.com/gin-gonic/gin"
	"github.com/zhquiz/go-server/server/db"
	"github.com/zhquiz/go-server/server/zh"
)

func routerSentence(apiRouter *gin.RouterGroup) {
	r := apiRouter.Group("/sentence")

	r.GET("/", cache.CachePage(persist, time.Hour, func(ctx *gin.Context) {
		var query struct {
			Entry string `form:"entry" binding:"required"`
		}

		if e := ctx.ShouldBindQuery(&query); e != nil {
			ctx.AbortWithError(400, e)
			return
		}

		var zhSentence zh.Sentence

		if r := resource.Zh.Current.First(&zhSentence); r.Error != nil {
			if errors.Is(r.Error, sql.ErrNoRows) {
				ctx.AbortWithStatus(404)
				return
			}

			panic(r.Error)
		}

		ctx.JSON(200, gin.H{
			"chinese": zhSentence.Chinese,
			"english": zhSentence.English,
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
			Where("user_id = ? AND [type] = 'sentence' AND srs_level IS NOT NULL AND next_review IS NOT NULL", userID).
			Find(&existing); r.Error != nil {
			panic(r.Error)
		}

		var entries []interface{}
		for _, it := range existing {
			entries = append(entries, it.Entry)
		}

		where := "[level] >= @levelMin AND [level] <= @level"
		cond := map[string]interface{}{
			"entries":  entries,
			"levelMin": levelMin,
			"level":    level,
		}

		if len(entries) > 0 {
			where = "chinese NOT IN @entries AND " + where
		}

		var items []struct {
			Result  string `json:"result"`
			English string `json:"english"`
			Level   int    `json:"level"`
		}

		if r := resource.Zh.Current.
			Model(&zh.Sentence{}).
			Select("chinese AS Result", "english").
			Where(where, cond).
			Find(&items); r.Error != nil {
			panic(r.Error)
		}

		if len(items) < 1 {
			if r := resource.Zh.Current.
				Model(&zh.Sentence{}).
				Select("chinese AS Result", "english").
				Where("chinese NOT IN @entries", cond).
				Find(&items); r.Error != nil {
				panic(r.Error)
			}
		}

		if len(items) < 1 {
			ctx.AbortWithError(404, fmt.Errorf("no matching entries found"))
			return
		}

		ctx.JSON(200, items[rand.Intn(len(items))])
	})
}
