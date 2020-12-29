package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-contrib/cache/persistence"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/oklog/ulid/v2"
	"github.com/zhquiz/go-server/server/db"
	"github.com/zhquiz/go-server/server/rand"
	"github.com/zhquiz/go-server/server/zh"
	"github.com/zhquiz/go-server/shared"
	"gorm.io/gorm"
)

var resource Resource
var validate *validator.Validate = validator.New()
var persist *persistence.InMemoryStore = persistence.NewInMemoryStore(time.Hour)

// Resource is a struct for reuse and cleanup.
type Resource struct {
	DB          db.DB
	Zh          zh.DB
	ulidEntropy io.Reader
}

// NewULID generates new ULID
func (res Resource) NewULID() string {
	return ulid.MustNew(ulid.Now(), res.ulidEntropy).String()
}

// Prepare initializes Resource for reuse and cleanup.
func Prepare() Resource {
	f, _ := os.Create(filepath.Join(shared.Paths().Root, "gin.log"))
	gin.DefaultWriter = io.MultiWriter(f, os.Stdout)

	entropy, err := rand.GenerateRandomBytes(64)
	if err != nil {
		log.Fatalln(err)
	}

	resource = Resource{
		DB:          db.Connect(),
		Zh:          zh.Connect(),
		ulidEntropy: bytes.NewReader(entropy),
	}

	return resource
}

// Register registers API paths to Gin Engine.
func (res Resource) Register(r *gin.Engine) {
	apiSecret := shared.GetenvOrDefaultFn("ZHQUIZ_API_SECRET", func() string {
		s, err := rand.GenerateRandomString(64)
		if err != nil {
			log.Fatalln(err)
		}
		return s
	})

	r.Use(sessions.Sessions("session", cookie.NewStore([]byte(apiSecret))))

	cotterAPIKey := os.Getenv("COTTER_API_KEY")

	if cotterAPIKey != "" {
		r.GET("/server/auth/cotter", func(ctx *gin.Context) {
			ctx.JSON(200, gin.H{
				"apiKey": cotterAPIKey,
			})
		})
	}

	// Send media files
	r.GET("/media/:filename", func(c *gin.Context) {
		filePath := filepath.Join(shared.Paths().MediaPath(), c.Param("filename"))
		if fileInfo, err := os.Stat(filePath); err == nil && !fileInfo.IsDir() {
			c.File(filePath)
			return
		}

		c.Status(404)
	})

	apiRouter := r.Group("/api")
	apiRouter.Use(CotterAuthMiddleware())

	routerChinese(apiRouter)
	routerExtra(apiRouter)
	routerHanzi(apiRouter)
	routerMedia(apiRouter)
	routerQuiz(apiRouter)
	routerSentence(apiRouter)
	routerUser(apiRouter)
	routerVocab(apiRouter)
}

// Cleanup cleans up Resource.
func (res Resource) Cleanup() {
	res.DB.Current.Commit()
}

// CotterAuthMiddleware middleware for auth with Cotter
func CotterAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		cotterAPIKey := os.Getenv("COTTER_API_KEY")

		if cotterAPIKey == "" {
			return
		}

		session := sessions.Default(c)

		authorization := c.GetHeader("Authorization")
		userName := c.GetHeader("X-User")

		if strings.HasPrefix(authorization, "Bearer ") {
			idToken := strings.Split(authorization, " ")[1]

			ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(30*time.Second))
			defer cancel()

			reqBody, err := json.Marshal(gin.H{
				"oauth_token": gin.H{
					"access_token": idToken,
				},
			})
			if err != nil {
				panic(err)
			}

			client := &http.Client{}
			req, err := http.NewRequestWithContext(ctx, "POST", "https://worker.cotter.app/verify", bytes.NewBuffer(reqBody))
			if err != nil {
				panic(err)
			}

			req.Header.Add("API_KEY_ID", cotterAPIKey)
			req.Header.Add("Content-Type", "application/json")

			res, err := client.Do(req)
			if err != nil {
				panic(err)
			}

			defer res.Body.Close()

			resBody, err := ioutil.ReadAll(res.Body)
			if err != nil {
				panic(err)
			}

			var resObj struct {
				Success bool
			}

			if err := json.Unmarshal(resBody, &resObj); err != nil {
				panic(err)
			}

			if resObj.Success {
				var dbUser db.User

				r := resource.DB.Current.Where("email = ?", userName).First(&dbUser)

				if errors.Is(r.Error, gorm.ErrRecordNotFound) {
					dbUser = db.User{}
					dbUser.New(resource.NewULID(), userName)

					if rCreate := resource.DB.Current.Create(&dbUser); rCreate.Error != nil {
						panic(rCreate.Error)
					}
				} else if r.Error != nil {
					panic(r.Error)
				}

				session.Set("userID", dbUser.ID)
				return
			}
		}

		session.Set("userID", "")
	}
}
