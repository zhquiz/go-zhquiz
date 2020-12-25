package api

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/zhquiz/go-server/server/db"
	"github.com/zhquiz/go-server/server/util"
	"gopkg.in/sakura-internet/go-rison.v3"
)

type tQuizRouter struct {
	Router *gin.RouterGroup
}

func (r tQuizRouter) init() {
	r.getMatchMany()
	r.doMark()
	r.getAllTags()
	r.initQuiz()
}

func (r tQuizRouter) getMatchMany() {
	r.Router.POST("/get", func(ctx *gin.Context) {
		session := sessions.Default(ctx)
		userID := session.Get("userID").(string)
		if userID == "" {
			ctx.AbortWithStatus(401)
		}

		var body struct {
			IDs     []string `json:"ids"`
			Entries []string `json:"entries"`
			Type    string   `json:"type"`
			Select  []string `json:"select" binding:"required,min=1"`
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

func (r tQuizRouter) doMark() {
	r.Router.POST("/get", func(ctx *gin.Context) {
		session := sessions.Default(ctx)
		userID := session.Get("userID").(string)
		if userID == "" {
			ctx.AbortWithStatus(401)
		}

		var query struct {
			ID   string `form:"id"`
			Type string `form:"type" binding:"oneof=right wrong repeat"`
		}

		if e := ctx.ShouldBindQuery(&query); e != nil {
			ctx.AbortWithError(400, e)
		}

		var quiz db.Quiz
		if r := resource.DB.Current.
			Where("UserID = ? AND ID = ?", userID, query.ID).
			First(&quiz); r.Error != nil {
			panic(r.Error)
		}

		quiz.UpdateSRSLevel(map[string]int8{
			"right":  1,
			"wrong":  -1,
			"repeat": 0,
		}[query.Type])

		if r := resource.DB.Current.Save(&quiz); r.Error != nil {
			panic(r.Error)
		}

		ctx.JSON(201, gin.H{
			"result": "updated",
		})
	})
}

func (r tQuizRouter) getAllTags() {
	r.Router.GET("/allTags", func(ctx *gin.Context) {
		session := sessions.Default(ctx)
		userID := session.Get("userID").(string)
		if userID == "" {
			ctx.AbortWithStatus(401)
		}

		var tagEls []struct {
			Name string
		}

		if r := resource.DB.Current.Model(&db.Quiz{}).
			Select("tag.Name").
			Joins("JOIN quiz_tag ON quiz_tag.quiz_id = quiz.id").
			Joins("JOIN tag ON quiz_tag.tag_id = tag.id").
			Group("tag.Name").
			Scan(&tagEls); r.Error != nil {
			panic(r.Error)
		}

		var result []string
		for _, t := range tagEls {
			result = append(result, t.Name)
		}

		ctx.JSON(200, gin.H{
			"result": result,
		})
	})
}

func (r tQuizRouter) initQuiz() {
	r.Router.GET("/init", func(ctx *gin.Context) {
		session := sessions.Default(ctx)
		userID := session.Get("userID").(string)
		if userID == "" {
			ctx.AbortWithStatus(401)
		}

		var query struct {
			RS string `form:"_"`
		}

		var rs struct {
			Type      []string `json:"type" validate:"required,min=1"`
			Stage     []string `json:"stage" validate:"required,min=1"`
			Direction []string `json:"direction" validate:"required,min=1"`
			IsDue     bool     `json:"isDue" validate:"required"`
			Tag       []string `json:"tag" validate:"required"`
		}

		if e := ctx.ShouldBindQuery(&query); e != nil {
			ctx.AbortWithError(400, e)
		}

		if e := rison.Unmarshal([]byte(query.RS), &rs, rison.Rison); e != nil {
			ctx.AbortWithError(400, e)
		}

		if e := validate.Struct(&rs); e != nil {
			ctx.AbortWithError(400, e)
		}

		// No need to await
		go func() {
			var user db.User
			if r := resource.DB.Current.First(&user, userID); r.Error != nil {
				panic(r.Error)
			}

			user.Meta.Quiz.Direction = rs.Direction
			user.Meta.Quiz.Stage = rs.Stage
			user.Meta.Quiz.Type = rs.Type
			user.Meta.Quiz.IsDue = rs.IsDue

			if r := resource.DB.Current.Save(&user); r.Error != nil {
				panic(r.Error)
			}
		}()

		var orCond []string

		stageSet := util.MakeSet(rs.Stage)
		if stageSet["new"] {
			orCond = append(orCond, "SRSLevel IS NULL")
		}

		if stageSet["leech"] {
			orCond = append(orCond, "WrongStreak >= 3")
		}

		if stageSet["learning"] {
			orCond = append(orCond, "SRSLevel < 3")
		}

		if stageSet["graduated"] {
			orCond = append(orCond, "SRSLevel >= 3")
		}

		q := resource.DB.Current.Where("UserID = ? AND [Type] IN ? AND Direction IN ?", userID, rs.Type, rs.Direction)

		if len(rs.Tag) > 0 {
			q = q.Where("Tag IN ?", rs.Tag)
		}

		if len(orCond) > 0 {
			q = q.Where(strings.Join(orCond, " OR "))
		}

		var quizzes []db.Quiz

		if r := q.Find(&quizzes); r.Error != nil {
			panic(r.Error)
		}

		var quiz []quizInitOutput
		var upcoming []quizInitOutput

		if rs.IsDue {
			now := time.Now()

			for _, it := range quizzes {
				if it.NextReview == nil || (*it.NextReview).Before(now) {
					quiz = append(quiz, quizInitOutput{
						NextReview:  it.NextReview,
						SRSLevel:    it.SRSLevel,
						WrongStreak: it.WrongStreak,
						ID:          it.ID,
					})
				} else {
					upcoming = append(upcoming, quizInitOutput{
						NextReview: it.NextReview,
						ID:         it.ID,
					})
				}
			}
		} else {
			for _, it := range quizzes {
				quiz = append(quiz, quizInitOutput{
					NextReview:  it.NextReview,
					SRSLevel:    it.SRSLevel,
					WrongStreak: it.WrongStreak,
					ID:          it.ID,
				})
			}
		}

		rand.Shuffle(len(quiz), func(i, j int) {
			quiz[i], quiz[j] = quiz[j], quiz[i]
		})

		sort.Sort(quizInitOutputList(upcoming))

		ctx.JSON(200, gin.H{
			"quiz":     quiz,
			"upcoming": upcoming,
		})
	})
}

type quizInitOutput struct {
	NextReview  *time.Time `json:"nextReview"`
	SRSLevel    *int8      `json:"srsLevel"`
	WrongStreak *uint      `json:"wrongStreak"`
	ID          string     `json:"id"`
}

type quizInitOutputList []quizInitOutput

func (ls quizInitOutputList) Len() int {
	return len(ls)
}

func (ls quizInitOutputList) Less(i, j int) bool {
	a, b := ls[i], ls[j]
	if a.NextReview == nil {
		return true
	}
	if b.NextReview == nil {
		return false
	}

	return a.NextReview.Before(*b.NextReview)
}

func (ls quizInitOutputList) Swap(i, j int) {
	ls[i], ls[j] = ls[j], ls[i]
}
