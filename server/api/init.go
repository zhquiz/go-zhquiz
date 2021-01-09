package api

import (
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/wangbin/jiebago"
	"github.com/zhquiz/go-zhquiz/server/db"
	"github.com/zhquiz/go-zhquiz/server/zh"
	"github.com/zhquiz/go-zhquiz/shared"
	"gorm.io/gorm"
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
	r.Use(sessions.Sessions("session", cookie.NewStore([]byte(shared.APISecret()))))

	r.GET("/server/settings", func(ctx *gin.Context) {
		speak := "web"
		if shared.SpeakFn() != "" {
			speak = "server"
		}

		ctx.JSON(200, gin.H{
			"speak": speak,
		})
	})

	// Send media files
	r.GET("/media/:filename", func(c *gin.Context) {
		filePath := filepath.Join(shared.MediaPath(), c.Param("filename"))
		if fileInfo, err := os.Stat(filePath); err == nil && !fileInfo.IsDir() {
			c.File(filePath)
			return
		}

		c.Status(404)
	})

	apiRouter := r.Group("/api")
	apiRouter.Use(AuthMiddleware())

	routerChinese(apiRouter)
	routerExtra(apiRouter)
	routerHanzi(apiRouter)
	routerLibrary(apiRouter)
	routerMedia(apiRouter)
	routerQuiz(apiRouter)
	routerSentence(apiRouter)
	routerUser(apiRouter)
	routerVocab(apiRouter)
}

// AuthMiddleware middleware for auth with user_id
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		var dbUser db.User

		r := resource.DB.Current.First(&dbUser)

		if errors.Is(r.Error, gorm.ErrRecordNotFound) {
			dbUser = db.User{}
			dbUser.New()

			if rCreate := resource.DB.Current.Create(&dbUser); rCreate.Error != nil {
				panic(rCreate.Error)
			}
		} else if r.Error != nil {
			panic(r.Error)
		}

		session.Set("userID", dbUser.ID)
	}
}
