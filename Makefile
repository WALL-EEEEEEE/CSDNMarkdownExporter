GOROOT := $(shell go env GOROOT)

run : 
	cd src && go run main.go $(args)

server : build-wasm
	@echo "- 启动静态网站服务 ..."
	go run src/server.go --dir static

build-wasm : static/wasm/main.wasm static/js/wasm_exec.js

static/wasm/main.wasm : src/*.go
	@echo "- 构建 wasm文件 ..."
	cd src && GOOS=js GOARCH=wasm go build  -o  ../static/wasm/main.wasm  main.go

static/js/wasm_exec.js : 
	@echo "- 复制 wasm_exec.js 文件 ..."
	cp "$(GOROOT)/misc/wasm/wasm_exec.js" static/js

clean : 
	-rm static/wasm/main.wasm static/js/wasm_exec.js

include test/src/template_test/template.mk