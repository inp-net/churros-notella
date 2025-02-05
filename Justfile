current_version := `git describe --tags --abbrev=0 | cut -c 2-`

dev:
	docker compose up -d
	go mod tidy
	go fmt
	go run server/*.go

build output="bin/server":
	echo "Building version {{current_version}}"
	just genprisma
	go mod tidy
	go build -v -ldflags="-X main.Version={{current_version}}" -o {{output}} server/*.go


install:
	just build
	cp bin/server ~/.local/bin/notella

updateschema url="https://git.inpt.fr/churros/churros/-/raw/main/packages/db/prisma/schema.prisma":
	curl -fsSL {{url}} -o schema.prisma
	sed -i '/^generator .* {/,/^}/d' schema.prisma
	sed -i '1i\
	generator goprisma {\n\
	provider        = "go run github.com/steebchen/prisma-client-go"\n\
	previewFeatures = ["fullTextSearchPostgres", "postgresqlExtensions"]\n\
	}\
	' schema.prisma
	go run github.com/steebchen/prisma-client-go format


genprisma:
    go get github.com/steebchen/prisma-client-go 
    go run github.com/steebchen/prisma-client-go generate

gen_typescript:
	go run scripts/typing.go 

generate:
	just updateschema
	just gen_typescript

release_typescript:
	just gen_typescript
	git add *.ts 
	git commit -m "chore: update typescript types"
	npm version minor
	npm publish --access=public
	git push 
	git push --tags
