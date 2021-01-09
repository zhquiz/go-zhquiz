package api

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zhquiz/go-zhquiz/server/db"
)

func routerVocab(apiRouter *gin.RouterGroup) {
	r := apiRouter.Group("/vocab")

	r.GET("/", func(ctx *gin.Context) {
		var query struct {
			Entry string `form:"entry" binding:"required"`
		}

		if e := ctx.BindQuery(&query); e != nil {
			ctx.AbortWithError(400, e)
			return
		}

		type Result struct {
			Simplified  string `json:"simplified"`
			Traditional string `json:"traditional"`
			Pinyin      string `json:"pinyin"`
			English     string `json:"english"`
		}

		result := make([]Result, 0)

		if r := resource.Zh.Current.Raw(`
		SELECT Simplified, Traditional, Pinyin, English English
		FROM vocab
		WHERE Simplified = ? OR Traditional = ?
		ORDER BY frequency DESC
		`, query.Entry, query.Entry).Find(&result); r.Error != nil {
			panic(r.Error)
		}

		ctx.JSON(200, gin.H{
			"result": result,
		})
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
			Simplified  string `json:"simplified"`
			Traditional string `json:"traditional"`
			Pinyin      string `json:"pinyin"`
			English     string `json:"english"`
		}

		var result []Result

		if r := resource.Zh.Current.Raw(`
		SELECT Simplified, Traditional, Pinyin, English
		FROM vocab
		WHERE simplified LIKE '%'||?||'%' OR traditional LIKE '%'||?||'%'
		ORDER BY frequency DESC
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

		if e := ctx.BindQuery(&query); e != nil {
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

		sqlString := "vocab_level >= @levelMin AND vocab_level <= @level"

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
		SELECT entry Result, (
			SELECT vocab.english FROM vocab WHERE simplified = entry AND frequency IS NOT NULL
		) English, vocab_level Level
		FROM token
		WHERE %s
		`, sqlString), cond).Find(&items); r.Error != nil {
			panic(r.Error)
		}

		if len(items) < 1 {
			items = []Item{}

			sqlString := "TRUE"

			if len(entries) > 0 {
				sqlString = "entry NOT IN @entries AND " + sqlString
			}

			if r := resource.Zh.Current.Raw(fmt.Sprintf(`
			SELECT entry Result, (
				SELECT vocab.english FROM vocab WHERE simplified = entry AND frequency IS NOT NULL
			) English, vocab_level Level
			FROM token
			WHERE %s
			`, sqlString), cond).Find(&items); r.Error != nil {
				panic(r.Error)
			}
		}

		if len(items) < 1 {
			ctx.AbortWithError(404, fmt.Errorf("no matched entries found"))
		}

		rand.Seed(time.Now().UnixNano())
		item := items[rand.Intn(len(items))]

		ctx.JSON(200, item)
	})
}
