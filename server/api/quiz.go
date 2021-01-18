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
	myrand "github.com/zhquiz/go-zhquiz/server/rand"
	"github.com/zhquiz/go-zhquiz/server/util"
	"github.com/zhquiz/go-zhquiz/server/zh"
	"gorm.io/gorm"
)

func routerQuiz(apiRouter *gin.RouterGroup) {
	r := apiRouter.Group("/quiz")

	r.GET("/many", func(ctx *gin.Context) {
		var query struct {
			IDs     string `form:"ids"`
			Entries string `form:"entries"`
			Type    string `form:"type" binding:"oneof=hanzi vocab sentence extra ''"`
			Select  string `form:"select"`
		}

		if e := ctx.BindQuery(&query); e != nil {
			ctx.AbortWithError(400, e)
		}

		var ids []string
		if query.IDs != "" {
			ids = strings.Split(query.IDs, ",")
		}

		var entries []string
		if query.Entries != "" {
			entries = strings.Split(query.Entries, ",")
		}

		quizGetter(ctx, getterBody{
			IDs:     ids,
			Entries: entries,
			Type:    query.Type,
			Select:  strings.Split(query.Select, ","),
		})
	})

	r.POST("/many", func(ctx *gin.Context) {
		var body getterBody

		if e := ctx.BindJSON(&body); e != nil {
			ctx.AbortWithError(400, e)
		}

		quizGetter(ctx, body)
	})

	r.POST("/srsLevel", func(ctx *gin.Context) {
		var body struct {
			Entries []string `form:"entries" binding:"required,min=1"`
			Type    string   `form:"type" binding:"oneof=hanzi vocab sentence extra ''"`
		}

		if e := ctx.BindJSON(&body); e != nil {
			ctx.AbortWithError(400, e)
			return
		}

		out := make([]gin.H, 0)

		chunkSize := 500
		for i := 0; i < len(body.Entries); i += chunkSize {
			chunkEnd := i + chunkSize
			if chunkEnd > len(body.Entries) {
				chunkEnd = len(body.Entries)
			}

			where := "[entry] IN @entries AND [Type] IN @type"
			cond := map[string]interface{}{
				"entries": body.Entries[i:chunkEnd],
				"type":    []string{body.Type, "extra"},
			}

			var quizzes []db.Quiz

			clause := resource.DB.Current.Model(&db.Quiz{}).
				Select("entry", "srs_level").
				Where(where, cond)

			if r := clause.Find(&quizzes); r.Error != nil {
				panic(r.Error)
			}

			for _, q := range quizzes {
				out = append(out, gin.H{
					"entry":    q.Entry,
					"srsLevel": q.SRSLevel,
				})
			}
		}

		ctx.JSON(200, gin.H{
			"result": out,
		})
	})

	r.PATCH("/mark", func(ctx *gin.Context) {
		var query struct {
			ID   string `form:"id" binding:"required"`
			Type string `form:"type" binding:"required,oneof=right wrong repeat"`
		}

		if e := ctx.BindQuery(&query); e != nil {
			ctx.AbortWithError(400, e)
			return
		}

		var quiz db.Quiz
		if r := resource.DB.Current.
			Where("id = ?", query.ID).
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

	r.GET("/init", func(ctx *gin.Context) {
		var query struct {
			Type         string `form:"type"`
			Stage        string `form:"stage"`
			Direction    string `form:"direction"`
			IncludeUndue string `form:"includeUndue"`
			IncludeExtra string `form:"includeExtra"`
			Q            string `form:"q"`
		}

		if e := ctx.BindQuery(&query); e != nil {
			ctx.AbortWithError(400, e)
			return
		}

		if len(query.Type) == 0 {
			ctx.JSON(200, gin.H{
				"quiz":     make([]string, 0),
				"upcoming": make([]string, 0),
			})
			return
		}

		if len(query.Stage) == 0 {
			ctx.JSON(200, gin.H{
				"quiz":     make([]string, 0),
				"upcoming": make([]string, 0),
			})
			return
		}

		if len(query.Direction) == 0 {
			ctx.JSON(200, gin.H{
				"quiz":     make([]string, 0),
				"upcoming": make([]string, 0),
			})
			return
		}

		qType := strings.Split(query.Type, ",")
		stage := strings.Split(query.Stage, ",")
		direction := strings.Split(query.Direction, ",")

		// No need to await
		go func() {
			var user db.User
			if r := resource.DB.Current.First(&user); r.Error != nil {
				panic(r.Error)
			}

			user.Meta.Settings.Quiz.Direction = direction
			user.Meta.Settings.Quiz.Stage = stage
			user.Meta.Settings.Quiz.Type = qType
			user.Meta.Settings.Quiz.IncludeExtra = (query.IncludeExtra != "")
			user.Meta.Settings.Quiz.IncludeUndue = (query.IncludeUndue != "")

			if r := resource.DB.Current.Where("id = ?", user.ID).Updates(&db.User{
				Meta: user.Meta,
			}); r.Error != nil {
				panic(r.Error)
			}
		}()

		q := resource.DB.Current.Model(&db.Quiz{})

		if query.Q != "" {
			var sel []struct {
				ID string
			}
			// Ignore errors
			resource.DB.Current.Raw("SELECT id FROM quiz_q WHERE quiz_q MATCH ?", query.Q).Find(&sel)

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

		stageSet := util.MakeSet(stage)
		if stageSet["new"] {
			orCond = append(orCond, "srs_level IS NULL")
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

		if !stageSet["leech"] {
			q = q.Where("NOT (wrong_streak > 2)")
		}

		var quizzes []db.Quiz

		if r := q.
			Where("[type] IN ? AND [direction] IN ?", qType, direction).
			Find(&quizzes); r.Error != nil {
			panic(r.Error)
		}

		var quiz []quizInitOutput
		var upcoming []quizInitOutput

		if query.IncludeUndue == "" {
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
		var body struct {
			Entries     []string `json:"entries" binding:"required,min=1"`
			Type        string   `json:"type" binding:"required,oneof=hanzi vocab sentence extra"`
			Description string   `json:"description"`
		}
		if e := ctx.BindJSON(&body); e != nil {
			ctx.AbortWithError(400, e)
		}

		var existingQ []db.Quiz

		if r := resource.DB.Current.
			Where("entry IN ? AND type = ?", body.Entries, body.Type).
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
			IDs    []string `json:"ids"`
			Entry  string   `json:"entry"`
			Type   string   `json:"type"`
			Source string   `json:"source"`
		}
		result := make([]Result, 0)
		ids := make([]string, 0)

		var newQ []db.Quiz
		var newExtra []db.Extra

		for _, entry := range body.Entries {
			subresult := Result{
				IDs:    make([]string, 0),
				Entry:  entry,
				Type:   "vocab",
				Source: "",
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
				} else {
					subresult.Source = "extra"
				}
			case "hanzi":
				if r := resource.Zh.Current.
					Where("entry = ? AND length(entry) = 1 AND english IS NOT NULL", entry).
					First(&zh.Token{}); r.Error != nil {
					if !errors.Is(r.Error, gorm.ErrRecordNotFound) {
						panic(r.Error)
					}
					subresult.Source = "extra"
				}
			case "sentence":
				if r := resource.Zh.Current.
					Where("chinese = ?", entry).
					First(&zh.Sentence{}); r.Error != nil {
					if !errors.Is(r.Error, gorm.ErrRecordNotFound) {
						panic(r.Error)
					}
					subresult.Source = "extra"
				}
			}

			if subresult.Source == "extra" {
				pinyin := ""
				english := ""

				for _, seg := range cutChinese(entry) {
					var vocab zh.Vocab
					if r := resource.Zh.Current.Where("simplified = ? OR traditional = ?", seg, seg).Order("frequency DESC").First(&vocab); r.Error != nil {
						if !errors.Is(r.Error, gorm.ErrRecordNotFound) {
							panic(r.Error)
						}
					}

					if vocab.English != "" {
						if english != "" {
							pinyin += " "
							english += "; "
						}

						pinyin += vocab.Pinyin
						english += vocab.English
					}
				}

				newExtra = append(newExtra, db.Extra{
					Chinese: entry,
					Pinyin:  pinyin,
					English: english,
					Type:    subresult.Type,
				})
			}

			lookupDir := lookup[entry]
			if lookupDir == nil {
				lookupDir = map[string]db.Quiz{}
			}

			for _, d := range directions {
				if lookupDir[d].ID == "" {
					id := myrand.NewULID()

					newQ = append(newQ, db.Quiz{
						ID:        id,
						Entry:     entry,
						Type:      subresult.Type,
						Direction: d,
						Source:    subresult.Source,
					})

					subresult.IDs = append(subresult.IDs, id)
					ids = append(ids, id)
				} else {
					subresult.IDs = append(subresult.IDs, lookupDir[d].ID)
					ids = append(ids, lookupDir[d].ID)
				}
			}
		}

		e := resource.DB.Current.Transaction(func(tx *gorm.DB) error {
			if len(newExtra) > 0 {
				if r := tx.CreateInBatches(&newExtra, 5); r.Error != nil {
					return r.Error
				}
			}

			if len(newQ) > 0 {
				if r := tx.CreateInBatches(&newQ, 5); r.Error != nil {
					return r.Error
				}
			}

			return nil
		})

		if e != nil {
			panic(e)
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

		if e := ctx.BindJSON(&body); e != nil {
			ctx.AbortWithError(400, e)
			return
		}

		e := resource.DB.Current.Transaction(func(tx *gorm.DB) error {
			if r := tx.Where("id IN ?", body.IDs).Delete(&db.Quiz{}); r.Error != nil {
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

type getterBody struct {
	IDs     []string `json:"ids"`
	Entries []string `json:"entries"`
	Type    string   `json:"type"`
	Select  []string `json:"select" binding:"required,min=1"`
}

func quizGetter(ctx *gin.Context, body getterBody) {
	sel := []string{}
	sMap := map[string]string{
		"id":        "quiz.id id",
		"entry":     "quiz.entry entry",
		"type":      "quiz.type type",
		"direction": "quiz.direction direction",
		"source":    "quiz.source source",
	}

	for _, s := range body.Select {
		k := sMap[s]
		if k != "" && k != "_" {
			sel = append(sel, k)
		}
	}

	if len(sel) == 0 {
		ctx.AbortWithError(400, fmt.Errorf("not enough select"))
		return
	}

	andWhere := make([]string, 0)
	cond := map[string]interface{}{}

	if len(body.IDs) > 0 {
		andWhere = append(andWhere, "quiz.id IN @ids")
		cond["ids"] = body.IDs
	} else if len(body.Entries) > 0 {
		andWhere = append(andWhere, "quiz.entry IN @entries")
		cond["entries"] = body.Entries
	} else {
		ctx.AbortWithError(400, fmt.Errorf("either IDs or Entries must be specified"))
		return
	}

	if body.Type != "" {
		andWhere = append(andWhere, "quiz.type = @type")
		cond["type"] = body.Type
	}

	out := make([]map[string]interface{}, 0)

	clause := resource.DB.Current.Model(&db.Quiz{}).
		Select(sel).
		Where(strings.Join(andWhere, " AND "), cond)

	if r := clause.Find(&out); r.Error != nil {
		panic(r.Error)
	}

	ctx.JSON(200, gin.H{
		"result": out,
	})
}
