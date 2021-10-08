template-test :	build-template-test
	@echo "- 启动 template-test ..."
	go run src/server.go --dir ../static

build-template-test	: ../static/wasm/template_test/main.wasm  ../static/js/wasm_exec.js

test/static/js/wasm_exec.js : 
	@echo "- 复制 wasm_exec.js 文件 ..."
	@echo /usr/local/go
	cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" ../static/js

test/static/wasm/template_test/main.wasm : *.go
	@echo "- 构建 template 测试的wasm文件 ..."
	GOOS=js GOARCH=wasm go build  -o ../static/wasm/template_test.wasm  main.go


