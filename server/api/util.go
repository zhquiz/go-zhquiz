package api

import (
	"bytes"
	"log"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	ulid "github.com/oklog/ulid/v2"
	"github.com/zhquiz/go-zhquiz/server/rand"
)

func getUserID(ctx *gin.Context) string {
	session := sessions.Default(ctx)
	k := session.Get("userID")

	if k == nil {
		return ""
	}

	return k.(string)
}

// NewULID generates new ULID
func NewULID() string {
	entropy, err := rand.GenerateRandomBytes(64)
	if err != nil {
		log.Fatalln(err)
	}
	return ulid.MustNew(ulid.Now(), bytes.NewReader(entropy)).String()
}
