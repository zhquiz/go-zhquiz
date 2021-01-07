package api

import (
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zhquiz/go-zhquiz/server/db"
	"github.com/zhquiz/go-zhquiz/server/util"
	"github.com/zhquiz/go-zhquiz/server/zh"
	"gopkg.in/sakura-internet/go-rison.v3"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func routerQuiz(apiRouter *gin.RouterGroup) {
	r := apiRouter.Group("/quiz")

	r.GET("/many", func(ctx *gin.Context) {
		userID := getUserID(ctx)
		if userID == "" {
			ctx.AbortWithStatus(401)
			return
		}

		var query struct {
			IDs     string `form:"ids"`
			Entries string `form:"entries"`
			Type    string `form:"type" binding:"oneof=hanzi vocab sentence extra ''"`
			Select  string `form:"select"`
		}

		if e := ctx.ShouldBindQuery(&query); e != nil {
			ctx.AbortWithError(400, e)
		}

		sel := []string{}
		sMap := map[string]string{
			"id":        "ID",
			"entry":     "[Entry]",
			"type":      "[Type]",
			"direction": "Direction",
			"front":     "Front",
			"back":      "Back",
			"mnemonic":  "Mnemonic",
		}

		for _, s := range strings.Split(query.Select, ",") {
			k := sMap[s]
			if k != "" && k != "_" {
				sel = append(sel, k)
			}
		}

		if len(sel) == 0 {
			ctx.AbortWithError(400, fmt.Errorf("not enough select"))
			return
		}

		var ids []string
		if query.IDs != "" {
			ids = strings.Split(query.IDs, ",")
		}

		var entries []string
		if query.Entries != "" {
			entries = strings.Split(query.Entries, ",")
		}

		where := "user_id = @userID"
		cond := map[string]interface{}{
			"userID": userID,
		}

		if len(ids) > 0 {
			where = where + " AND id IN @ids"
			cond["ids"] = ids
		} else if len(entries) > 0 {
			where = where + " AND [entry] IN @entries"
			cond["entries"] = entries
		} else {
			ctx.AbortWithError(400, fmt.Errorf("either IDs or Entries must be specified"))
			return
		}

		if query.Type != "" {
			where = where + " AND [Type] = @type"
			cond["type"] = query.Type
		}

		var quizzes []db.Quiz

		clause := resource.DB.Current.Model(&db.Quiz{}).
			Select(sel).
			Where(where, cond)

		if r := clause.Find(&quizzes); r.Error != nil {
			panic(r.Error)
		}

		out := make([]gin.H, 0)
		getMap := map[string]func(q *db.Quiz) interface{}{
			"id":        func(q *db.Quiz) interface{} { return q.ID },
			"entry":     func(q *db.Quiz) interface{} { return q.Entry },
			"type":      func(q *db.Quiz) interface{} { return q.Type },
			"direction": func(q *db.Quiz) interface{} { return q.Direction },
		}

		for _, q := range quizzes {
			it := gin.H{}
			for _, s := range strings.Split(query.Select, ",") {
				k := getMap[s]
				if k != nil {
					it[s] = k(&q)
				}
			}

			out = append(out, it)
		}

		ctx.JSON(200, gin.H{
			"result": out,
		})
	})

	r.POST("/srsLevel", func(ctx *gin.Context) {
		userID := getUserID(ctx)
		if userID == "" {
			ctx.AbortWithStatus(401)
			return
		}

		var body struct {
			Entries []string `form:"entries" binding:"required,min=1"`
			Type    string   `form:"type" binding:"oneof=hanzi vocab sentence extra ''"`
		}

		if e := ctx.ShouldBindJSON(&body); e != nil {
			ctx.AbortWithError(400, e)
			return
		}

		where := "user_id = @userID AND [entry] IN @entries AND [Type] = @type"
		cond := map[string]interface{}{
			"userID":  userID,
			"entries": body.Entries,
			"type":    body.Type,
		}

		var quizzes []db.Quiz

		clause := resource.DB.Current.Model(&db.Quiz{}).
			Select("entry", "srs_level").
			Where(where, cond)

		if r := clause.Find(&quizzes); r.Error != nil {
			panic(r.Error)
		}

		out := make([]gin.H, 0)

		for _, q := range quizzes {
			out = append(out, gin.H{
				"entry":    q.Entry,
				"srsLevel": q.SRSLevel,
			})
		}

		ctx.JSON(200, gin.H{
			"result": out,
		})
	})

	r.PATCH("/mark", func(ctx *gin.Context) {
		userID := getUserID(ctx)
		if userID == "" {
			ctx.AbortWithStatus(401)
			return
		}

		var query struct {
			ID   string `form:"id" binding:"required"`
			Type string `form:"type" binding:"required,oneof=right wrong repeat"`
		}

		if e := ctx.ShouldBindQuery(&query); e != nil {
			ctx.AbortWithError(400, e)
			return
		}

		var quiz db.Quiz
		if r := resource.DB.Current.
			Where("user_id = ? AND id = ?", userID, query.ID).
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

	r.GET("/allTags", func(ctx *gin.Context) {
		userID := getUserID(ctx)
		if userID == "" {
			ctx.AbortWithStatus(401)
			return
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

	r.GET("/init", func(ctx *gin.Context) {
		userID := getUserID(ctx)
		if userID == "" {
			ctx.AbortWithStatus(401)
			return
		}

		var query struct {
			RS string `form:"_"`
		}

		var rs struct {
			Type      []string `json:"type" validate:"required,min=1"`
			Stage     []string `json:"stage" validate:"required,min=1"`
			Direction []string `json:"direction" validate:"required,min=1"`
			IsDue     bool     `json:"isDue"`
			Q         string   `json:"q"`
		}

		if e := ctx.ShouldBindQuery(&query); e != nil {
			ctx.AbortWithError(400, e)
			return
		}

		if e := rison.Unmarshal([]byte(query.RS), &rs, rison.Rison); e != nil {
			ctx.AbortWithError(400, e)
			return
		}

		if e := validate.Struct(&rs); e != nil {
			ctx.AbortWithError(400, e)
			return
		}

		// No need to await
		go func() {
			var user db.User
			if r := resource.DB.Current.Where("id = ?", userID).First(&user); r.Error != nil {
				panic(r.Error)
			}

			user.Meta.Settings.Quiz.Direction = rs.Direction
			user.Meta.Settings.Quiz.Stage = rs.Stage
			user.Meta.Settings.Quiz.Type = rs.Type
			user.Meta.Settings.Quiz.IsDue = rs.IsDue

			if r := resource.DB.Current.Save(&user); r.Error != nil {
				panic(r.Error)
			}
		}()

		q := resource.DB.Current.Model(&db.Quiz{})

		if rs.Q != "" {
			var sel []struct {
				ID string
			}
			// Ignore errors
			resource.DB.Current.Raw("SELECT id FROM quiz_q WHERE quiz_q MATCH ?", rs.Q).Find(&sel)

			var ids []string
			for _, s := range sel {
				ids = append(ids, s.ID)
			}

			if len(ids) > 0 {
				q = q.Where("id IN ?", ids)
			} else {
				q = q.Where("FALSE")
			}
		}

		var orCond []string

		stageSet := util.MakeSet(rs.Stage)
		if stageSet["new"] {
			orCond = append(orCond, "srs_level IS NULL")
		}

		if stageSet["leech"] {
			orCond = append(orCond, "wrong_streak >= 3")
		}

		if stageSet["learning"] {
			orCond = append(orCond, "srs_level < 3")
		}

		if stageSet["graduated"] {
			orCond = append(orCond, "srs_level >= 3")
		}

		if len(orCond) > 0 {
			q = q.Where(strings.Join(orCond, " OR "))
		}

		var quizzes []db.Quiz

		if r := q.
			Where("user_id = ? AND [type] IN ? AND direction IN ?", userID, rs.Type, rs.Direction).
			Find(&quizzes); r.Error != nil {
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

		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(quiz), func(i, j int) {
			quiz[i], quiz[j] = quiz[j], quiz[i]
		})

		sort.Sort(quizInitOutputList(upcoming))

		if len(quiz) == 0 {
			quiz = make([]quizInitOutput, 0)
		}

		if len(upcoming) == 0 {
			upcoming = make([]quizInitOutput, 0)
		}

		ctx.JSON(200, gin.H{
			"quiz":     quiz,
			"upcoming": upcoming,
		})
	})

	r.PUT("/", func(ctx *gin.Context) {
		userID := getUserID(ctx)
		if userID == "" {
			ctx.AbortWithStatus(401)
			return
		}

		var body struct {
			Entries []string `json:"entries" binding:"required,min=1"`
			Type    string   `json:"type" binding:"required,oneof=hanzi vocab sentence extra"`
		}
		if e := ctx.ShouldBindJSON(&body); e != nil {
			ctx.AbortWithError(400, e)
		}

		var existingQ []db.Quiz

		if r := resource.DB.Current.
			Where("user_id = ? AND entry IN ? AND type = ?", userID, body.Entries, body.Type).
			Find(&existingQ); r.Error != nil {
			panic(r.Error)
		}

		lookup := map[string]map[string]db.Quiz{}

		for _, it := range existingQ {
			if lookup[it.Entry] == nil {
				lookup[it.Entry] = map[string]db.Quiz{}
			}
			lookup[it.Entry][it.Direction] = it
		}

		type Result struct {
			IDs  []string `json:"ids"`
			Type string   `json:"type"`
		}
		result := make([]Result, 0)
		ids := make([]string, 0)

		var newQ []db.Quiz
		var newExtra []db.Extra

		for _, entry := range body.Entries {
			subresult := Result{
				IDs:  make([]string, 0),
				Type: "extra",
			}
			result = append(result, subresult)

			directions := []string{"se", "ec"}

			switch body.Type {
			case "vocab":
				var items []zh.Vocab
				if r := resource.Zh.Current.
					Where("simplified = ? OR traditional = ?", entry, entry).
					Find(&items); r.Error != nil {
					panic(r.Error)
				}

				if len(items) > 0 {
					for _, it := range items {
						if len(directions) < 3 && it.Traditional != "" {
							directions = append(directions, "te")
						}
					}
					subresult.Type = "vocab"
				} else {
					subresult.Type = "extra"
				}
			case "hanzi":
				if r := resource.Zh.Current.
					Where("entry = ? AND length(entry) = 1 AND english IS NOT NULL", entry).
					First(&zh.Token{}); r.Error != nil {
					if !errors.Is(r.Error, gorm.ErrRecordNotFound) {
						panic(r.Error)
					}
					subresult.Type = "extra"
				} else {
					subresult.Type = "hanzi"
				}
			case "sentence":
				if r := resource.Zh.Current.
					Where("chinese = ?", entry).
					First(&zh.Sentence{}); r.Error != nil {
					if !errors.Is(r.Error, gorm.ErrRecordNotFound) {
						panic(r.Error)
					}
					subresult.Type = "extra"
				} else {
					subresult.Type = "sentence"
				}
			}

			if subresult.Type == "extra" {
				newExtra = append(newExtra, db.Extra{
					ID:      NewULID(),
					UserID:  userID,
					Chinese: entry,
				})
			}

			lookupDir := lookup[entry]
			if lookupDir == nil {
				lookupDir = map[string]db.Quiz{}
			}

			for _, d := range directions {
				if lookupDir[d].ID == "" {
					id := NewULID()

					newQ = append(newQ, db.Quiz{
						ID:        id,
						UserID:    userID,
						Entry:     entry,
						Type:      body.Type,
						Direction: d,
					})

					subresult.IDs = append(subresult.IDs, id)
					ids = append(ids, id)
				} else {
					subresult.IDs = append(subresult.IDs, lookupDir[d].ID)
					ids = append(ids, lookupDir[d].ID)
				}
			}
		}

		if len(newExtra) > 0 {
			resource.DB.Current.Clauses(clause.OnConflict{
				DoNothing: true,
			}).Create(&newExtra)
		}

		if len(newQ) > 0 {
			if r := resource.DB.Current.CreateInBatches(newQ, 10); r.Error != nil {
				panic(r.Error)
			}
		}

		ctx.JSON(201, gin.H{
			"result": result,
			"ids":    ids,
		})
	})

	r.POST("/delete", func(ctx *gin.Context) {
		var body struct {
			IDs []string `json:"ids" binding:"required,min=1"`
		}

		if e := ctx.ShouldBindJSON(&body); e != nil {
			ctx.AbortWithError(400, e)
			return
		}

		if r := resource.DB.Current.Where("id IN ?", body.IDs).Delete(&db.Quiz{}); r.Error != nil {
			panic(r.Error)
		}

		ctx.JSON(201, gin.H{
			"result": "deleted",
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
