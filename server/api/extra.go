package api

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zhquiz/go-zhquiz/server/db"
	"gorm.io/gorm"
)

func routerExtra(apiRouter *gin.RouterGroup) {
	r := apiRouter.Group(("/extra"))

	sMap := map[string]string{
		"id":          "extra.id id",
		"chinese":     "extra.chinese chinese",
		"pinyin":      "extra.pinyin pinyin",
		"english":     "extra_q.english english",
		"type":        "extra_q.type [type]",
		"description": "extra.description [description]",
		"tag":         "extra_q.tag tag",
	}

	r.GET("/q", func(ctx *gin.Context) {
		var query struct {
			Q       string  `form:"q"`
			Select  string  `form:"select"`
			Sort    string  `form:"sort"`
			Page    *string `form:"page"`
			PerPage *string `form:"perPage"`
		}

		if e := ctx.BindQuery(&query); e != nil {
			ctx.AbortWithError(400, e)
			return
		}

		if query.Sort == "" {
			query.Sort = "-updatedAt"
		}

		sorter := query.Sort
		sortDirection := ""

		if string((query.Sort)[0]) == "-" {
			sorter = string((query.Sort)[1:])
			sortDirection = " desc"
		}

		sorter = map[string]string{
			"updatedAt": "updated_at",
		}[sorter]

		if sorter == "" {
			sorter = "updated_at"
		}

		page := 1
		{
			a, e := strconv.Atoi(*query.Page)
			if e != nil {
				ctx.AbortWithError(400, e)
				return
			}
			page = a
		}

		perPage := 10
		{
			a, e := strconv.Atoi(*query.PerPage)
			if e != nil {
				ctx.AbortWithError(400, e)
				return
			}
			perPage = a
		}

		sel := []string{}

		for _, s := range strings.Split(query.Select, ",") {
			k := sMap[s]
			if k != "" {
				sel = append(sel, k)
			}
		}

		if len(sel) == 0 {
			ctx.AbortWithError(400, fmt.Errorf("not enough select"))
			return
		}

		q := resource.DB.Current.Model(&db.Extra{}).Joins("LEFT JOIN extra_q ON extra_q.id = extra.id")

		if query.Q != "" {
			q = q.Where(`extra.id IN (
				SELECT id FROM extra_q WHERE extra_q MATCH ?
			)`, query.Q)
		}

		q = q.Group("extra.id")

		var count int64

		if r := q.Count(&count); r.Error != nil {
			panic(r.Error)
		}

		out := struct {
			Result []map[string]interface{} `json:"result"`
			Count  int64                    `json:"count"`
		}{
			Result: make([]map[string]interface{}, 0),
			Count:  count,
		}

		if r := q.
			Select(strings.Join(sel, ",")).
			Order(sorter + sortDirection).
			Limit(perPage).
			Offset((page - 1) * perPage).
			Find(&out.Result); r.Error != nil {
			panic(r.Error)
		}

		ctx.JSON(200, out)
	})

	r.GET("/", func(ctx *gin.Context) {
		var query struct {
			Entry  string `form:"entry" binding:"required"`
			Select string `form:"select"`
		}

		if e := ctx.BindQuery(&query); e != nil {
			ctx.AbortWithError(400, e)
			return
		}

		sel := []string{}

		for _, s := range strings.Split(query.Select, ",") {
			k := sMap[s]
			if k != "" {
				sel = append(sel, k)
			}
		}

		if len(sel) == 0 {
			ctx.AbortWithError(400, fmt.Errorf("not enough select"))
			return
		}

		out := map[string]interface{}{}

		if r := resource.DB.Current.
			Model(&db.Extra{}).
			Joins("LEFT JOIN extra_q ON extra_q.id = extra.id").
			Select(strings.Join(sel, ",")).
			Where("extra.chinese = ?", query.Entry).
			Group("extra.id").
			First(&out); r.Error != nil {
			panic(r.Error)
		}

		ctx.JSON(200, out)
	})

	r.PUT("/", func(ctx *gin.Context) {
		var body struct {
			Chinese     string `json:"chinese" binding:"required"`
			Pinyin      string `json:"pinyin" binding:"required"`
			English     string `json:"english" binding:"required"`
			Type        string `json:"type"`
			Description string `json:"description"`
			Tag         string `json:"tag"`
			Forced      bool   `json:"forced"`
		}

		if e := ctx.BindJSON(&body); e != nil {
			ctx.AbortWithError(400, e)
			return
		}

		checkVocab := func() bool {
			var simplified string

			if r := resource.Zh.Current.Raw(`
			SELECT simplified
			FROM vocab
			WHERE simplified = ? OR traditional = ?
			LIMIT 1
			`, body.Chinese, body.Chinese).First(&simplified); r.Error != nil {
				if errors.Is(r.Error, gorm.ErrRecordNotFound) {
					return false
				}
				panic(r.Error)
			}

			ctx.JSON(200, gin.H{
				"existing": gin.H{
					"type":  "vocab",
					"entry": simplified,
				},
			})

			return true
		}

		checkHanzi := func() bool {
			var entry string
			if r := resource.Zh.Current.Raw(`
			SELECT [entry]
			FROM token
			WHERE [entry] = ? AND english IS NOT NULL
			LIMIT 1
			`, body.Chinese).First(&entry); r.Error != nil {
				if errors.Is(r.Error, gorm.ErrRecordNotFound) {
					return false
				}

				panic(r.Error)
			}

			ctx.JSON(200, gin.H{
				"existing": gin.H{
					"type":  "hanzi",
					"entry": entry,
				},
			})

			return true
		}

		checkSentence := func() bool {
			var chinese string
			if r := resource.Zh.Current.Raw(`
			SELECT chinese
			FROM sentence
			WHERE chinese = ?
			LIMIT 1
			`, body.Chinese).First(&chinese); r.Error != nil {
				if errors.Is(r.Error, gorm.ErrRecordNotFound) {
					return false
				}

				panic(r.Error)
			}

			ctx.JSON(200, gin.H{
				"existing": gin.H{
					"type":  "sentence",
					"entry": chinese,
				},
			})

			return true
		}

		if !body.Forced {
			if checkVocab() {
				return
			}

			if len([]rune(body.Chinese)) == 1 {
				if checkHanzi() {
					return
				}
			} else {
				if checkSentence() {
					return
				}
			}
		}

		it := db.Extra{
			Chinese:     body.Chinese,
			Pinyin:      body.Pinyin,
			English:     body.English,
			Type:        body.Type,
			Description: body.Description,
			Tag:         body.Tag,
		}

		e := resource.DB.Current.Transaction(func(tx *gorm.DB) error {
			if r := tx.Create(&it); r.Error != nil {
				return r.Error
			}

			return nil
		})

		if e != nil {
			panic(e)
		}

		ctx.JSON(201, gin.H{
			"id": it.ID,
		})
	})

	r.PATCH("/", func(ctx *gin.Context) {
		id := ctx.Query("id")
		if id == "" {
			ctx.AbortWithError(400, fmt.Errorf("id to update not specified"))
			return
		}

		var body struct {
			Chinese     string `json:"chinese" binding:"required"`
			Pinyin      string `json:"pinyin" binding:"required"`
			English     string `json:"english" binding:"required"`
			Type        string `json:"type"`
			Description string `json:"description"`
			Tag         string `json:"tag"`
		}

		if e := ctx.BindJSON(&body); e != nil {
			ctx.AbortWithError(400, e)
			return
		}

		u := db.Extra{
			ID:          id,
			Chinese:     body.Chinese,
			Pinyin:      body.Pinyin,
			English:     body.English,
			Type:        body.Type,
			Description: body.Description,
			Tag:         body.Tag,
		}

		if u.Description == "" {
			u.Description = " "
		}

		if u.Tag == "" {
			u.Tag = " "
		}

		e := resource.DB.Current.Transaction(func(tx *gorm.DB) error {
			if r := u.Update(tx); r != nil {
				return r
			}

			return nil
		})

		if e != nil {
			panic(e)
		}

		ctx.JSON(201, gin.H{
			"result": "updated",
		})
	})

	r.DELETE("/", func(ctx *gin.Context) {
		id := ctx.Query("id")
		if id == "" {
			ctx.AbortWithError(400, fmt.Errorf("id to update not specified"))
			return
		}

		e := resource.DB.Current.Transaction(func(tx *gorm.DB) error {
			ex := db.Extra{
				ID: id,
			}
			if e := ex.Delete(tx); e != nil {
				return e
			}

			return nil
		})

		if e != nil {
			panic(e)
		}

		ctx.JSON(201, gin.H{
			"result": "deleted",
		})
	})
}
