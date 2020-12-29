package api

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/cache"
	"github.com/gin-gonic/gin"
	"github.com/zhquiz/go-server/server/db"
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

		stmt, e := resource.Zh.Current.Prepare(`
		SELECT
			GROUP_CONCAT(token_sub.child, '') sub,
			GROUP_CONCAT(token_sup.child, '') sup,
			GROUP_CONCAT(token_var.child, '') variants,
			pinyin,
			english
		FROM token
		LEFT JOIN token_sub ON token_sub.parent = entry
		LEFT JOIN token_sup ON token_sup.parent = entry
		LEFT JOIN token_var ON token_var.parent = entry
		WHERE entry = ?
		GROUP BY entry
		`)

		if e != nil {
			panic(e)
		}

		r := stmt.QueryRow(query.Entry)

		if e := r.Err(); e != nil {
			panic(e)
		}

		var out struct {
			Sub      string `json:"sub"`
			Sup      string `json:"sup"`
			Variants string `json:"variants"`
			Pinyin   string `json:"pinyin"`
			English  string `json:"english"`
		}
		if e := r.Scan(out.Sub, out.Sup, out.Variants, out.Pinyin, out.English); e == sql.ErrNoRows {
			ctx.AbortWithStatus(404)
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
			Level    *string `json:"level"`
			LevelMin *string `json:"levelMin"`
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
		if e := resource.DB.Current.
			Where("user_id = ? AND [type] = 'hanzi' AND srs_level IS NOT NULL AND next_review IS NOT NULL", userID).
			Find(&existing); e != nil {
			panic(e)
		}

		var its []interface{}
		for _, it := range existing {
			its = append(its, it.Entry)
		}

		its = append(its, levelMin, level)

		sqlString := `
		SELECT
			entry,
			english,
			hanzi_level
		FROM token
		WHERE english IS NOT NULL AND hanzi_level >= ? AND hanzi_level <= ?
		ORDER BY RANDOM()`

		if len(existing) > 0 {
			sqlString = fmt.Sprintf(`
			SELECT
				entry,
				english,
				hanzi_level
			FROM token
			WHERE entry NOT IN (%s) AND english IS NOT NULL AND hanzi_level >= ? AND hanzi_level <= ?
			ORDER BY RANDOM()`, string(strings.Repeat(",?", len(existing))[1:]))
		}

		stmt, e := resource.Zh.Current.Prepare(sqlString)

		if e != nil {
			panic(e)
		}

		r := stmt.QueryRow(its...)

		if e := r.Err(); e != nil {
			panic(e)
		}

		var out struct {
			Result  string `json:"result"`
			English string `json:"english"`
			Level   int    `json:"level"`
		}
		if e := r.Scan(out.Result, out.English, out.Level); e == sql.ErrNoRows {
			ctx.AbortWithStatus(404)
			return
		}

		ctx.JSON(200, out)
	})
}
