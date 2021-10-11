const go = new Go();
let  vue;
let  blog_exporter;

init_blog_exporter()

window.addEventListener('load', function () {
    vue = new Vue({
        el:"#app",
        vuetify: new Vuetify(),
        methods: {
            callBlogExporter: function() {
                run_blog_exporter(this.site, this.params)
            }
        },
        data: {
            params: {
                "cookie": "",
                "user": "",
                "output": "",
            },
            site: "CSDN",
            log: false,
        }
    })

})

function init_blog_exporter() {
    WebAssembly.instantiateStreaming(fetch("wasm/main.wasm"), go.importObject).then((result) => {
            blog_exporter = result.instance
    });   
}

function run_blog_exporter(site, site_params) {
    params = ["run", "-s", site]
    //site_params['cookie'] = encodeURIComponent(site_params["cookie"])
    for (const param_key in site_params) {
        if (param_key=="cookie") {
            param = "--"+param_key+"=\""+site_params[param_key]+"\""
        } else {
            param = "--"+param_key+"="+site_params[param_key]
        }
        params.push(param)
    }
    go.argv = ["BlogExporter"].concat(params)//["BlogExporter","list"] //params
    console.log(go.argv)
    go.run(blog_exporter)
}