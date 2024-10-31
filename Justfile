current_version := `git describe --tags --abbrev=0 | cut -c 2-`

dev:
	docker compose up -d
	go mod tidy
	go fmt
	go run server/*.go

build:
	just genprisma
	go mod tidy
	go build -v -ldflags="-X main.Version={{current_version}}" -o bin/server server/*.go

docker:
	docker build -t notella:{{current_version}} .
	docker tag notella:{{current_version}} notella:latest
	docker tag notella:{{current_version}} harbor.k8s.inpt.fr/net7/notella:{{current_version}}
	docker tag notella:{{current_version}} harbor.k8s.inpt.fr/net7/notella:latest
	docker push notella:{{current_version}}
	docker push notella:latest
	docker push harbor.k8s.inpt.fr/net7/notella:{{current_version}}
	docker push harbor.k8s.inpt.fr/net7/notella:latest


install:
	just build
	cp bin/server ~/.local/bin/notella

updateschema:
	curl -fsSL https://git.inpt.fr/churros/churros/-/raw/main/packages/db/prisma/schema.prisma -o schema.prisma
	sed -i '/^generator .* {/,/^}/d' schema.prisma
	sed -i '1i\
	generator goprisma {\n\
	provider        = "go run github.com/steebchen/prisma-client-go"\n\
	previewFeatures = ["fullTextSearch", "postgresqlExtensions"]\n\
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
