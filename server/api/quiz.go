package api

import (
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jkomyno/nanoid"
	"github.com/zhquiz/go-zhquiz/server/db"
	"github.com/zhquiz/go-zhquiz/server/util"
	"github.com/zhquiz/go-zhquiz/server/zh"
	"gorm.io/gorm"
)

func routerQuiz(apiRouter *gin.RouterGroup) {
	r := apiRouter.Group("/quiz")

	r.GET("/many", func(ctx *gin.Context) {
		var query struct {
			IDs       string `form:"ids"`
			Entries   string `form:"entries"`
			Type      string `form:"type" binding:"oneof=hanzi vocab sentence extra ''"`
			Source    string `form:"source"`
			Direction string `form:"direction"`
			Select    string `form:"select"`
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
			IDs:       ids,
			Entries:   entries,
			Type:      query.Type,
			Source:    query.Source,
			Direction: query.Direction,
			Select:    strings.Split(query.Select, ","),
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

	r.GET("/leech", func(ctx *gin.Context) {
		var query struct {
			Q       string `form:"q"`
			Page    string `form:"page" binding:"required"`
			PerPage string `form:"perPage" binding:"required"`
			Sort    string `form:"sort"`
			Order   string `form:"order" binding:"oneof=desc asc"`
		}

		if e := ctx.BindQuery(&query); e != nil {
			ctx.AbortWithError(400, e)
			return
		}

		page := 1
		p, e := strconv.Atoi(query.Page)
		if e != nil {
			ctx.AbortWithError(400, e)
			return
		}
		page = p

		perPage := 5
		p, e = strconv.Atoi(query.PerPage)
		if e != nil {
			ctx.AbortWithError(400, e)
			return
		}
		perPage = p

		sort := map[string]string{
			"id":        "id",
			"entry":     "entry",
			"type":      "[type]",
			"direction": "direction",
			"source":    "source",
		}[query.Sort]

		if sort == "" {
			sort = "wrong_streak DESC, last_right DESC"
		} else {
			sort = sort + " " + query.Order
		}

		result := make([]db.Quiz, 0)

		var count int64 = 0

		q := resource.DB.Current.Model(&db.Quiz{}).
			Select("id", "entry", "type", "direction", "last_right", "wrong_streak").
			Where("wrong_streak >= 2")

		if query.Q != "" {
			q = q.Where(qSearch(query.Q))
		}

		if r := q.Count(&count); r.Error != nil {
			panic(r.Error)
		}

		if r := q.Limit(perPage).Order(sort).Offset((page - 1) * perPage).Find(&result); r.Error != nil {
			panic(r.Error)
		}

		ctx.JSON(200, gin.H{
			"result": result,
			"count":  count,
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
			user.Meta.Settings.Quiz.Q = query.Q

			if r := resource.DB.Current.Where("id = ?", user.ID).Updates(&db.User{
				Meta: user.Meta,
			}); r.Error != nil {
				panic(r.Error)
			}
		}()

		q := resource.DB.Current.Model(&db.Quiz{})

		if query.Q != "" {
			q = q.Where(qSearch(query.Q))
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

		quiz := make([]quizInitOutput, 0)
		upcoming := make([]quizInitOutput, 0)

		if query.IncludeUndue == "" {
			now := time.Now()

			for _, it := range quizzes {
				if it.NextReview == nil || (*it.NextReview).Before(now) {
					quiz = append(quiz, quizInitOutput{
						NextReview:  it.NextReview,
						SRSLevel:    it.SRSLevel,
						WrongStreak: it.WrongStreak,
						ID:          it.ID,
						Entry:       it.Entry,
						Direction:   it.Direction,
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

		remainingQuiz := quiz[:]
		quiz = []quizInitOutput{}

	RAND_LOOP:
		for {
			switch len(remainingQuiz) {
			case 0:
				break RAND_LOOP
			case 1:
				quiz = append(quiz, remainingQuiz[0])
				break RAND_LOOP
			case 2:
				if len(quiz) > 0 {
					if quiz[0].Entry == remainingQuiz[0].Entry {
						quiz = append(quiz, remainingQuiz[1], remainingQuiz[0])
					} else {
						quiz = append(quiz, remainingQuiz...)
					}
				} else {
					quiz = remainingQuiz
				}
				break RAND_LOOP
			}

			entry := ""

			if len(quiz) > 0 {
				entry = quiz[0].Entry
			}

			n := rand.Intn(len(remainingQuiz))
			current := remainingQuiz[n]

			if current.Entry != entry {
				quiz = append(quiz, current)

				clone := remainingQuiz
				remainingQuiz = []quizInitOutput{}

				if n > 0 {
					remainingQuiz = append(remainingQuiz, clone[:n]...)
				}

				if n+1 < len(clone) {
					remainingQuiz = append(remainingQuiz, clone[n+1:]...)
				}
			}
		}

		sort.Sort(quizInitOutputList(upcoming))

		ctx.JSON(200, gin.H{
			"quiz":     quiz,
			"upcoming": upcoming,
		})
	})

	r.PUT("/", func(ctx *gin.Context) {
		var body struct {
			Entries     []string          `json:"entries" binding:"required,min=1"`
			Type        string            `json:"type" binding:"required,oneof=hanzi vocab sentence extra"`
			Description string            `json:"description"`
			Pinyin      map[string]string `json:"pinyin"`
			English     map[string]string `json:"english"`
		}
		if e := ctx.BindJSON(&body); e != nil {
			ctx.AbortWithError(400, e)
		}

		if body.Pinyin == nil {
			body.Pinyin = make(map[string]string)
		}

		if body.English == nil {
			body.English = make(map[string]string)
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
				Type:   body.Type,
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
				pinyin := body.Pinyin[entry]
				english := body.English[entry]

				if pinyin == "" || english == "" {
					pSegs := make([]string, 0)
					eSegs := make([]string, 0)
					reHan := regexp.MustCompile("\\p{Han}+")

					for _, seg := range cutChinese(entry) {
						if reHan.MatchString(seg) {
							var vocab zh.Vocab
							if r := resource.Zh.Current.Where("simplified = ? OR traditional = ?", seg, seg).Order("frequency DESC").First(&vocab); r.Error != nil {
								if !errors.Is(r.Error, gorm.ErrRecordNotFound) {
									panic(r.Error)
								}
							}

							if vocab.English != "" {
								pSegs = append(pSegs, vocab.Pinyin)
								eSegs = append(eSegs, vocab.English)
							} else {
								pSegs = append(pSegs, seg)
								eSegs = append(eSegs, seg)
							}
						} else {
							pSegs = append(pSegs, seg)
							eSegs = append(eSegs, seg)
						}
					}

					if pinyin == "" {
						pinyin = strings.Join(pSegs, " ")
					}

					if english == "" {
						english = strings.Join(eSegs, "; ")
					}
				}

				newExtra = append(newExtra, db.Extra{
					Chinese:     entry,
					Pinyin:      pinyin,
					English:     english,
					Type:        subresult.Type,
					Description: body.Description,
				})
			}

			lookupDir := lookup[entry]
			if lookupDir == nil {
				lookupDir = map[string]db.Quiz{}
			}

			for _, d := range directions {
				if lookupDir[d].ID == "" {
					id := ""

					for {
						id1, err := nanoid.Nanoid(6)
						if err != nil {
							panic(err)
						}

						var count int64
						if r := resource.DB.Current.Model(db.Quiz{}).Where("id = ?", id1).Count(&count); r.Error != nil {
							panic(err)
						}

						if count == 0 {
							id = id1
							break
						}
					}

					newQ = append(newQ, db.Quiz{
						ID:          id,
						Entry:       entry,
						Type:        subresult.Type,
						Direction:   d,
						Source:      subresult.Source,
						Description: body.Description,
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
			for _, it := range newExtra {
				it.Create(tx)
			}

			for _, it := range newQ {
				it.Create(tx)
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
			for _, id := range body.IDs {
				q := db.Quiz{
					ID: id,
				}

				if e := q.Delete(tx); e != nil {
					return e
				}
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
	Entry       string     `json:"-"`
	Direction   string     `json:"-"`
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
	IDs       []string `json:"ids"`
	Entries   []string `json:"entries"`
	Type      string   `json:"type"`
	Source    string   `json:"source"`
	Direction string   `json:"direction"`
	Select    []string `json:"select" binding:"required,min=1"`
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

	if body.Source != "" {
		andWhere = append(andWhere, "quiz.source = @source")
		cond["source"] = body.Source
	}

	if body.Direction != "" {
		andWhere = append(andWhere, "quiz.direction = @direction")
		cond["direction"] = body.Direction
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

func qSearch(q string) *gorm.DB {
	qBuilder := resource.DB.Current.Model(&db.Quiz{})
	segs := make([]string, 0)

	reCmp := regexp.MustCompile("([><]=?)(\\d+)")

	parseV := func(k string, v string, myBuilder *gorm.DB) error {
		for _, m := range reCmp.FindAllSubmatch([]byte(v), -1) {
			if len(m) > 2 {
				n, _ := strconv.Atoi(string(m[2]))
				myBuilder = myBuilder.Where(fmt.Sprintf("%s %s ?", k, string(m[1])), n)
			}
		}

		if n, e := strconv.Atoi(v); e == nil {
			myBuilder = myBuilder.Where(fmt.Sprintf("%s = ?", k), n)
		}

		return nil
	}

	level := ""
	for _, seg := range strings.Split(q, " ") {
		kv := strings.SplitN(seg, ":", 2)

		switch kv[0] {
		case "srsLevel":
			if e := parseV("srs_level", kv[1], qBuilder); e != nil {
				panic(e)
			}
		case "level":
			level = kv[1]
		default:
			if seg != "" {
				segs = append(segs, seg)
			}
		}
	}

	if len(segs) > 0 {
		qBuilder = qBuilder.Where(`id IN (
			SELECT id FROM quiz_q WHERE quiz_q MATCH ?
		)`, strings.Join(segs, " "))
	}

	if level != "" {
		orCond := resource.DB.Current

		quizzes := make([]db.Quiz, 0)
		if r := qBuilder.Find(&quizzes); r.Error != nil {
			panic(r.Error)
		}

		hanzis := make([]string, 0)
		vocabs := make([]string, 0)
		sentences := make([]string, 0)

		for _, el := range quizzes {
			if el.Source == "extra" {
				continue
			}

			switch el.Type {
			case "hanzi":
				hanzis = append(hanzis, el.Entry)
			case "vocab":
				vocabs = append(vocabs, el.Entry)
			case "sentence":
				sentences = append(sentences, el.Entry)
			}
		}

		if len(hanzis) > 0 {
			ts := make([]zh.Token, 0)
			tBuilder := resource.Zh.Current.Where("entry IN ?", hanzis)
			if e := parseV("hanzi_level", level, tBuilder); e != nil {
				panic(e)
			}
			if r := tBuilder.Find(&ts); r.Error != nil {
				panic(r.Error)
			}
			hanzis = []string{}
			for _, t := range ts {
				hanzis = append(hanzis, t.Entry)
			}

			if len(hanzis) > 0 {
				orCond = orCond.Or("[type] = 'hanzi' AND [entry] IN ?", hanzis)
			}
		}

		if len(vocabs) > 0 {
			ts := make([]zh.Token, 0)
			tBuilder := resource.Zh.Current.Where("entry IN ?", vocabs)
			if e := parseV("vocab_level", level, tBuilder); e != nil {
				panic(e)
			}
			if r := tBuilder.Find(&ts); r.Error != nil {
				panic(r.Error)
			}
			vocabs = []string{}
			for _, t := range ts {
				vocabs = append(vocabs, t.Entry)
			}

			if len(vocabs) > 0 {
				orCond = orCond.Or("[type] = 'vocab' AND [entry] IN ?", vocabs)
			}
		}

		if len(sentences) > 0 {
			ts := make([]zh.Sentence, 0)
			tBuilder := resource.Zh.Current.Where("chinese IN ?", sentences)
			if e := parseV("level", level, tBuilder); e != nil {
				panic(e)
			}
			if r := tBuilder.Find(&ts); r.Error != nil {
				panic(r.Error)
			}
			sentences = []string{}
			for _, t := range ts {
				sentences = append(sentences, t.Chinese)
			}

			if len(sentences) > 0 {
				orCond = orCond.Or("[type] = 'sentence' AND [entry] IN ?", sentences)
			}
		}

		qBuilder = qBuilder.Where(orCond.Or("FALSE"))
	}

	return qBuilder
}
