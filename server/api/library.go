package api

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v2"
)

func routerLibrary(apiRouter *gin.RouterGroup) {
	r := apiRouter.Group("/library")

	r.GET("/q", func(ctx *gin.Context) {
		var query struct {
			Q       string `form:"q"`
			Page    string `form:"page" binding:"required"`
			PerPage string `form:"perPage" binding:"required"`
		}

		if e := ctx.ShouldBindQuery(&query); e != nil {
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
			Title   string   `json:"title"`
			Entries []string `json:"entries"`
		}
		result := make([]Result, 0)

		type lib struct {
			Title   string
			Entries string
		}
		var preresult []lib
		count := 0

		if query.Q != "" {
			if r := resource.DB.Current.Raw(fmt.Sprintf(`
			SELECT Title, Entries FROM library WHERE library MATCH ?
			ORDER BY rank
			LIMIT %d OFFSET %d
			`, perPage, (page-1)*perPage), query.Q).Find(&preresult); r.Error != nil {
				panic(r.Error)
			}

			if err := resource.DB.Current.Raw(`
			SELECT COUNT(*) FROM library WHERE library MATCH ?
			`, query.Q).Row().Scan(&count); err != nil {
				panic(err)
			}
		} else {
			if r := resource.DB.Current.Raw(fmt.Sprintf(`
			SELECT Title, Entries FROM library
			LIMIT %d OFFSET %d
			`, perPage, (page-1)*perPage)).Find(&preresult); r.Error != nil {
				panic(r.Error)
			}

			if err := resource.DB.Current.Raw(`
			SELECT COUNT(*) FROM library
			`).Row().Scan(&count); err != nil {
				panic(err)
			}
		}

		for _, p := range preresult {
			entries := make([]string, 0)

			if err := yaml.Unmarshal([]byte(p.Entries), &entries); err != nil {
				panic(err)
			}

			result = append(result, Result{
				Title:   p.Title,
				Entries: entries,
			})
		}

		ctx.JSON(200, gin.H{
			"result": result,
			"count":  count,
		})
	})
}
