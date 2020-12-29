package api

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/cache"
	"github.com/gin-gonic/gin"
	"github.com/zhquiz/go-server/server/db"
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

		stmt, e := resource.Zh.Current.Prepare(`
		SELECT
			Simplified,
			Traditional,
			cedict.pinyin 	Pinyin,
			cedict.english	English
		FROM cedict
		LEFT JOIN token ON token.entry = simplified
		WHERE Simplified = ? OR Traditional = ?
		GROUP BY cedict.ROWID
		ORDER BY token.frequency DESC
		`)

		if e != nil {
			panic(e)
		}

		r := stmt.QueryRow(query.Entry)

		if e := r.Err(); e != nil {
			panic(e)
		}

		var out struct {
			Simplified  string
			Traditional string
			Pinyin      string
			English     string
		}
		if e := r.Scan(&out); e == sql.ErrNoRows {
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

		var its []interface{}
		for _, it := range existing {
			its = append(its, it.Entry)
		}

		entries := its
		its = append(its, levelMin, level)

		sqlString := `
		SELECT
			entry,
			cedict.english english,
			vocab_level
		FROM token
		LEFT JOIN cedict ON cedict.simplified = entry
		WHERE cedict.english IS NOT NULL AND vocab_level >= ? AND vocab_level <= ?
		GROUP BY entry
		ORDER BY RANDOM()`

		if len(entries) > 0 {
			sqlString = fmt.Sprintf(`
			SELECT
				entry,
				cedict.english english,
				vocab_level
			FROM token
			LEFT JOIN cedict ON cedict.simplified = entry
			WHERE entry NOT IN (%s) AND cedict.english IS NOT NULL AND vocab_level >= ? AND vocab_level <= ?
			GROUP BY entry
			ORDER BY RANDOM()`, string(strings.Repeat(",?", len(entries))[1:]))
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
		if e := r.Scan(&out.Result, &out.English, &out.Level); errors.Is(e, sql.ErrNoRows) {
			sqlString := `
			SELECT
				entry,
				cedict.english english,
				vocab_level
			FROM token
			LEFT JOIN cedict ON cedict.simplified = entry
			WHERE cedict.english IS NOT NULL
			GROUP BY entry
			ORDER BY RANDOM()`

			if len(entries) > 0 {
				sqlString = fmt.Sprintf(`
				SELECT
					entry,
					cedict.english english,
					vocab_level
				FROM token
				LEFT JOIN cedict ON cedict.simplified = entry
				WHERE entry NOT IN (%s) AND cedict.english IS NOT NULL
				GROUP BY entry
				ORDER BY RANDOM()`, string(strings.Repeat(",?", len(entries))[1:]))
			}

			stmt, e := resource.Zh.Current.Prepare(sqlString)

			if e != nil {
				panic(e)
			}

			r := stmt.QueryRow(entries...)

			if e := r.Err(); e != nil {
				panic(e)
			}

			if e := r.Scan(&out.Result, &out.English, &out.Level); errors.Is(e, sql.ErrNoRows) {
				ctx.AbortWithStatus(404)
				return
			}
		}

		ctx.JSON(200, out)
	})
}
