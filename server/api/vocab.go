package api

import (
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zhquiz/go-server/server/db"
)

func routerVocab(apiRouter *gin.RouterGroup) {
	r := apiRouter.Group("/vocab")

	r.GET("/", func(ctx *gin.Context) {
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

		var result []Result

		if r := resource.Zh.Current.Raw(`
		SELECT Simplified, Traditional, cedict.Pinyin, cedict.English English
		FROM cedict
		LEFT JOIN token ON token.entry = simplified
		WHERE Simplified = ? OR Traditional = ?
		GROUP BY cedict.ROWID
		ORDER BY token.frequency DESC
		`, query.Entry, query.Entry).Find(&result); r.Error != nil {
			if errors.Is(r.Error, sql.ErrNoRows) {
				result = make([]Result, 0)
			} else {
				panic(r.Error)
			}
		}

		ctx.JSON(200, gin.H{
			"result": result,
		})
	})

	r.GET("/q", func(ctx *gin.Context) {
		var query struct {
			Q string `form:"q" binding:"required"`
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

		var result []Result

		if r := resource.Zh.Current.Raw(`
		SELECT Simplified, Traditional, cedict.Pinyin, cedict.english English
		FROM cedict
		LEFT JOIN token ON token.entry = simplified
		WHERE simplified LIKE '%'||?||'%' OR traditional LIKE '%'||?||'%'
		GROUP BY cedict.ROWID
		ORDER BY token.frequency DESC
		LIMIT 10
		`, query.Q, query.Q).Find(&result); r.Error != nil {
			panic(r.Error)
		}

		if len(result) == 0 {
			result = make([]Result, 0)
		}

		ctx.JSON(200, gin.H{
			"result": result,
		})
	})

	r.GET("/level", func(ctx *gin.Context) {
		userID := getUserID(ctx)
		if userID == "" {
			ctx.AbortWithStatus(401)
			return
		}

		var existing []db.Quiz
		if r := resource.DB.Current.
			Where("user_id = ? AND [type] = 'vocab' AND srs_level IS NOT NULL", userID).
			Find(&existing); r.Error != nil {
			panic(r.Error)
		}

		srsLevelMap := map[string]*int8{}
		for _, it := range existing {
			srsLevelMap[it.Entry] = it.SRSLevel
		}

		type Item struct {
			Entry    string `json:"entry"`
			Level    int    `json:"level"`
			SRSLevel *int8  `json:"srs_level"`
		}
		var items []Item

		if r := resource.Zh.Current.Raw(`
		SELECT Entry, vocab_level Level
		FROM token
		WHERE vocab_level IS NOT NULL
		`).Find(&items); r.Error != nil {
			panic(r.Error)
		}

		for i, it := range items {
			items[i].SRSLevel = srsLevelMap[it.Entry]
		}

		ctx.JSON(200, gin.H{
			"result": items,
		})
	})

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

		type Item struct {
			Result  string `json:"result"`
			English string `json:"english"`
			Level   int    `json:"level"`
		}
		var items []Item

		if r := resource.Zh.Current.Raw(fmt.Sprintf(`
		SELECT entry Result, cedict.english English, vocab_level Level
		FROM token
		LEFT JOIN cedict ON cedict.simplified = entry
		WHERE %s
		GROUP BY entry
		`, sqlString), cond).Find(&items); r.Error != nil {
			panic(r.Error)
		}

		if len(items) < 1 {
			items = []Item{}

			sqlString := "cedict.english IS NOT NULL"

			if len(entries) > 0 {
				sqlString = "entry NOT IN @entries AND " + sqlString
			}

			if r := resource.Zh.Current.Raw(fmt.Sprintf(`
			SELECT entry Result, cedict.english English, vocab_level Level
			FROM token
			LEFT JOIN cedict ON cedict.simplified = entry
			WHERE %s
			GROUP BY entry
			`, sqlString), cond).Find(&items); r.Error != nil {
				panic(r.Error)
			}
		}

		if len(items) < 1 {
			ctx.AbortWithError(404, fmt.Errorf("no matched entries found"))
		}

		item := items[rand.Intn(len(items))]

		ctx.JSON(200, item)
	})
}
