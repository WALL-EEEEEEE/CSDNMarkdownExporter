package inter

type Spider interface {
	Crawl()
	New(spider_args ...interface{}) interface{}
	SetProxy(proxy string)
}
