current_version := `git describe --tags --abbrev=0 | cut -c 2-`

dev:
	docker compose up -d
	go mod tidy
	go run server/*.go

build:
	go mod tidy
	go build -v -ldflags="-X main.Version={{current_version}}" -o bin/server server/*.go

install:
	just build
	cp bin/server ~/.local/bin/notella

updateschema:
	curl -fsSL https://git.inpt.fr/churros/churros/-/raw/main/packages/db/prisma/schema.prisma -o schema.prisma

genprisma:
    go get github.com/steebchen/prisma-client-go 
    go run github.com/steebchen/prisma-client-go generate

generate:
	just updateschema
	just updateopenapi
