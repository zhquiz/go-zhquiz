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

func routerHanzi(apiRouter *gin.RouterGroup) {
	r := apiRouter.Group("/hanzi")

	r.GET("/", cache.CachePage(persist, time.Hour, func(ctx *gin.Context) {
		var query struct {
			Entry string `form:"entry" binding:"required"`
		}

		if e := ctx.ShouldBindQuery(&query); e != nil {
			ctx.AbortWithError(400, e)
			return
		}

		var out struct {
			Sub      string `json:"sub"`
			Sup      string `json:"sup"`
			Variants string `json:"variants"`
			Pinyin   string `json:"pinyin"`
			English  string `json:"english"`
		}

		if r := resource.Zh.Current.Model(&zh.Token{}).Select(`
		GROUP_CONCAT(token_sub.child, '') sub,
		GROUP_CONCAT(token_sup.child, '') sup,
		GROUP_CONCAT(token_var.child, '') variants,
		pinyin,
		english
		`).Joins(`
		LEFT JOIN token_sub ON token_sub.parent = entry
		LEFT JOIN token_sup ON token_sup.parent = entry
		LEFT JOIN token_var ON token_var.parent = entry
		`).Where("entry = ?", query.Entry).Group("entry").First(&out); r.Error != nil {
			if errors.Is(r.Error, sql.ErrNoRows) {
				ctx.AbortWithStatus(404)
				return
			}

			panic(r.Error)
		}

		ctx.JSON(200, out)
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
			Where("user_id = ? AND [type] = 'hanzi' AND srs_level IS NOT NULL AND next_review IS NOT NULL", userID).
			Find(&existing); r.Error != nil {
			panic(r.Error)
		}

		var entries []interface{}
		for _, it := range existing {
			entries = append(entries, it.Entry)
		}

		params := map[string]interface{}{
			"entries":  entries,
			"levelMin": levelMin,
			"level":    level,
		}

		where := "english IS NOT NULL AND hanzi_level >= @levelMin AND hanzi_level <= @level"
		if len(entries) > 0 {
			where = "entry NOT IN @entries AND " + where
		}

		var out struct {
			Result  string `json:"result"`
			English string `json:"english"`
			Level   int    `json:"level"`
		}

		if r := resource.Zh.Current.
			Model(&zh.Token{}).
			Select("entry Result, English, hanzi_level Level").
			Where(where, params).
			Order("RANDOM()").
			First(&out); r.Error != nil {
			if errors.Is(r.Error, sql.ErrNoRows) {
				where := "english IS NOT NULL"
				if len(entries) > 0 {
					where = "entry NOT IN @entries AND " + where
				}

				if r := resource.Zh.Current.
					Model(&zh.Token{}).
					Select("entry Result, English, hanzi_level level").
					Where(where, params).
					Order("RANDOM()").
					First(&out); r.Error != nil {
					if errors.Is(r.Error, sql.ErrNoRows) {
						ctx.AbortWithStatus(404)
						return
					}

					panic(r.Error)
				}
			} else {
				panic(r.Error)
			}
		}

		ctx.JSON(200, out)
	})
}
