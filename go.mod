module github.com/zhquiz/go-zhquiz

go 1.15

require (
	github.com/PuerkitoBio/goquery v1.6.0
	github.com/gen2brain/dlgs v0.0.0-20201118155338-03fe7f81ad25
	github.com/getlantern/systray v1.1.0
	github.com/gin-contrib/sessions v0.0.3
	github.com/gin-gonic/contrib v0.0.0-20201101042839-6a891bf89f19
	github.com/gin-gonic/gin v1.6.3
	github.com/go-playground/validator/v10 v10.4.1
	github.com/golang/protobuf v1.4.3 // indirect
	github.com/google/go-cmp v0.5.4 // indirect
	github.com/gopherjs/gopherjs v0.0.0-20200217142428-fce0ec30dd00 // indirect
	github.com/gorilla/sessions v1.2.0 // indirect
	github.com/joho/godotenv v1.3.0
	github.com/kr/pretty v0.1.0 // indirect
	github.com/mattn/go-sqlite3 v2.0.3+incompatible // indirect
	github.com/oklog/ulid/v2 v2.0.2
	github.com/stretchr/testify v1.5.1 // indirect
	github.com/wangbin/jiebago v0.3.2
	golang.org/x/net v0.0.0-20201031054903-ff519b6c9102 // indirect
	golang.org/x/sys v0.0.0-20201201145000-ef89a241ccb3 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
	gopkg.in/sakura-internet/go-rison.v3 v3.1.0
	gopkg.in/square/go-jose.v2 v2.5.1
	gorm.io/driver/sqlite v1.1.4
	gorm.io/gorm v1.20.8
)

replace github.com/mattn/go-sqlite3 v2.0.3+incompatible => github.com/mattn/go-sqlite3 v1.14.6
