package api

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zhquiz/zhquiz-desktop/server/db"
	"gorm.io/gorm"
)

func routerHanzi(apiRouter *gin.RouterGroup) {
	r := apiRouter.Group("/hanzi")

	r.GET("/", func(ctx *gin.Context) {
		var query struct {
			Entry string `form:"entry" binding:"required"`
		}

		if e := ctx.BindQuery(&query); e != nil {
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

		if r := resource.Zh.Current.Raw(`
		SELECT
			(
				SELECT GROUP_CONCAT(child, '') FROM token_sub WHERE parent = entry GROUP BY parent
			) Sub,
			(
				SELECT GROUP_CONCAT(child, '') FROM token_sup WHERE parent = entry GROUP BY parent
			) Sup,
			(
				SELECT GROUP_CONCAT(child, '') FROM token_var WHERE parent = entry GROUP BY parent
			) Variants,
			Pinyin,
			English
		FROM token
		WHERE [entry] = ?
		`, query.Entry).First(&out); r.Error != nil {
			if errors.Is(r.Error, gorm.ErrRecordNotFound) {
				ctx.AbortWithStatus(404)
				return
			}

			panic(r.Error)
		}

		ctx.JSON(200, out)
	})

	r.GET("/q", func(ctx *gin.Context) {
		var query struct {
			Q string `form:"q" binding:"required"`
		}

		if e := ctx.BindQuery(&query); e != nil {
			ctx.AbortWithError(400, e)
			return
		}

		type Result struct {
			Entry string `json:"entry"`
		}

		result := make([]Result, 0)

		if r := resource.Zh.Current.Raw(`
		SELECT Entry FROM token
		WHERE Entry IN (
			SELECT Entry FROM token_q WHERE token_q MATCH @Q AND length(Entry) = 1
		)
		ORDER BY frequency DESC
		`, query).Find(&result); r.Error != nil {
			panic(r.Error)
		}

		ctx.JSON(200, gin.H{
			"result": result,
		})
	})

	r.GET("/random", func(ctx *gin.Context) {
		var user db.User
		if r := resource.DB.Current.First(&user); r.Error != nil {
			panic(r.Error)
		}
		levelMin := *user.Meta.LevelMin
		if levelMin == 0 {
			levelMin = 1
		}
		levelMax := *user.Meta.Level
		if levelMax == 0 {
			levelMax = 60
		}

		var existing []db.Quiz
		if r := resource.DB.Current.
			Where("[type] = 'hanzi' AND srs_level IS NOT NULL AND next_review IS NOT NULL").
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
			"level":    levelMax,
		}

		where := "english IS NOT NULL AND hanzi_level >= @levelMin AND hanzi_level <= @level"
		if len(entries) > 0 {
			where = "entry NOT IN @entries AND " + where
		}

		type Item struct {
			Result  string `json:"result"`
			English string `json:"english"`
			Level   int    `json:"level"`
		}
		var items []Item

		if r := resource.Zh.Current.Raw(fmt.Sprintf(`
		SELECT entry Result, English, hanzi_level Level
		FROM token
		WHERE %s
		`, where), params).Find(&items); r.Error != nil {
			panic(r.Error)
		}

		if len(items) < 1 {
			where := "english IS NOT NULL"
			if len(entries) > 0 {
				where = "entry NOT IN @entries AND " + where
			}

			if r := resource.Zh.Current.Raw(fmt.Sprintf(`
			SELECT entry Result, English, hanzi_level Level
			FROM token
			WHERE %s
			`, where), params).Find(&items); r.Error != nil {
				panic(r.Error)
			}
		}

		if len(items) < 1 {
			ctx.AbortWithError(404, fmt.Errorf("no matching entries found"))
			return
		}

		rand.Seed(time.Now().UnixNano())
		ctx.JSON(200, items[rand.Intn(len(items))])
	})
}
