current_version := `git describe --tags --abbrev=0 | cut -c 2-`

dev:
	go mod tidy
	go run server/*.go

build:
	go mod tidy
	go build -ldflags="-X notella.Version={{current_version}}" -o bin/server server/*.go

updateschema:
	curl -fsSL https://git.inpt.fr/churros/churros/-/raw/main/packages/db/prisma/schema.prisma -o schema.prisma

updateopenapi:
	go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=openapi-config.yaml openapi.yaml

generate:
	just updateschema
	just updateopenapi
