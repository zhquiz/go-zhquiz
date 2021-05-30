package api

import (
	"fmt"
	"math/rand"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zhquiz/zhquiz-desktop/server/db"
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

		where := "simplified LIKE @q OR traditional LIKE @q"
		cond := map[string]interface{}{
			"q": "%" + query.Q + "%",
		}

		if regexp.MustCompile(`^[^\p{Han}]+$`).MatchString(query.Q) {
			cond["q"] = query.Q
			where = "simplified IN (SELECT entry FROM token_q WHERE token_q MATCH @q)"
		}

		if r := resource.Zh.Current.Raw(fmt.Sprintf(`
		SELECT Simplified, Traditional, vocab.Pinyin, vocab.English
		FROM vocab
		WHERE %s
		ORDER BY frequency DESC
		LIMIT 10
		`, where), cond).Find(&result); r.Error != nil {
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
		var existing []db.Quiz
		if r := resource.DB.Current.
			Where("[type] = 'vocab' AND srs_level IS NOT NULL").
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
			Source   string `json:"source"`
			SRSLevel *int8  `json:"srs_level"`
		}
		var items []Item

		if r := resource.Zh.Current.Raw(`
		SELECT Entry, vocab_level Level, COALESCE((
			SELECT 'cedict' FROM vocab WHERE simplified = entry OR traditional = entry
		), 'hsk') Source
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
			Where("[type] = 'vocab' AND srs_level IS NOT NULL AND next_review IS NOT NULL").
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
			"level":    levelMax,
		}

		sqlString := "length(entry) > 1 AND vocab_level >= @levelMin AND vocab_level <= @level"

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
		SELECT entry Result, vocab_level Level
		FROM token
		WHERE %s
		`, sqlString), cond).Find(&items); r.Error != nil {
			panic(r.Error)
		}

		if len(items) > 0 {
			simp := make([]string, 0)
			for _, it := range items {
				simp = append(simp, it.Result)
			}

			rows, e := resource.Zh.Current.Raw(`
			SELECT simplified, english FROM vocab WHERE simplified IN ? AND frequency IS NOT NULL GROUP BY simplified
			`, simp).Rows()

			if e != nil {
				panic(e)
			}

			simpMap := make(map[string]string)
			for rows.Next() {
				var simplified string
				var english string
				if e := rows.Scan(&simplified, &english); e != nil {
					panic(e)
				}
				simpMap[simplified] = english
			}

			newItems := make([]Item, 0)
			for _, it := range items {
				it.English = simpMap[it.Result]
				if it.English != "" {
					newItems = append(newItems, it)
				}
			}

			items = newItems
		}

		if len(items) < 1 {
			items = []Item{}

			sqlString := "TRUE"

			if len(entries) > 0 {
				sqlString = "length(simplified) > 1 AND simplified NOT IN @entries AND " + sqlString
			}

			if r := resource.Zh.Current.Raw(fmt.Sprintf(`
			SELECT simplified Result, English
			FROM vocab
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

		if item.Level == 0 {
			resource.Zh.Current.Raw(`
			SELECT vocab_level FROM token WHERE entry = ?
			`, item.Result).Row().Scan(&item.Level)
		}

		ctx.JSON(200, item)
	})
}
