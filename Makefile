build:
	go build -o go-server.app
desktop:
	make build
	./go-server.app
server:
	make build
	ZHQUIZ_DESKTOP=1 ./go-server.app
dev:
	export PORT=3000
	# increase the file watch limit, might be required on MacOS
	ulimit -n 1000
	reflex -s -r '\.go$$' make server
