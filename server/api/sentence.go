package api

import (
	"errors"
	"fmt"
	"math/rand"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
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
			Q       string  `form:"q" binding:"required"`
			Page    *string `form:"page"`
			PerPage *string `form:"perPage"`
		}

		if e := ctx.BindQuery(&query); e != nil {
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
		result := make([]Result, 0)

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

		if isCount {
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

		if len(out.Result) <= 5 {
			moreResult := func() []Result {
				doc, err := goquery.NewDocument(fmt.Sprintf("http://www.jukuu.com/search.php?q=%s", url.QueryEscape(query.Q)))
				if err != nil {
					return []Result{}
				}

				moreResult := make([]Result, 10-len(out.Result))

				doc.Find("table tr.c td:last-child").Each(func(i int, item *goquery.Selection) {
					if i < len(moreResult) {
						moreResult[i].Chinese = item.Text()
					}
				})

				doc.Find("table tr.e td:last-child").Each(func(i int, item *goquery.Selection) {
					if i < len(moreResult) {
						moreResult[i].English = item.Text()
					}
				})

				cleaned := make([]Result, 0)
				for _, r := range moreResult {
					if r.Chinese != "" {
						cleaned = append(cleaned, r)
					}
				}

				return cleaned
			}()

			out.Result = append(out.Result, moreResult...)
		}

		ctx.JSON(200, out)
	})

	r.GET("/all", func(ctx *gin.Context) {
		userID := getUserID(ctx)
		if userID == "" {
			ctx.AbortWithStatus(401)
			return
		}

		var query struct {
			Page     string `form:"page" binding:"required"`
			PerPage  string `form:"perPage" binding:"required"`
			Level    string `form:"level" binding:"required"`
			LevelMin string `form:"levelMin" binding:"required"`
		}

		if e := ctx.BindQuery(&query); e != nil {
			ctx.AbortWithError(400, e)
			return
		}

		page := 1
		{
			i, err := strconv.Atoi(query.Page)
			if err != nil {
				ctx.AbortWithError(400, errors.New("page must be int"))
				return
			}
			page = i
		}

		isCount := false
		perPage := 5
		{
			i, err := strconv.Atoi(query.PerPage)
			if err != nil {
				ctx.AbortWithError(400, errors.New("perPage must be int"))
				return
			}
			perPage = i
			isCount = true
		}

		level := 60
		{
			i, err := strconv.Atoi(query.Level)
			if err != nil {
				ctx.AbortWithError(400, errors.New("level must be int"))
				return
			}
			level = i
		}

		levelMin := 1
		{
			i, err := strconv.Atoi(query.LevelMin)
			if err != nil {
				ctx.AbortWithError(400, errors.New("levelMin must be int"))
				return
			}
			levelMin = i
		}

		where := "level <= @level AND level >= @levelMin"
		cond := map[string]interface{}{
			"level":    level,
			"levelMin": levelMin,
		}

		var dbUser db.User

		if r := resource.DB.Current.Select("Meta").Where("ID = ?", userID).First(&dbUser); r.Error != nil {
			panic(r.Error)
		}

		if dbUser.Meta.Settings.Sentence.Min != nil {
			cond["sentenceMin"] = dbUser.Meta.Settings.Sentence.Min
			where = where + " AND length(chinese) >= @sentenceMin"
		}

		if dbUser.Meta.Settings.Sentence.Max != nil {
			cond["sentenceMax"] = dbUser.Meta.Settings.Sentence.Max
			where = where + " AND length(chinese) <= @sentenceMax"
		}

		type Result struct {
			Chinese string `json:"chinese"`
			English string `json:"english"`
		}
		var result []Result

		if r := resource.Zh.Current.Raw(fmt.Sprintf(`
		SELECT Chinese, English
		FROM sentence
		WHERE %s
		ORDER BY level, frequency DESC
		LIMIT %d OFFSET %d
		`, where, perPage, (page-1)*perPage), cond).Find(&result); r.Error != nil {
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
			if r := resource.Zh.Current.Raw(fmt.Sprintf(`
			SELECT COUNT(*) Count
			FROM sentence
			WHERE %s
			`, where), cond).Scan(&count); r.Error != nil {
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

		if r := resource.DB.Current.Select("Meta").Where("ID = ?", userID).First(&dbUser); r.Error != nil {
			panic(r.Error)
		}

		where := "user_id = @userID AND [type] = 'sentence' AND srs_level IS NOT NULL AND next_review IS NOT NULL"
		cond := map[string]interface{}{
			"userID": userID,
		}

		if dbUser.Meta.Settings.Sentence.Min != nil {
			cond["sentenceMin"] = dbUser.Meta.Settings.Sentence.Min
			where = where + " AND length(entry) >= @sentenceMin"
		}

		if dbUser.Meta.Settings.Sentence.Max != nil {
			cond["sentenceMax"] = dbUser.Meta.Settings.Sentence.Max
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

		rand.Seed(time.Now().UnixNano())
		r := result[rand.Intn(len(result))]

		ctx.JSON(200, r)
	})
}
