package api

import (
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/wangbin/jiebago"
	"github.com/zhquiz/go-zhquiz/server/db"
	"github.com/zhquiz/go-zhquiz/server/zh"
	"github.com/zhquiz/go-zhquiz/shared"
)

var resource Resource
var validate *validator.Validate = validator.New()
var jieba jiebago.Segmenter

// Resource is a struct for reuse and cleanup.
type Resource struct {
	DB db.DB
	Zh zh.DB
}

// Cleanup cleans up Resource.
func (res Resource) Cleanup() {
	log.Println("Cleaning up")
	res.DB.Current.Commit()
}

// Prepare initializes Resource for reuse and cleanup.
func Prepare() Resource {
	f, _ := os.Create(filepath.Join(shared.ExecDir, "gin.log"))
	gin.DefaultWriter = io.MultiWriter(f, os.Stdout)

	resource = Resource{
		DB: db.Connect(),
		Zh: zh.Connect(),
	}

	jieba.LoadDictionary(filepath.Join(shared.ExecDir, "assets", "dict.txt"))

	return resource
}

// Register registers API paths to Gin Engine.
func (res Resource) Register(r *gin.Engine) {
	r.GET("/server/settings", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"ready": true,
		})
	})

	apiRouter := r.Group("/api")

	routerChinese(apiRouter)
	routerExtra(apiRouter)
	routerHanzi(apiRouter)
	routerLibrary(apiRouter)
	routerQuiz(apiRouter)
	routerSentence(apiRouter)
	routerUser(apiRouter)
	routerVocab(apiRouter)
}
