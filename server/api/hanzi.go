package api

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/gin-contrib/cache"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/zhquiz/go-server/server/db"
	"gopkg.in/sakura-internet/go-rison.v3"
)

type tHanziRouter struct {
	Router *gin.RouterGroup
}

func (r tHanziRouter) init() {
	r.getMatch()
}

func (r tHanziRouter) getMatch() {
	r.Router.GET("/", cache.CachePage(store, time.Hour, func(ctx *gin.Context) {
		var query struct {
			Entry string
		}

		if e := ctx.ShouldBindQuery(&query); e != nil {
			panic(e)
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
			Sub      string
			Sup      string
			Variants string
			Pinyin   string
			English  string
		}
		if e := r.Scan(out.Sub, out.Sup, out.Variants, out.Pinyin, out.English); e == sql.ErrNoRows {
			ctx.AbortWithStatus(404)
		}

		ctx.JSON(200, out)
	}))
}

func (r tHanziRouter) getRandom() {
	r.Router.GET("/random", func(ctx *gin.Context) {
		session := sessions.Default(ctx)
		userID := session.Get("userID").(string)
		if userID == "" {
			ctx.AbortWithStatus(401)
		}

		var query struct {
			RS string `form:"_"`
		}

		var rs struct {
			Level    *int
			LevelMin *int `json:"perPage"`
		}

		if e := ctx.ShouldBindQuery(&query); e != nil {
			panic(e)
		}

		if e := rison.Unmarshal([]byte(query.RS), &rs, rison.Rison); e != nil {
			panic(e)
		}

		if rs.Level == nil {
			*rs.Level = 60
		}

		if rs.LevelMin == nil {
			*rs.LevelMin = 1
		}

		var existing []db.Quiz
		if e := resource.DB.Current.
			Where("UserID = ? AND [type] = 'hanzi' AND SRSLevel IS NOT NULL AND NextReview IS NOT NULL", userID).
			Find(&existing); e != nil {
			panic(e)
		}

		var its []interface{}
		for _, it := range existing {
			its = append(its, it.Entry)
		}

		its = append(its, *rs.LevelMin, *rs.Level)

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
			WHERE entry NOT IN (%s) AND english IS NOT NULL
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
			Result  string
			English string
			Level   int
		}
		if e := r.Scan(out.Result, out.English, out.Level); e == sql.ErrNoRows {
			ctx.AbortWithStatus(404)
		}

		ctx.JSON(200, out)
	})
}