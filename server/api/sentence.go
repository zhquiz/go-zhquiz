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

		stmt, e := resource.Zh.Current.Prepare(`
		SELECT
			Chinese,
			English
		FROM sentence
		WHERE Chinese = ?
		ORDER BY frequency DESC
		`)

		if e != nil {
			panic(e)
		}

		r := stmt.QueryRow(query.Entry)

		if e := r.Err(); e != nil {
			panic(e)
		}

		var out struct {
			Chinese string
			English string
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
			Where("user_id = ? AND [type] = 'sentence' AND srs_level IS NOT NULL AND next_review IS NOT NULL", userID).
			Find(&existing); r.Error != nil {
			panic(r.Error)
		}

		var its []interface{}
		for _, it := range existing {
			its = append(its, it.Entry)
		}

		sqlString := `
		SELECT
			Chinese,
			English
		FROM sentence
		ORDER BY RANDOM()`

		if len(its) > 0 {
			sqlString = fmt.Sprintf(`
			SELECT
				Chinese,
				English
			FROM sentence
			WHERE Chinese NOT IN (%s)
			ORDER BY RANDOM()`, string(strings.Repeat(",?", len(its))[1:]))
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
		if e := r.Scan(&out.Result, &out.English); errors.Is(e, sql.ErrNoRows) {
			ctx.AbortWithStatus(404)
			return
		}

		out.Level = level + levelMin

		ctx.JSON(200, out)
	})
}
