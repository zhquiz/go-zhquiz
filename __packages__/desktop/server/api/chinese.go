package api

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gin-gonic/gin"
)

func routerChinese(apiRouter *gin.RouterGroup) {
	r := apiRouter.Group("/chinese")

	r.GET("/jieba", func(ctx *gin.Context) {
		var query struct {
			Q string `form:"q" binding:"required"`
		}

		if e := ctx.BindQuery(&query); e != nil {
			ctx.AbortWithError(400, e)
			return
		}

		ctx.JSON(200, gin.H{
			"result": cutChineseAll(query.Q),
		})
	})

	r.GET("/speak", func(ctx *gin.Context) {
		var query struct {
			Q string `form:"q" binding:"required"`
		}

		if e := ctx.BindQuery(&query); e != nil {
			ctx.AbortWithError(400, e)
			return
		}

		params := url.Values{}
		params.Add("ie", "UTF-8")
		params.Add("tl", "zh-CN")
		params.Add("q", query.Q)
		params.Add("total", "1")
		params.Add("idx", "0")
		params.Add("client", "tw-ob")
		params.Add("textlen", strconv.Itoa(len(query.Q)))

		req, err := http.NewRequest("GET", fmt.Sprintf("http://translate.google.com/translate_tts?%s", params.Encode()), nil)
		if err != nil {
			log.Fatalln(err)
		}

		req.Header.Add("Referrer", "http://translate.google.com/")
		req.Header.Add("User-Agent", getUserAgent())
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded;charset=utf-8")

		client := &http.Client{}
		response, err := client.Do(req)
		if err != nil {
			panic(err)
		}

		defer response.Body.Close()

		ctx.DataFromReader(200, response.ContentLength, response.Header.Get("Content-Type"), response.Body, map[string]string{})
	})
}

func cutChineseAll(s string) []string {
	out := make([]string, 0)
	func(ch <-chan string) {
		for word := range ch {
			out = append(out, word)
		}
	}(jieba.CutAll(s))

	return out
}

func cutChinese(s string) []string {
	out := make([]string, 0)
	func(ch <-chan string) {
		for word := range ch {
			out = append(out, word)
		}
	}(jieba.Cut(s, true))

	return out
}
