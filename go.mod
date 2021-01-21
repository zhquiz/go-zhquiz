module github.com/zhquiz/go-zhquiz

go 1.15

require (
	github.com/PuerkitoBio/goquery v1.6.0
	github.com/getlantern/systray v1.1.0
	github.com/gin-gonic/contrib v0.0.0-20201101042839-6a891bf89f19
	github.com/gin-gonic/gin v1.6.3
	github.com/go-playground/validator/v10 v10.4.1
	github.com/golang/protobuf v1.4.3 // indirect
	github.com/google/go-cmp v0.5.4 // indirect
	github.com/jkomyno/nanoid v0.0.0-20170914145641-30c81465692e
	github.com/joho/godotenv v1.3.0
	github.com/mattn/go-sqlite3 v2.0.3+incompatible // indirect
	github.com/ncruces/zenity v0.5.2
	github.com/stretchr/testify v1.5.1 // indirect
	github.com/wangbin/jiebago v0.3.2
	github.com/zserge/lorca v0.1.9
	golang.org/x/net v0.0.0-20201031054903-ff519b6c9102 // indirect
	golang.org/x/sys v0.0.0-20201201145000-ef89a241ccb3 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gorm.io/driver/sqlite v1.1.4
	gorm.io/gorm v1.20.8
)

replace github.com/mattn/go-sqlite3 v2.0.3+incompatible => github.com/mattn/go-sqlite3 v1.14.6
