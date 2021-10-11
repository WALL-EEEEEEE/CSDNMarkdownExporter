GOROOT := $(shell go env GOROOT)

include makefiles/os_detect.mk

run : 
	cd src && go run main.go $(args)

deps-caddy :
ifeq ($(SYSTEM),"OSX")
	@echo "- 检查caddy是否安装 ..." 
	@if brew list caddy &>/dev/null; then \
		echo "- caddy已经安装 ... 跳过"; \
	else  \
		echo "- caddy未安装 ... 尝试安装" && brew install caddy;  \
	fi 
else
	ifeq ($(SYSTEM), "LINUX")
	    $(error 你的系统为 Linux, 暂不支持caddy安装!)
	else
		ifeq ($(SYSTEM), "WIN32")
			$(error 你的系统为 Windows, 暂不支持caddy安装!)
		else
			$(error 你的系统暂不支持！)
		endif
	endif
endif
	


server : build-wasm deps-caddy
	@echo "- 启动静态网站服务 ..."

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