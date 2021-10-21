set_languages("go")
set_project("BlogExporter")
set_xmakever("2.5.1")
set_version("0.1", { build = "%Y%m%d%H%M" })

if is_os("macosx") then
  add_requires("brew::caddy", { alias = "caddy", optional = true })
end

if is_os("linux") then
  add_requires("apt::caddy", { alias = "caddy", optional = true })
end

target("console")
set_kind("binary")
add_files("main.go")
set_targetdir(path.join(os.projectdir(), "target"))
set_filename("BlogExporter")

on_build(function(target)
  -- 导入配置模块
  os.runv("go", { "build", "-o", target:targetfile() })
end)

on_run(function(target)
  cprintf("运行 %s ... \n", target:targetfile())
  local args = os.getenv("ARGS")
  if (not args)
  then
    args = ""
  end
  args = args:split(" ")
  os.execv(target:targetfile(), args)
end)

target("wasm")
set_kind("binary")
add_files("main.go")
on_build(function(target)
  -- 导入配置模块
  import("core.project.config")
  local wasm_path      = vformat("$(projectdir)/%s", "web/wasm/main.wasm")
  local go_wasm_js     = "$(shell go env GOROOT)/misc/wasm/wasm_exec.js"
  local target_wasm_js = "$(projectdir)/web/js/wasm_exec.js"
  os.setenv("GOOS", "js")
  os.setenv("GOARCH", "wasm")
  cprint("构建 wasm ...")
  os.runv("go", { "build", "-o", wasm_path })
  cprint("准备 Go语言的wasm的执行环境 ...")
  os.cp(go_wasm_js, target_wasm_js)
end)

target("server")
add_deps("wasm")
on_run(function(target)
  import("core.project.project")
  local web_dir = vformat("$(projectdir)/%s", "web")
  local caddy_conf = vformat("$(projectdir)/%s", "config/Caddyfile")
  cprint("运行 %s wasm.app ...", project:name())
  os.execv("sudo", { "caddy", "run", "--config", caddy_conf })
end)
