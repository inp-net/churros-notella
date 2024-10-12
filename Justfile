dev:
	go run server/main.go

build:
	go build -o bin/server server/main.go

updateschema:
	curl -fsSL https://git.inpt.fr/churros/churros/-/raw/main/packages/db/prisma/schema.prisma -o schema.prisma
