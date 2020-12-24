package api

import "github.com/gin-gonic/gin"

type tAPIRouter struct {
	Router *gin.RouterGroup
}

func (t tAPIRouter) init() {
	tChineseRouter{
		Router: t.Router.Group("/chinese"),
	}.init()

	tMediaRouter{
		Router: t.Router.Group("/media"),
	}.init()
}
