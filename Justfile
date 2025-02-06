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
	provider		= "go run github.com/steebchen/prisma-client-go"\n\
	previewFeatures = ["fullTextSearchPostgres", "postgresqlExtensions"]\n\
	}\
	' schema.prisma
	go run github.com/steebchen/prisma-client-go format


genprisma:
	go get github.com/steebchen/prisma-client-go 
	go run github.com/steebchen/prisma-client-go generate

gen_typescript:
	bash scripts/sync-event-enum.sh
	go run scripts/typing.go 
	node typescript-dist/index.js 
	just gen_typescript_lib_txt

gen_typescript_lib_txt:
	#!/bin/bash
	mkdir -p testarea; cd testarea
	npm init -y
	jq '.type = "module"' < package.json > package.json.new
	mv package.json.new package.json
	cp ../typescript-dist/index.js notella.js
	echo "import * as notella from './notella.js'; console.log(notella)" > index.js
	node index.js > lib.txt
	cp lib.txt ../typescript/lib.txt
	cd ..
	rm -rf testarea

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
