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

	r.GET("/q", func(ctx *gin.Context) {
		userID := getUserID(ctx)
		if userID == "" {
			ctx.AbortWithStatus(401)
			return
		}

		var query struct {
			Q       string  `form:"q"`
			Select  string  `form:"select"`
			Sort    string  `form:"sort"`
			Page    *string `form:"page"`
			PerPage *string `form:"perPage"`
		}

		if e := ctx.ShouldBindQuery(&query); e != nil {
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

		if query.Page == nil {
			a, e := strconv.Atoi(*query.Page)
			if e != nil {
				ctx.AbortWithError(400, e)
				return
			}
			page = a
		}

		perPage := 10

		if query.PerPage == nil {
			a, e := strconv.Atoi(*query.PerPage)
			if e != nil {
				ctx.AbortWithError(400, e)
				return
			}
			perPage = a
		}

		sel := []string{}
		sMap := map[string]string{
			"id":          "ID",
			"chinese":     "Chinese",
			"pinyin":      "Pinyin",
			"english":     "English",
			"type":        "Type",
			"description": "Description",
			"tag":         "Tag",
		}

		for _, s := range strings.Split(query.Select, ",") {
			k := sMap[s]
			if k != "" {
				sel = append(sel, k)
			}
		}

		if len(sel) == 0 {
			sel = []string{"ID", "Chinese", "Pinyin", "English"}
		}

		q := resource.DB.Current.Model(&db.Extra{}).Where("user_id = ?", userID)

		if query.Q != "" {
			var ids []string
			var result []struct {
				ID string
			}

			// Don't error if malformed Q
			resource.DB.Current.Raw(`
			SELECT id FROM extra_q WHERE extra_q MATCH ?
			`, query.Q).Find(&result)

			for _, r := range result {
				ids = append(ids, r.ID)
			}

			if len(ids) > 0 {
				q = q.Where("id IN ?", ids)
			} else {
				q = q.Where("FALSE")
			}
		}

		var getCount struct {
			Count int
		}

		if r := q.Select("COUNT(ID) AS [Count]").Scan(&getCount); r.Error != nil {
			panic(r.Error)
		}

		out := struct {
			Result []db.Extra `json:"result"`
			Count  int        `json:"count"`
		}{
			Result: make([]db.Extra, 0),
			Count:  getCount.Count,
		}

		if r := q.
			Select(sel).
			Order(sorter + sortDirection).
			Limit(perPage).
			Offset((page - 1) * perPage).
			Scan(&out.Result); r.Error != nil {
			panic(r.Error)
		}

		ctx.JSON(200, out)
	})

	r.GET("/", func(ctx *gin.Context) {
		userID := getUserID(ctx)
		if userID == "" {
			ctx.AbortWithStatus(401)
			return
		}

		var query struct {
			Entry  string `form:"entry" binding:"required"`
			Select string `form:"select"`
		}

		if e := ctx.ShouldBindQuery(&query); e != nil {
			ctx.AbortWithError(400, e)
			return
		}

		sel := []string{}
		sMap := map[string]string{
			"chinese": "Chinese",
			"pinyin":  "Pinyin",
			"english": "English",
		}

		for _, s := range strings.Split(query.Select, ",") {
			k := sMap[s]
			if k != "" {
				sel = append(sel, k)
			}
		}

		if len(sel) == 0 {
			sel = []string{"Chinese", "Pinyin", "English"}
		}

		var out db.Extra

		if r := resource.DB.Current.
			Select(sel).
			Where("user_id = ? AND chinese = ?", userID, query.Entry).
			First(&out); r.Error != nil {
			panic(r.Error)
		}

		ctx.JSON(200, out)
	})

	r.PUT("/", func(ctx *gin.Context) {
		userID := getUserID(ctx)
		if userID == "" {
			ctx.AbortWithStatus(401)
			return
		}

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
			FROM sentence_q
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
			UserID:      userID,
		}

		if r := resource.DB.Current.Create(&it); r.Error != nil {
			panic(r.Error)
		}

		ctx.JSON(201, gin.H{
			"id": it.ID,
		})
	})

	r.PATCH("/", func(ctx *gin.Context) {
		userID := getUserID(ctx)
		if userID == "" {
			ctx.AbortWithStatus(401)
			return
		}

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

		if e := ctx.ShouldBindJSON(&body); e != nil {
			ctx.AbortWithError(400, e)
			return
		}

		update := db.Extra{
			Chinese:     body.Chinese,
			Pinyin:      body.Pinyin,
			English:     body.English,
			Type:        body.Type,
			Description: body.Description,
			Tag:         body.Tag,
		}

		if r := resource.DB.Current.
			Model(&db.Extra{}).
			Where("user_id = ? AND id = ?", userID, id).
			Updates(update); r.Error != nil {
			panic(r.Error)
		}

		ctx.JSON(201, gin.H{
			"result": "updated",
		})
	})

	r.DELETE("/", func(ctx *gin.Context) {
		userID := getUserID(ctx)
		if userID == "" {
			ctx.AbortWithStatus(401)
			return
		}

		id := ctx.Query("id")
		if id == "" {
			ctx.AbortWithError(400, fmt.Errorf("id to update not specified"))
			return
		}

		if r := resource.DB.Current.
			Where("user_id = ? AND id = ?", userID, id).
			Delete(&db.Extra{}); r.Error != nil {
			panic(r.Error)
		}

		ctx.JSON(201, gin.H{
			"result": "deleted",
		})
	})
}
