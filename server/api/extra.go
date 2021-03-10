package api

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zhquiz/go-zhquiz/server/db"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type rExtra struct {
	Base *gin.RouterGroup
}

func (r rExtra) init() {
	router := r.Base.Group("/extra")

	router.GET("/q", r.getQuery)
}

// @Accept json
// @Produce json
// @Param q query string false "text to search"
// @Param sort query string false "column to sort by"
// @Param page query int true "page number"
// @Param limit query int true "number of contents per page"
// @Success 200 {object} ExtraQueryResponse
// @Router /extra/q [get]
func (rExtra) getQuery(ctx *gin.Context) {
	var query struct {
		Q     string `form:"q"`
		Sort  string `form:"sort"`
		Page  string `form:"page"`
		Limit string `form:"limit"`
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
		sortDirection = "desc"
	}

	sorter = map[string]string{
		"updatedAt": "updated_at",
	}[sorter]

	if sorter == "" {
		sorter = "updated_at"
	}

	page := 1
	{
		a, e := strconv.Atoi(query.Page)
		if e != nil {
			ctx.AbortWithError(400, e)
			return
		}
		page = a
	}

	limit := 10
	{
		a, e := strconv.Atoi(query.Limit)
		if e != nil {
			ctx.AbortWithError(400, e)
			return
		}
		limit = a
	}

	q := resource.DB.Model(&db.Extra{}).Where("userID = ")

	var count int64

	if r := q.Count(&count); r.Error != nil {
		panic(r.Error)
	}

	out := ExtraQueryResponse{
		Result: make([]ExtraItem, 0),
		Count:  count,
	}

	var items []db.Extra

	if r := q.
		Order(sorter + " " + sortDirection).
		Limit(limit).
		Offset((page - 1) * limit).
		Find(&items); r.Error != nil {
		panic(r.Error)
	}

	for _, it := range items {
		out.Result = append(out.Result, ExtraItem{
			Entry:       it.Entry,
			Type:        it.Type,
			Reading:     it.Reading,
			English:     it.English,
			Description: it.Description,
		})
	}

	ctx.JSON(200, out)
}

// ExtraItem is outputtable Extra item format
type ExtraItem struct {
	Entry       string   `json:"entry"`
	Type        string   `json:"type"`
	Reading     string   `json:"reading"`
	English     string   `json:"english"`
	Description string   `json:"description"`
	Tag         []string `json:"tag"`
}

// ExtraQueryResponse -
type ExtraQueryResponse struct {
	Result []ExtraItem `json:"result"`
	Count  int64       `json:"count"`
}

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

		if r := resource.DB.
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

			if r := resource.Zh.Raw(`
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
			if r := resource.Zh.Raw(`
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
			if r := resource.Zh.Raw(`
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
			Entry:       body.Chinese,
			Reading:     body.Pinyin,
			English:     body.English,
			Type:        body.Type,
			Description: body.Description,
		}

		if strings.TrimSpace(body.Tag) != "" {
			resource.DB.Transaction(func(tx *gorm.DB) error {
				if r := tx.Delete(
					tx.
						Where("user_id = ?", 2, body.Type).
						Where("`type` = ?", body.Type).
						Where("entry = ?", body.Chinese),
				); r.Error != nil {
					return r.Error
				}

				tags := make([]db.Tag, 0)

				for _, t := range strings.Split(body.Tag, " ") {
					tags = append(tags, db.Tag{
						UserID: 2,
						Entry:  body.Chinese,
						Type:   body.Type,
						Name:   t,
					})
				}

				if r := tx.Clauses(clause.OnConflict{
					DoNothing: true,
				}).Create(&tags); r.Error != nil {
					return r.Error
				}

				return nil
			})
		}

		e := resource.DB.Transaction(func(tx *gorm.DB) error {
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
		id0 := ctx.Query("id")
		if id0 == "" {
			ctx.AbortWithError(400, fmt.Errorf("id to update not specified"))
			return
		}

		id, e := strconv.Atoi(id0)
		if e != nil {
			ctx.AbortWithError(400, fmt.Errorf("invalid id format"))
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

		resource.DB.Create(&db.Extra{
			ID:          id,
			Entry:       body.Chinese,
			Reading:     body.Pinyin,
			English:     body.English,
			Type:        body.Type,
			Description: body.Description,
		})

		if strings.TrimSpace(body.Tag) != "" {
			resource.DB.Transaction(func(tx *gorm.DB) error {
				if r := tx.Delete(
					tx.
						Where("user_id = ?", 2, body.Type).
						Where("`type` = ?", body.Type).
						Where("entry = ?", body.Chinese),
				); r.Error != nil {
					return r.Error
				}

				tags := make([]db.Tag, 0)

				for _, t := range strings.Split(body.Tag, " ") {
					tags = append(tags, db.Tag{
						UserID: 2,
						Entry:  body.Chinese,
						Type:   body.Type,
						Name:   t,
					})
				}

				if r := tx.Clauses(clause.OnConflict{
					DoNothing: true,
				}).Create(&tags); r.Error != nil {
					return r.Error
				}

				return nil
			})
		}

		ctx.JSON(201, gin.H{
			"result": "updated",
		})
	})

	r.DELETE("/", func(ctx *gin.Context) {
		id0 := ctx.Query("id")
		if id0 == "" {
			ctx.AbortWithError(400, fmt.Errorf("id to update not specified"))
			return
		}

		id, e := strconv.Atoi(id0)
		if e != nil {
			ctx.AbortWithError(400, fmt.Errorf("invalid id format"))
			return
		}

		e = resource.DB.Transaction(func(tx *gorm.DB) error {
			if r := tx.Delete(&db.Extra{
				ID: id,
			}); r.Error != nil {
				return r.Error
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
