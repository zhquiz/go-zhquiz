package api

import (
	"database/sql"
	"fmt"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/zhquiz/go-server/server/db"
	"gopkg.in/sakura-internet/go-rison.v3"
)

type tExtraRouter struct {
	Router *gin.RouterGroup
}

func (r tExtraRouter) init() {
	r.getQ()
	r.getMatch()
	r.doCreate()
	r.doUpdate()
	r.doDelete()
}

func (r tExtraRouter) getQ() {
	r.Router.GET("/q", func(ctx *gin.Context) {
		session := sessions.Default(ctx)
		userID := session.Get("userID").(string)
		if userID == "" {
			ctx.AbortWithStatus(401)
		}

		var query struct {
			RS string `form:"_"`
		}

		var rs struct {
			Select  []string `json:"select"`
			Sort    *string  `json:"sort"`
			Page    *int     `json:"page"`
			PerPage *int     `json:"perPage"`
		}

		if e := ctx.ShouldBindQuery(&query); e != nil {
			ctx.AbortWithError(400, e)
		}

		if e := rison.Unmarshal([]byte(query.RS), &rs, rison.Rison); e != nil {
			ctx.AbortWithError(400, e)
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
			"id":      "ID",
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
			sel = []string{"Chinese", "Pinyin", "English"}
		}

		var out struct {
			Result []gin.H `json:"result"`
			Count  int     `json:"count"`
		}

		if r := resource.DB.Current.
			Model(&db.Extra{}).
			Select("COUNT(ID) AS [Count]").
			Where("userID = ?", userID).
			Find(&out); r.Error != nil {
			panic(r.Error)
		}

		if r := resource.DB.Current.
			Model(&db.Extra{}).
			Select(sel).
			Order(*rs.Sort).
			Limit(*rs.PerPage).
			Offset((*rs.Page-1)**rs.PerPage).
			Where("userID = ?", userID).
			Find(&out.Result); r.Error != nil {
			panic(r.Error)
		}

		ctx.JSON(200, out)
	})
}

func (r tExtraRouter) getMatch() {
	r.Router.GET("/", func(ctx *gin.Context) {
		session := sessions.Default(ctx)
		userID := session.Get("userID").(string)
		if userID == "" {
			ctx.AbortWithStatus(401)
		}

		var query struct {
			Entry string `form:"entry" binding:"required"`
			RS    string `form:"_"`
		}

		var rs struct {
			Select  []string `json:"select"`
			Sort    *string  `json:"sort"`
			Page    *int     `json:"page"`
			PerPage *int     `json:"perPage"`
		}

		if e := ctx.ShouldBindQuery(&query); e != nil {
			ctx.AbortWithError(400, e)
		}

		if e := rison.Unmarshal([]byte(query.RS), &rs, rison.Rison); e != nil {
			ctx.AbortWithError(400, e)
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
			sel = []string{"Chinese", "Pinyin", "English"}
		}

		var out gin.H

		if r := resource.DB.Current.
			Model(&db.Extra{}).
			Select(sel).
			Where("userID = ?", userID).
			First(&out); r.Error != nil {
			panic(r.Error)
		}

		ctx.JSON(200, out)
	})
}

func (r tExtraRouter) doCreate() {
	r.Router.PUT("/", func(ctx *gin.Context) {
		session := sessions.Default(ctx)
		userID := session.Get("userID").(string)
		if userID == "" {
			ctx.AbortWithStatus(401)
		}

		var body struct {
			Chinese string `json:"chinese" binding:"required"`
			Pinyin  string `json:"pinyin" binding:"required"`
			English string `json:"english" binding:"required"`
		}

		if e := ctx.BindJSON(&body); e != nil {
			ctx.AbortWithError(400, e)
		}

		checkVocab := func() bool {
			stmt, e := resource.Zh.Current.Prepare(`
			SELECT simplified FROM cedict
			WHERE simplified = ? OR traditional = ?
			`)

			if e != nil {
				panic(e)
			}

			r := stmt.QueryRow(body.Chinese, body.Chinese)

			if e := r.Err(); e != nil {
				panic(e)
			}

			var entry string
			if e := r.Scan(entry); e == sql.ErrNoRows {
				return false
			}

			ctx.JSON(200, gin.H{
				"existing": gin.H{
					"type":  "vocab",
					"entry": entry,
				},
			})

			return true
		}

		checkHanzi := func() bool {
			stmt, e := resource.Zh.Current.Prepare(`
			SELECT entry FROM token
			WHERE entry = ? AND english IS NOT NULL
			`)

			if e != nil {
				panic(e)
			}

			r := stmt.QueryRow(body.Chinese)

			if e := r.Err(); e != nil {
				panic(e)
			}

			var entry string
			if e := r.Scan(entry); e == sql.ErrNoRows {
				return false
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
			stmt, e := resource.Zh.Current.Prepare(`
			SELECT chinese FROM sentence
			WHERE chinese = ?
			`)

			if e != nil {
				panic(e)
			}

			r := stmt.QueryRow(body.Chinese)

			if e := r.Err(); e != nil {
				panic(e)
			}

			var entry string
			if e := r.Scan(entry); e == sql.ErrNoRows {
				return false
			}

			ctx.JSON(200, gin.H{
				"existing": gin.H{
					"type":  "sentence",
					"entry": entry,
				},
			})

			return true
		}

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

		it := db.Extra{
			Chinese: body.Chinese,
			Pinyin:  body.Pinyin,
			English: body.English,
			UserID:  userID,
		}

		if r := resource.DB.Current.Create(&it); r.Error != nil {
			panic(r.Error)
		}

		ctx.JSON(201, gin.H{
			"id": it.ID,
		})
	})
}

func (r tExtraRouter) doUpdate() {
	r.Router.PUT("/", func(ctx *gin.Context) {
		session := sessions.Default(ctx)
		userID := session.Get("userID").(string)
		if userID == "" {
			ctx.AbortWithStatus(401)
		}

		id := ctx.Query("id")
		if id == "" {
			ctx.AbortWithError(400, fmt.Errorf("id to update not specified"))
		}

		var body struct {
			Chinese string `json:"chinese" binding:"required"`
			Pinyin  string `json:"pinyin" binding:"required"`
			English string `json:"english" binding:"required"`
		}

		if e := ctx.BindJSON(&body); e != nil {
			ctx.AbortWithError(400, e)
		}

		if r := resource.DB.Current.
			Model(&db.Extra{}).
			Where("UserID = ? AND ID = ?", userID, id).
			Updates(body); r.Error != nil {
			panic(r.Error)
		}

		ctx.JSON(201, gin.H{
			"result": "updated",
		})
	})
}

func (r tExtraRouter) doDelete() {
	r.Router.PUT("/", func(ctx *gin.Context) {
		session := sessions.Default(ctx)
		userID := session.Get("userID").(string)
		if userID == "" {
			ctx.AbortWithStatus(401)
		}

		id := ctx.Query("id")
		if id == "" {
			ctx.AbortWithError(400, fmt.Errorf("id to update not specified"))
		}

		if r := resource.DB.Current.Unscoped().
			Where("UserID = ? AND ID = ?", userID, id).
			Delete(&db.Extra{}); r.Error != nil {
			panic(r.Error)
		}

		ctx.JSON(201, gin.H{
			"result": "deleted",
		})
	})
}
