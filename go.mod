module github.com/zhquiz/go-server

go 1.15

require (
	cloud.google.com/go/firestore v1.4.0 // indirect
	cloud.google.com/go/storage v1.12.0 // indirect
	firebase.google.com/go v3.13.0+incompatible
	github.com/gin-contrib/cache v1.1.0
	github.com/gin-contrib/sessions v0.0.3
	github.com/gin-gonic/contrib v0.0.0-20201101042839-6a891bf89f19
	github.com/gin-gonic/gin v1.6.3
	github.com/go-playground/validator/v10 v10.4.1
	github.com/joho/godotenv v1.3.0
	github.com/mattn/go-sqlite3 v1.14.5
	github.com/tebeka/atexit v0.3.0
	github.com/yanyiwu/gojieba v1.1.2
	github.com/zserge/lorca v0.1.9
	golang.org/x/oauth2 v0.0.0-20201208152858-08078c50e5b5 // indirect
	gopkg.in/sakura-internet/go-rison.v3 v3.1.0
	gorm.io/driver/sqlite v1.1.4
	gorm.io/gorm v1.20.8
)

replace github.com/yanyiwu/gojieba v1.1.2 => github.com/ttys3/gojieba v1.1.3
