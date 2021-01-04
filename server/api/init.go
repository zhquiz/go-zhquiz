package api

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/zhquiz/go-server/server/db"
	"github.com/zhquiz/go-server/server/rand"
	"github.com/zhquiz/go-server/server/zh"
	"github.com/zhquiz/go-server/shared"
	"gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"
	"gorm.io/gorm"
)

var resource Resource
var validate *validator.Validate = validator.New()

// Resource is a struct for reuse and cleanup.
type Resource struct {
	DB db.DB
	Zh zh.DB
}

// Prepare initializes Resource for reuse and cleanup.
func Prepare() Resource {
	f, _ := os.Create(filepath.Join(shared.Paths().Root, "gin.log"))
	gin.DefaultWriter = io.MultiWriter(f, os.Stdout)

	resource = Resource{
		DB: db.Connect(),
		Zh: zh.Connect(),
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

	zhQuizSpeak := shared.GetenvOrDefaultFn("ZHQUIZ_SPEAK", func() string {
		stat, err := os.Stat(filepath.Join(shared.Paths().Dir, "assets", "speak.sh"))
		if err == nil && !stat.IsDir() {
			return filepath.Join(shared.Paths().Dir, "assets", "speak.sh")
		}

		return "0"
	})
	r.GET("/server/settings", func(ctx *gin.Context) {
		speak := "web"
		if zhQuizSpeak != "0" {
			speak = "server"
		}

		ctx.JSON(200, gin.H{
			"speak":     speak,
			"plausible": os.Getenv("PLAUSIBLE"),
		})
	})

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
	const JWKSURL = "https://www.cotter.app/api/v0/token/jwks"
	const JWKSLookupKeyID = "SPACE_JWT_PUBLIC:8028AAA3-EC2D-4BAA-BE7A-7C8359CCB9F9"

	cotterAPIKey := os.Getenv("COTTER_API_KEY")

	var jwksKey []byte
	// Fetch the key from the JWKS URL
	getKey := func() []byte {
		if len(jwksKey) > 0 {
			return jwksKey
		}

		// Fetch the JWT Public Key from the URL
		resp, err := http.Get(JWKSURL)
		if err != nil {
			log.Fatalln(err)
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalln(err)
		}
		// Parse the response into our keys struct
		keyset := make(map[string][]map[string]interface{})
		err = json.Unmarshal(body, &keyset)
		if err != nil {
			log.Fatalln(err)
		}
		// It's a Key Set = there might be multiple keys
		// Find the key with kid = JWKSLookupKeyID
		if len(keyset["keys"]) <= 0 {
			log.Fatalln(errors.New("Key set is empty"))
		}
		for _, k := range keyset["keys"] {
			if k["kid"] == JWKSLookupKeyID {
				key, err := json.Marshal(k)
				if err != nil {
					log.Fatalln(err)
				}

				jwksKey = key
				return key
			}
		}

		log.Fatalln(errors.New("Cannot find key with kid"))
		return []byte{}
	}

	// validateClientAccessToken validates access token created above
	validateClientAccessToken := func(accessToken string) (map[string]interface{}, error) {
		tok, err := jwt.ParseSigned(accessToken)
		if err != nil {
			return nil, errors.New("Fail parsing access token")
		}
		keys := getKey()
		key := jose.JSONWebKey{}
		key.UnmarshalJSON(keys)
		token := make(map[string]interface{})
		if err := tok.Claims(key, &token); err != nil {
			return nil, errors.New("Fail parsing access token to claims")
		}
		// Check that the aud is our API KEY ID
		apiKeyID, ok := token["aud"].(string)
		if !ok {
			return nil, errors.New("fail asserting aud from jwt.MapClaims")
		}
		if apiKeyID != cotterAPIKey {
			return nil, errors.New("Invalid aud, not meant for this api key id")
		}
		return token, nil
	}

	return func(c *gin.Context) {
		if cotterAPIKey == "" {
			return
		}

		session := sessions.Default(c)

		authorization := c.GetHeader("Authorization")

		if strings.HasPrefix(authorization, "Bearer ") {
			accessToken := strings.Split(authorization, " ")[1]

			// Validate that the access token and signature is valid
			token, err := validateClientAccessToken(accessToken)
			if err != nil {
				session.Set("userID", "")
				return
			}

			userName := token["identifier"].(string)

			if userName != "" {
				var dbUser db.User

				r := resource.DB.Current.Where("email = ?", userName).First(&dbUser)

				if errors.Is(r.Error, gorm.ErrRecordNotFound) {
					dbUser = db.User{}
					dbUser.New(NewULID(), userName)

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
