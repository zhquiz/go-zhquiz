package api

import (
	"fmt"
	"time"

	"github.com/gin-contrib/cache"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gopkg.in/sakura-internet/go-rison.v3"
)

type tExtraRouter struct {
	Router *gin.RouterGroup
}

func (r tExtraRouter) init() {
	r.getQ()
}

func (r tExtraRouter) getQ() {
	r.Router.GET("/q", cache.CachePage(store, time.Hour, func(ctx *gin.Context) {
		session := sessions.Default(ctx)
		userID := session.Get("userID").(string)
		if userID == "" {
			ctx.AbortWithStatus(401)
		}

		var query struct {
			RS string `form:"_" binding:"required"`
		}

		var rs struct {
			Select  []string `validate:"min=1"`
			Sort    *string
			Page    *int
			PerPage *int `json:"perPage"`
		}

		if e := ctx.ShouldBindQuery(&query); e != nil {
			panic(e)
		}

		if e := rison.Unmarshal([]byte(query.RS), &rs, rison.Rison); e != nil {
			panic(e)
		}

		if e := validate.Struct(&rs); e != nil {
			panic(e)
		}

		if rs.Sort == nil {
			*rs.Sort = "-updatedAt"
		}

		if string((*rs.Sort)[0]) == "-" {
			*rs.Sort = string((*rs.Sort)[1:]) + " desc"
		}

		if rs.Page == nil {
			*rs.Page = 1
		}

		if rs.PerPage == nil {
			*rs.PerPage = 10
		}

		sel := []string{}
		sMap := map[string]string{
			"chinese": "Chinese",
			"pinyin":  "Pinyin",
			"english": "English",
		}

		for _, s := range rs.Select {
			k := sMap[s]
			if k != "" {
				sel = append(sel, k)
			}
		}

		if len(sel) == 0 {
			ctx.AbortWithError(400, fmt.Errorf("not enough select parameters"))
		}

		var out struct {
			result []gin.H
			count  int
		}

		resource.DB.Current.
			Select("COUNT(*) AS [Count]").
			Where("userID = ?", userID).
			Find(&out)

		resource.DB.Current.
			Select(sel).
			Order(*rs.Sort).
			Limit(*rs.PerPage).
			Offset((*rs.Page-1)**rs.PerPage).
			Where("userID = ?", userID).
			Find(&out.result)

		ctx.JSON(200, out)
	}))
}
