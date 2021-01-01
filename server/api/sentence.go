package api

import (
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zhquiz/go-server/server/db"
)

func routerSentence(apiRouter *gin.RouterGroup) {
	r := apiRouter.Group("/sentence")

	r.GET("/", func(ctx *gin.Context) {
		var query struct {
			Entry string `form:"entry" binding:"required"`
		}

		if e := ctx.ShouldBindQuery(&query); e != nil {
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
		`).First(&result); r.Error != nil {
			if errors.Is(r.Error, sql.ErrNoRows) {
				ctx.AbortWithStatus(404)
				return
			}

			panic(r.Error)
		}

		ctx.JSON(200, result)
	})

	r.GET("/q", func(ctx *gin.Context) {
		var query struct {
			Q       string  `form:"q" binding:"required"`
			Page    *string `form:"page"`
			PerPage *string `form:"perPage"`
		}

		if e := ctx.ShouldBindQuery(&query); e != nil {
			ctx.AbortWithError(400, e)
			return
		}

		page := 1
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

		type Result struct {
			Chinese string `json:"chinese"`
			English string `json:"english"`
		}
		var result []Result

		if r := resource.Zh.Current.Raw(fmt.Sprintf(`
		SELECT Chinese, English
		FROM sentence
		WHERE chinese LIKE '%%'||?||'%%'
		ORDER BY level, frequency DESC
		LIMIT %d OFFSET %d
		`, perPage, (page-1)*perPage), query.Q).Find(&result); r.Error != nil {
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

		if len(out.Result) == 0 {
			out.Result = make([]Result, 0)
		} else if isCount {
			var count int
			if r := resource.Zh.Current.Raw(`
			SELECT COUNT(*)
			FROM sentence
			WHERE chinese LIKE '%'||?||'%'
			`, query.Q).Scan(&count); r.Error != nil {
				panic(r.Error)
			}
			out.Count = &count
		}

		ctx.JSON(200, out)
	})

	r.GET("/all", func(ctx *gin.Context) {
		var query struct {
			Page    *string `form:"page"`
			PerPage *string `form:"perPage"`
		}

		if e := ctx.ShouldBindQuery(&query); e != nil {
			ctx.AbortWithError(400, e)
			return
		}

		page := 1
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

		type Result struct {
			Chinese string `json:"chinese"`
			English string `json:"english"`
		}
		var result []Result

		if r := resource.Zh.Current.Raw(fmt.Sprintf(`
		SELECT Chinese, English
		FROM sentence
		ORDER BY level, frequency DESC
		LIMIT %d OFFSET %d
		`, perPage, (page-1)*perPage)).Find(&result); r.Error != nil {
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

		if len(out.Result) == 0 {
			out.Result = make([]Result, 0)
		} else if isCount {
			var count int
			if r := resource.Zh.Current.Raw(`
			SELECT COUNT(*) Count
			FROM sentence
			`).Scan(&count); r.Error != nil {
				panic(r.Error)
			}
			out.Count = &count
		}

		ctx.JSON(200, out)
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
			Where("user_id = ? AND [type] = 'sentence' AND srs_level IS NOT NULL AND next_review IS NOT NULL", userID).
			Find(&existing); r.Error != nil {
			panic(r.Error)
		}

		var entries []interface{}
		for _, it := range existing {
			entries = append(entries, it.Entry)
		}

		where := "[level] >= @levelMin AND [level] <= @level"
		cond := map[string]interface{}{
			"entries":  entries,
			"levelMin": levelMin,
			"level":    level,
		}

		if len(entries) > 0 {
			where = "chinese NOT IN @entries AND " + where
		}

		type Result struct {
			Result  string  `json:"result"`
			English string  `json:"english"`
			Level   float64 `json:"level"`
		}
		var result []Result

		if r := resource.Zh.Current.Raw(fmt.Sprintf(`
		SELECT chinese Result, English, Level
		FROM sentence
		WHERE %s
		`, where), cond).Find(&result); r.Error != nil {
			panic(r.Error)
		}

		if len(result) < 1 {
			result = []Result{}

			if r := resource.Zh.Current.Raw(fmt.Sprintf(`
			SELECT chinese Result, English, Level
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

		r := result[rand.Intn(len(result))]

		ctx.JSON(200, r)
	})
}
