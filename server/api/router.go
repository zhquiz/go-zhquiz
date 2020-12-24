package api

import (
	"time"

	"github.com/gin-contrib/cache/persistence"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate
var store *persistence.InMemoryStore

type tAPIRouter struct {
	Router *gin.RouterGroup
}

func (t tAPIRouter) init() {
	validate = validator.New()
	store = persistence.NewInMemoryStore(time.Hour)

	tChineseRouter{
		Router: t.Router.Group("/chinese"),
	}.init()

	tMediaRouter{
		Router: t.Router.Group("/media"),
	}.init()
}
