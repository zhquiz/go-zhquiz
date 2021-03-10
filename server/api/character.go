package api

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zhquiz/go-zhquiz/server/db"
	"github.com/zhquiz/go-zhquiz/server/zh"
	"gorm.io/gorm"
)

type rCharacter struct {
	Base *gin.RouterGroup
}

func (r rCharacter) init() {
	router := r.Base.Group("/character")

	router.GET("/", r.getOne)
	router.GET("/q", r.getQuery)
	router.GET("/random", r.getRandom)
}

// @Produce json
// @Param entry query string true "hanzi entry"
// @Success 200 {object} CharacterItemFull
// @Router /character/ [get]
func (rCharacter) getOne(ctx *gin.Context) {
	var query struct {
		Entry string `form:"entry" binding:"required"`
	}

	if e := ctx.BindQuery(&query); e != nil {
		ctx.AbortWithError(400, e)
		return
	}

	var out CharacterItemFull

	if r := resource.Zh.Raw(`
	SELECT
		(
			SELECT GROUP_CONCAT(child, '')
			FROM token_sub s
			JOIN token t ON t.entry = s.child
			WHERE parent = entry
			ORDER BY frequency
			GROUP BY parent
		) Sub,
		(
			SELECT GROUP_CONCAT(child, '')
			FROM token_sup s
			JOIN token t ON t.entry = s.child
			WHERE parent = entry
			ORDER BY frequency
			GROUP BY parent
		) Sup,
		(
			SELECT GROUP_CONCAT(child, '')
			FROM token_var s
			JOIN token t ON t.entry = s.child
			WHERE parent = entry
			ORDER BY frequency
			GROUP BY parent
		) Variants,
		Pinyin,
		English
	FROM token
	WHERE [entry] = ?
	`, query.Entry).First(&out); r.Error != nil {
		if errors.Is(r.Error, gorm.ErrRecordNotFound) {
			ctx.AbortWithStatus(404)
			return
		}

		panic(r.Error)
	}

	ctx.JSON(200, out)
}

// CharacterItemFull -
type CharacterItemFull struct {
	Sub      string `json:"sub"`
	Sup      string `json:"sup"`
	Variants string `json:"variants"`
	Pinyin   string `json:"pinyin"`
	English  string `json:"english"`
}

// @Produce json
// @Param q query string true "search term"
// @Success 200 {object} EntryQueryResponse
// @Router /character/q [get]
func (rCharacter) getQuery(ctx *gin.Context) {
	var query struct {
		Q string `form:"q" binding:"required"`
	}

	if e := ctx.BindQuery(&query); e != nil {
		ctx.AbortWithError(400, e)
		return
	}

	result := make([]zh.Token, 0)

	q := resource.Zh.Where("entry = ? AND length(entry) = 1", query)

	if r := q.Select("entry").Order("frequency DESC").Find(&result); r.Error != nil {
		panic(r.Error)
	}

	out := EntryQueryResponse{
		Result: make([]string, 0),
	}

	for _, t := range result {
		out.Result = append(out.Result, t.Entry)
	}

	ctx.JSON(200, out)
}

// EntryQueryResponse -
type EntryQueryResponse struct {
	Result []string `json:"result"`
}

// @Produce json
// @Success 200 {object} RandomResponse
// @Router /character/random [get]
func (rCharacter) getRandom(ctx *gin.Context) {
	u := userID(ctx)
	if u == 0 {
		return
	}

	user := db.User{
		ID: u,
	}
	if r := resource.DB.First(&user); r.Error != nil {
		panic(r.Error)
	}
	levelMin := *user.LevelMin
	if levelMin == 0 {
		levelMin = 1
	}
	levelMax := *user.Level
	if levelMax == 0 {
		levelMax = 10
	}

	var existing []db.Quiz
	if r := resource.DB.
		Where("[type] = 'hanzi' AND srs_level IS NOT NULL AND next_review IS NOT NULL").
		Find(&existing); r.Error != nil {
		panic(r.Error)
	}

	var entries []interface{}
	for _, it := range existing {
		entries = append(entries, it.Entry)
	}

	params := map[string]interface{}{
		"entries":  entries,
		"levelMin": levelMin,
		"level":    levelMax,
	}

	where := "english IS NOT NULL AND hanzi_level >= @levelMin AND hanzi_level <= @level"
	if len(entries) > 0 {
		where = "entry NOT IN @entries AND " + where
	}

	var items []RandomResponse

	if r := resource.Zh.Raw(fmt.Sprintf(`
		SELECT entry Result, English, hanzi_level Level
		FROM token
		WHERE %s
		`, where), params).Find(&items); r.Error != nil {
		panic(r.Error)
	}

	if len(items) < 1 {
		where := "english IS NOT NULL"
		if len(entries) > 0 {
			where = "entry NOT IN @entries AND " + where
		}

		if r := resource.Zh.Raw(fmt.Sprintf(`
			SELECT entry Result, English, hanzi_level Level
			FROM token
			WHERE %s
			`, where), params).Find(&items); r.Error != nil {
			panic(r.Error)
		}
	}

	if len(items) < 1 {
		ctx.AbortWithError(404, fmt.Errorf("no matching entries found"))
		return
	}

	rand.Seed(time.Now().UnixNano())
	ctx.JSON(200, items[rand.Intn(len(items))])
}
