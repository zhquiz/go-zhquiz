package api

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zhquiz/go-zhquiz/server/db"
	"gorm.io/gorm"
)

func routerLibrary(apiRouter *gin.RouterGroup) {
	r := apiRouter.Group("/library")

	r.GET("/q", func(ctx *gin.Context) {
		var query struct {
			Q       string `form:"q"`
			Page    string `form:"page" binding:"required"`
			PerPage string `form:"perPage" binding:"required"`
		}

		if e := ctx.BindQuery(&query); e != nil {
			ctx.AbortWithError(400, e)
			return
		}

		page, err := strconv.Atoi(query.Page)
		if err != nil {
			ctx.AbortWithError(400, err)
			return
		}

		perPage, err := strconv.Atoi(query.PerPage)
		if err != nil {
			ctx.AbortWithError(400, err)
			return
		}

		type Result struct {
			ID      string   `json:"id"`
			Title   string   `json:"title"`
			Entries []string `json:"entries"`
		}
		result := make([]Result, 0)

		type lib struct {
			ID    string
			Title string
			Entry string
		}
		var preresult []lib
		count := 0

		if query.Q != "" {
			if r := resource.DB.Current.Raw(fmt.Sprintf(`
			SELECT ID, Title, Entry FROM library_q WHERE library_q MATCH ?
			ORDER BY rank
			LIMIT %d OFFSET %d
			`, perPage, (page-1)*perPage), query.Q).Find(&preresult); r.Error != nil {
				panic(r.Error)
			}

			if err := resource.DB.Current.Raw(`
			SELECT COUNT(*) FROM library_q WHERE library_q MATCH ?
			`, query.Q).Row().Scan(&count); err != nil {
				panic(err)
			}
		} else {
			if r := resource.DB.Current.Raw(fmt.Sprintf(`
			SELECT library.id ID, library.Title Title, Entry FROM library_q
			LEFT JOIN library ON library.id = library_q.id
			ORDER BY updated_at DESC
			LIMIT %d OFFSET %d
			`, perPage, (page-1)*perPage)).Find(&preresult); r.Error != nil {
				panic(r.Error)
			}

			if err := resource.DB.Current.Raw(`
			SELECT COUNT(*) FROM library_q
			`).Row().Scan(&count); err != nil {
				panic(err)
			}
		}

		for _, p := range preresult {
			entries := strings.Split(p.Entry, " ")

			if p.ID[0] == ' ' {
				p.ID = ""
			}

			result = append(result, Result{
				ID:      p.ID,
				Title:   p.Title,
				Entries: entries,
			})
		}

		ctx.JSON(200, gin.H{
			"result": result,
			"count":  count,
		})
	})

	r.PUT("/", func(ctx *gin.Context) {
		var body struct {
			Title       string   `json:"title" binding:"required"`
			Entries     []string `json:"entries" binding:"required,min=1"`
			Description string   `json:"description"`
			Tag         string   `json:"tag"`
		}

		if e := ctx.BindJSON(&body); e != nil {
			ctx.AbortWithError(400, e)
			return
		}

		it := db.Library{
			Title:       body.Title,
			Entries:     body.Entries,
			Description: body.Description,
			Tag:         body.Tag,
		}

		e := resource.DB.Current.Transaction(func(tx *gorm.DB) error {
			if e := it.Create(tx); e != nil {
				return e
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
			Title       string   `json:"title" binding:"required"`
			Entries     []string `json:"entries" binding:"required,min=1"`
			Description string   `json:"description"`
			Tag         string   `json:"tag"`
		}

		if e := ctx.BindJSON(&body); e != nil {
			ctx.AbortWithError(400, e)
			return
		}

		u := db.Library{
			ID:          id,
			Title:       body.Title,
			Entries:     body.Entries,
			Description: body.Description,
			Tag:         body.Tag,
		}

		e := resource.DB.Current.Transaction(func(tx *gorm.DB) error {
			if e := u.Update(tx); e != nil {
				return e
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
			lib := db.Library{
				ID: id,
			}

			if e := lib.Delete(tx); e != nil {
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
