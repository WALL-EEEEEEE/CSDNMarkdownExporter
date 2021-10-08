
static/wasm/main.wasm : src/*.go
	@echo "- 构建 wasm文件 ..."
	cd src && GOOS=js GOARCH=wasm go build  -o ../static/wasm/main.wasm  main.go

test/static/wasm/template_test/main.wasm : test/src/template_test/*.go
	@echo "- 构建 template 测试的wasm文件 ..."
	cd test/src/ && GOOS=js GOARCH=wasm go build  -o ../static/wasm/template_test.wasm  template_test/main.go

static/js/wasm_exec.js : 
	@echo "- 复制 wasm_exec.js 文件 ..."
	cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" static/js

test/static/js/wasm_exec.js : 
	@echo "- 复制 wasm_exec.js 文件 ..."
	@echo /usr/local/go
	cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" test/static/js



build-wasm	:	static/wasm/main.wasm static/js/wasm_exec.js

build-template-test	: test/static/wasm/template_test/main.wasm  test/static/js/wasm_exec.js

template-test :	build-template-test
	@echo "- 启动 template-test ..."
	go run src/server.go --dir test/static


server	:	build-wasm
	@echo "- 启动静态网站服务 ..."
	go run src/server.go --dir src/static

clean : 
	rm static/wasm/main.wasm