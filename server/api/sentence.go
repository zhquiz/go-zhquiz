package api

import (
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zhquiz/go-zhquiz/server/db"
	"gorm.io/gorm"
)

func routerSentence(apiRouter *gin.RouterGroup) {
	r := apiRouter.Group("/sentence")

	r.GET("/", func(ctx *gin.Context) {
		var query struct {
			Entry string `form:"entry" binding:"required"`
		}

		if e := ctx.BindQuery(&query); e != nil {
			ctx.AbortWithError(400, e)
			return
		}

		type Result struct {
			Chinese string `json:"chinese"`
			English string `json:"english"`
		}
		var result Result

		if r := resource.Zh.Current.Raw(`
		SELECT Chinese, English
		FROM sentence
		WHERE chinese = ?
		`, query.Entry).First(&result); r.Error != nil {
			if errors.Is(r.Error, gorm.ErrRecordNotFound) {
				ctx.AbortWithStatus(404)
				return
			}

			panic(r.Error)
		}

		ctx.JSON(200, result)
	})

	r.GET("/q", func(ctx *gin.Context) {
		var query struct {
			Q        string  `form:"q"`
			Page     *string `form:"page"`
			PerPage  *string `form:"perPage"`
			Generate *string `form:"generate"`
		}

		if e := ctx.BindQuery(&query); e != nil {
			ctx.AbortWithError(400, e)
			return
		}

		page := 0
		if query.Page != nil {
			i, err := strconv.Atoi(*query.Page)
			if err != nil {
				ctx.AbortWithError(400, errors.New("page must be int"))
				return
			}
			page = i
		}

		isCount := false
		perPage := 5
		if query.PerPage != nil {
			i, err := strconv.Atoi(*query.PerPage)
			if err != nil {
				ctx.AbortWithError(400, errors.New("perPage must be int"))
				return
			}
			perPage = i
			isCount = true
		}

		generate := 0
		if query.Generate != nil {
			i, err := strconv.Atoi(*query.Generate)
			if err != nil {
				ctx.AbortWithError(400, errors.New("generate must be int"))
				return
			}
			generate = i
		}

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

		var sentenceMin uint = 0
		if user.Meta.Settings.Sentence.Min != nil {
			sentenceMin = *user.Meta.Settings.Sentence.Min
		}

		var sentenceMax uint = 0
		if user.Meta.Settings.Sentence.Max != nil {
			sentenceMax = *user.Meta.Settings.Sentence.Max
		}

		type Result struct {
			ID      int64  `json:"-"`
			Chinese string `json:"chinese"`
			English string `json:"english"`
		}
		result := make([]Result, 0)

		andCond := []string{
			"[level] >= @levelMin",
			"[level] <= @levelMax",
		}

		cond := map[string]interface{}{
			"q":           query.Q,
			"levelMin":    levelMin,
			"levelMax":    levelMax,
			"sentenceMin": sentenceMin,
			"sentenceMax": sentenceMax,
		}

		if query.Q != "" {
			if regexp.MustCompile(`\p{Han}`).MatchString(query.Q) {
				cond["q"] = "%" + string(regexp.MustCompile(`[^\p{Han}]+`).ReplaceAll([]byte(query.Q), []byte("%"))) + "%"
				andCond = append(andCond, "chinese LIKE @q")
			} else {
				andCond = append(andCond, `id IN (
					SELECT id FROM sentence_q WHERE sentence_q MATCH @q
				)`)
			}
		} else {
			if sentenceMin != 0 {
				andCond = append(andCond, "length(chinese) >= @sentenceMin")
			}

			if sentenceMax != 0 {
				andCond = append(andCond, "length(chinese) <= @sentenceMax")
			}
		}

		order := "ORDER BY RANDOM()"
		if page != 0 {
			order = fmt.Sprintf(`
			ORDER BY level, frequency DESC
			LIMIT %d OFFSET %d
			`, perPage, (page-1)*perPage)
		}

		if r := resource.Zh.Current.Raw(fmt.Sprintf(`
		SELECT ID, Chinese, English
		FROM sentence
		WHERE %s
		%s
		`, strings.Join(andCond, " AND "), order), cond).Find(&result); r.Error != nil {
			panic(r.Error)
		}

		for i, it := range result {
			result[i].English = strings.Split(it.English, "\u001f")[0]
		}

		out := struct {
			Result []Result `json:"result"`
			Count  *int     `json:"count"`
		}{
			Result: result,
		}

		if isCount {
			var count int
			if r := resource.Zh.Current.Raw(fmt.Sprintf(`
			SELECT COUNT(*)
			FROM sentence
			WHERE %s
			`, strings.Join(andCond, " AND ")), cond).Scan(&count); r.Error != nil {
				panic(r.Error)
			}
			out.Count = &count
		}

		if generate < perPage {
			generate = perPage
		}

		if len(out.Result) <= generate {
			var dbSentences []db.Sentence
			if r := resource.DB.Current.Where("chinese LIKE ?", cond["q"]).Limit(generate - len(out.Result)).Order("RANDOM()").Find(&dbSentences); r.Error != nil {
				panic(r.Error)
			}

			for _, s := range dbSentences {
				out.Result = append(out.Result, Result{
					Chinese: s.Chinese,
					English: s.English,
				})
			}

			if len(out.Result) > generate {
				var newResult []Result
				for _, r := range out.Result {
					if len(newResult) < generate {
						newResult = append(newResult, r)
					}
				}
				out.Result = newResult
			}
		}

		if out.Result == nil {
			out.Result = make([]Result, 0)
		}

		ctx.JSON(200, out)
	})

	r.GET("/random", func(ctx *gin.Context) {
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

		var dbUser db.User

		if r := resource.DB.Current.Select("Meta").First(&dbUser); r.Error != nil {
			panic(r.Error)
		}

		where := "[type] = @type AND srs_level IS NOT NULL AND next_review IS NOT NULL"
		cond := map[string]interface{}{
			"type":        "sentence",
			"sentenceMin": dbUser.Meta.Settings.Sentence.Min,
			"sentenceMax": dbUser.Meta.Settings.Sentence.Max,
		}

		if dbUser.Meta.Settings.Sentence.Min != nil {
			where = where + " AND length(entry) >= @sentenceMin"
		}

		if dbUser.Meta.Settings.Sentence.Max != nil {
			where = where + " AND length(entry) <= @sentenceMax"
		}

		var existing []db.Quiz
		if r := resource.DB.Current.
			Where(where, cond).
			Find(&existing); r.Error != nil {
			panic(r.Error)
		}

		var entries []interface{}
		for _, it := range existing {
			entries = append(entries, it.Entry)
		}

		where = "[level] >= @levelMin AND [level] <= @level"
		cond = map[string]interface{}{
			"entries":  entries,
			"levelMin": levelMin,
			"level":    level,
		}

		if dbUser.Meta.Settings.Sentence.Min != nil {
			cond["sentenceMin"] = dbUser.Meta.Settings.Sentence.Min
			where = where + " AND length(chinese) >= @sentenceMin"
		}

		if dbUser.Meta.Settings.Sentence.Max != nil {
			cond["sentenceMax"] = dbUser.Meta.Settings.Sentence.Max
			where = where + " AND length(chinese) <= @sentenceMax"
		}

		if len(entries) > 0 {
			where = "chinese NOT IN @entries AND " + where
		}

		type Result struct {
			ID      int64   `json:"-"`
			Result  string  `json:"result"`
			English string  `json:"english"`
			Level   float64 `json:"level"`
		}
		var result []Result

		if r := resource.Zh.Current.Raw(fmt.Sprintf(`
		SELECT ID, chinese Result, Level, English
		FROM sentence
		WHERE %s
		`, where), cond).Find(&result); r.Error != nil {
			panic(r.Error)
		}

		if len(result) < 1 {
			result = []Result{}

			if r := resource.Zh.Current.Raw(fmt.Sprintf(`
			SELECT ID, chinese Result, Level
			FROM sentence
			WHERE %s
			`, where), cond).Find(&result); r.Error != nil {
				panic(r.Error)
			}
		}

		if len(result) < 1 {
			ctx.AbortWithError(404, fmt.Errorf("no matching entries found"))
			return
		}

		rand.Seed(time.Now().UnixNano())
		r := result[rand.Intn(len(result))]

		ctx.JSON(200, r)
	})
}
