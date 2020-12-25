package api

import (
	"fmt"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/zhquiz/go-server/server/db"
)

type tQuizRouter struct {
	Router *gin.RouterGroup
}

func (r tQuizRouter) init() {
	r.getMatchMany()
}

func (r tQuizRouter) getMatchMany() {
	r.Router.POST("/get", func(ctx *gin.Context) {
		session := sessions.Default(ctx)
		userID := session.Get("userID").(string)
		if userID == "" {
			ctx.AbortWithStatus(401)
		}

		var body struct {
			IDs     []string
			Entries []string
			Type    string
			Select  []string `binding:"required,min=1"`
		}

		if e := ctx.BindJSON(&body); e != nil {
			panic(e)
		}

		sel := []string{}
		sMap := map[string]string{
			"id":        "ID",
			"tag":       "Tag",
			"entry":     "Entry",
			"type":      "Type",
			"direction": "Direction",
			"front":     "Front",
			"back":      "Back",
			"mnemonic":  "Mnemonic",
		}

		for _, s := range body.Select {
			k := sMap[s]
			if k != "" {
				sel = append(sel, "["+k+"]")
			}
		}

		if len(sel) == 0 {
			ctx.AbortWithError(400, fmt.Errorf("not enough select"))
		}

		var out []gin.H

		if len(body.IDs) > 0 {
			if r := resource.DB.Current.Model(&db.Quiz{}).
				Select(sel).
				Where("UserID = ? AND ID IN ?", userID, body.IDs).
				Find(&out); r.Error != nil {
				panic(r.Error)
			}
		} else if len(body.Entries) > 0 && body.Type != "" {
			if r := resource.DB.Current.Model(&db.Quiz{}).
				Select(sel).
				Where("UserID = ? AND [Type] = ? AND Entry IN ?", userID, body.Type, body.Entries).
				Find(&out); r.Error != nil {
				panic(r.Error)
			}
		} else {
			ctx.AbortWithError(400, fmt.Errorf("either IDs or Entries must be specified"))
		}

		ctx.JSON(200, gin.H{
			"result": out,
		})
	})
}
