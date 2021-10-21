package spiders

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	. "github.com/duanqiaobb/BlogExporter/pkg"
	. "github.com/duanqiaobb/BlogExporter/pkg/inter"
	"github.com/gocolly/colly"
	log "github.com/sirupsen/logrus"
)

const (
	blog_list_api          = "https://blog.csdn.net/community/home-api/v1/get-business-list?page=%d&size=20&businessType=blog&orderby=&noMore=false&username=%s"
	blog_markdown_api      = "https://bizapi.csdn.net/blog-console-api/v3/editor/getArticle?id=%s&model_type="
	blog_size              = 20
	x_ca_key               = "203803574"
	x_ca_signature_headers = "x-ca-key,x-ca-nonce"
	encrypt_key            = "9znpamsyl2c7cdrr9sas0le9vbc3r6ba"
	uuid_template          = "xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx"
	signature_url_template = "GET\n%s\n\n\n\nx-ca-key:%s\nx-ca-nonce:%s\n%s"
)

var default_header map[string]string = map[string]string{
	"accept":                 "*/*",           //需要指定该头，不然就会报签名错误
	"accept-encoding":        "gzip, deflate", //需要指定该头，不然就会报签名错误
	"x-ca-key":               x_ca_key,
	"x-ca-signature-headers": x_ca_signature_headers,
	"x-ca-nonce":             "",
	"x-ca-signature":         "",
	"cookie":                 "",
}

type CSDNSpider struct {
	blog_counter          int
	blog_markdown_counter int
	blog_finished_counter int
	blog_total            int
	blog_page             int
	user                  string
	outputDir             string
	blogs                 []*Blog
	header                *http.Header
	transport             *http.Transport
	list_collector        *colly.Collector
	markdown_collector    *colly.Collector
	proxy_url             *url.URL
}

func (spider *CSDNSpider) New(spider_args ...interface{}) interface{} {
	var user, cookie, outputDir string = spider_args[0].(string), spider_args[1].(string), spider_args[2].(string)
	header := &http.Header{}
	for key, value := range default_header {
		header.Add(key, value)
	}
	log.Info(cookie)
	header.Set("cookie", cookie)
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	collector := colly.NewCollector()
	collector.WithTransport(transport)
	limitRule := &colly.LimitRule{
		DomainGlob:  "*",
		Delay:       3 * time.Second,
		Parallelism: 1,
	}
	collector.Limit(limitRule)
	list_collector := collector.Clone()
	markdown_collector := collector.Clone()
	spider = &CSDNSpider{
		blog_counter:          1,
		blog_finished_counter: 1,
		blog_markdown_counter: 1,
		blog_total:            -1,
		blog_page:             1,
		user:                  user,
		outputDir:             outputDir,
		blogs:                 []*Blog{},
		header:                header,
	}
	list_collector.OnResponse(spider.parse_blog_list)
	list_collector.OnError(spider.parse_blog_list_error)
	markdown_collector.OnResponse(spider.parse_blog_markdown)
	markdown_collector.OnError(spider.parse_blog_markdown_error)
	spider.list_collector = list_collector
	spider.markdown_collector = markdown_collector
	return spider
}

func (spider *CSDNSpider) intRange(start int, end int, step int) []int {
	var seq_cap int = (end - start) / step
	seq := make([]int, 0, seq_cap)
	for i := 0; i < seq_cap; i++ {
		seq = append(seq, start)
		start = start + step
	}
	return seq
}

func (spider *CSDNSpider) SetProxy(proxy string) {
	proxy_url, err := url.Parse(proxy)
	if err != nil {
		log.Errorf("Proxy url %s is invalid, caused by: %s!", proxy_url, err)
		return
	}
	spider.proxy_url = proxy_url
}

func (spider *CSDNSpider) proxy(request *http.Request) {
	request.Header.Set("X-Caddy-Upstream-Host", request.URL.Hostname())
	if len(request.URL.Port()) > 0 {
		request.Header.Set("X-Caddy-Upstream-Port", ":"+request.URL.Port())
	} else {
		if request.URL.Scheme == "https" {
			request.Header.Set("X-Caddy-Upstream-Port", ":443")
		} else {
			request.Header.Set("X-Caddy-Upstream-Port", ":80")
		}
	}
	request.URL.Host = spider.proxy_url.Host
	log.Infof("Cors代理URL：%s", request.URL.String())
	log.Infof("Header: %+v", request.Header)

}

func (spider *CSDNSpider) createUuid() string {
	var text string = ""
	var char_list []byte
	seq := append(spider.intRange(49, 58, 1), spider.intRange(97, 97+6, 1)...)
	rand.Seed(time.Now().Unix())
	for _, c := range seq {
		char_list = append(char_list, byte(c))
	}
	for _, elem := range uuid_template {
		if elem == rune('4') {
			text += "4"
		} else if elem == rune('-') {
			text += "-"
		} else {
			text += string(char_list[rand.Intn(len(char_list))])
		}
	}
	return text
}

func (spider *CSDNSpider) signRequest(request *http.Request) {
	var sign = spider.getSign(request)
	request.Header.Set("x-ca-signature", sign)
}
func (spider *CSDNSpider) getSign(request *http.Request) string {
	uuid := request.Header.Get("x-ca-nonce")
	if len(uuid) == 0 {
		uuid = spider.createUuid()
		request.Header.Set("x-ca-nonce", uuid)
	}
	//fmt.Println("uuid:", uuid)
	accept := request.Header.Get("accept")
	url := request.URL
	url_query := ""
	var url_query_keys []string
	var url_query_map map[string][]string = url.Query()
	for key, _ := range url.Query() {
		url_query_keys = append(url_query_keys, key)
	}
	sort.Strings(url_query_keys)
	for _, key := range url_query_keys {
		params := ""
		values := url_query_map[key]
		for _, value := range values {
			if value != "" {
				params = key + "=" + value
			} else {
				params = key + value
			}
			if url_query == "" {
				url_query += params
			} else {
				url_query += "&" + params
			}

		}
	}
	//fmt.Println("url_query:", url_query)
	url_compnent := url.Path + "?" + url_query
	//fmt.Println("url_component:", url_compnent)
	url_str := fmt.Sprintf(signature_url_template, accept, x_ca_key, uuid, url_compnent)
	//fmt.Println("url_str:\n===================================================\n" + url_str + "\n===================================================")
	to_enc := []byte(url_str)
	ekey := []byte(encrypt_key)
	hmac_encoder := hmac.New(sha256.New, ekey)
	hmac_encoder.Write(to_enc)
	hmac_digest := hmac_encoder.Sum(nil)
	sign := base64.StdEncoding.EncodeToString(hmac_digest)
	//fmt.Println("signature: ", sign)
	return sign
}

func (spider *CSDNSpider) crawl_blog_markdown(blog *Blog) {
	blog_id := blog.Id
	blog_markdown_url := fmt.Sprintf(blog_markdown_api, blog_id)
	var blog_markdown_req, err = http.NewRequest("GET", blog_markdown_url, nil)
	if err != nil {
		log.Infof("Crawl markdown for blog: %s  ... %d/%d failed! (cause: %s)", blog.Title, spider.blog_markdown_counter, spider.blog_total, err)
		return
	}
	blog_markdown_req.Header = spider.header.Clone()
	spider.signRequest(blog_markdown_req)
	ctx := colly.NewContext()
	ctx.Put("blog", blog)
	ctx.Put("counter", spider.blog_markdown_counter)
	log.Infof("Crawl markdown for blog: %s  ... %d/%d ", blog.Title, spider.blog_markdown_counter, spider.blog_total)
	log.Info(blog_markdown_req.Header)
	spider.proxy(blog_markdown_req)
	spider.markdown_collector.Request(blog_markdown_req.Method, blog_markdown_req.URL.String(), nil, ctx, blog_markdown_req.Header)
	spider.blog_markdown_counter += 1
}

func (spider *CSDNSpider) parse_blog_list(resp *colly.Response) {
	user := resp.Ctx.Get("user")
	log.Info("Request Url: ", resp.Request.URL)
	log.Info("Request Header: ", resp.Request.Headers)
	log.Info("Response Header: ", resp.Headers)
	log.Info("Response Body: ", resp.Body)
	if resp.StatusCode != 200 || len(resp.Body) < 1 {
		log.Panicf("%s's blog list  can't get!", user)
	}
	var resp_json map[string]interface{}
	json.Unmarshal(resp.Body, &resp_json)
	log.Info(resp_json)
	total_info, ok := resp_json["data"].(map[string]interface{})["total"]
	if ok && (spider.blog_total == -1) {
		spider.blog_total = int(total_info.(float64))
	}
	if spider.blog_total <= 0 {
		log.Infof("No blogs found on user %s.", user)
	}
	blog_list, _ := resp_json["data"].(map[string]interface{})["list"].([]interface{})
	for _, blog := range blog_list {
		vblog := blog.(map[string]interface{})
		zblog := &Blog{}
		id, _ := vblog["articleId"].(float64)
		title, _ := vblog["title"].(string)
		blog_type := int(vblog["type"].(float64))
		if blog_type == 2 {
			log.Warningf("Skip crawl markdown for blog: %s  ... %d/%d, it's a reprinted blog.", title, spider.blog_counter, spider.blog_total)
			spider.blog_markdown_counter++
			spider.blog_counter++
			continue
		}
		postTime, err := time.Parse("2006-01-02 15:04:05", vblog["postTime"].(string))
		url, _ := vblog["url"].(string)
		log.Infof("Crawl blog %s ( %s ) ...  %d/%d", title, url, int(spider.blog_counter), int(spider.blog_total))
		if err != nil {
			log.Warningf("Failed to parse blog date (%s)", err)
			return
		}
		zblog.Desc = vblog["description"].(string)
		zblog.Id = fmt.Sprintf("%d", int(id))
		zblog.Url = url
		zblog.Title = title
		zblog.Comment = int(vblog["commentCount"].(float64))
		zblog.View = int(vblog["viewCount"].(float64))
		zblog.CreateTime = postTime
		spider.crawl_blog_markdown(zblog)
		spider.blogs = append(spider.blogs, zblog)
		spider.blog_counter += 1
	}
	if spider.blog_counter < int(spider.blog_total) {
		spider.blog_page += 1
		total_blog_page := int(math.Ceil(float64(spider.blog_total) / blog_size))
		blog_list_url := fmt.Sprintf(blog_list_api, spider.blog_page, user)
		blog_list_req, err := http.NewRequest("GET", blog_list_url, nil)
		spider.proxy(blog_list_req)
		if err != nil {
			log.Infof("Crawl user %s blogs ... at page %d/%d failed! (cause: %s)", user, spider.blog_page, total_blog_page, err)
			return
		}
		log.Infof("Crawl user %s blogs ... at page %d/%d", user, spider.blog_page, total_blog_page)
		spider.list_collector.Request(blog_list_req.Method, blog_list_req.URL.String(), blog_list_req.Body, resp.Ctx, blog_list_req.Header)
	}
}
func (spider *CSDNSpider) parse_blog_list_error(resp *colly.Response, err error) {
	user := resp.Ctx.GetAny("user")
	page := resp.Ctx.GetAny("page")
	x_ca_error_messsage := ""
	if resp != nil && resp.Headers != nil {
		x_ca_error_messsage = resp.Headers.Get("x-ca-error-message")
	}
	log.WithFields(
		log.Fields{
			"x-ca-error-message": x_ca_error_messsage,
		}).Errorf("Crawl user %s blogs failed (cause: %s) ... at page %s ", user, err, page)
}
func (spider *CSDNSpider) parse_blog_markdown_error(resp *colly.Response, err error) {
	blog := resp.Ctx.GetAny("blog").(*Blog)
	counter := resp.Ctx.GetAny("counter")
	x_ca_error_messsage := ""
	if resp != nil && resp.Headers != nil {
		x_ca_error_messsage = resp.Headers.Get("x-ca-error-message")
	}
	log.WithFields(
		log.Fields{
			"x-ca-error-message": x_ca_error_messsage,
		}).Errorf("Crawl markdown for blog: %s  ... failed (cause: %s) %d/%d ", blog.Title, err, counter, spider.blog_total)
}

func (spider *CSDNSpider) parse_blog_markdown(resp *colly.Response) {
	blog := resp.Ctx.GetAny("blog").(*Blog)
	if resp.StatusCode != 200 {
		spider.blog_finished_counter += 1
		log.WithFields(
			log.Fields{
				"x-ca-error-message": resp.Headers.Get("x-ca-error-message"),
			}).Errorf("Crawl markdown article [%s] failed!\n", blog.Title)
		return
	}
	var resp_json map[string]interface{}
	json.Unmarshal(resp.Body, &resp_json)
	status := int(resp_json["code"].(float64))
	if status != 200 {
		log.WithFields(
			log.Fields{
				"x-ca-error-message": resp.Headers.Get("x-ca-error-message"),
			}).Errorf("Crawl markdown article [%s] failed (cause: %s) ... skip %d/%d \n", blog.Title, "network error!", spider.blog_finished_counter, spider.blog_total)
		spider.blog_finished_counter += 1
		return
	}
	blog_md := resp_json["data"].(map[string]interface{})["markdowncontent"].(string)
	if len(blog_md) < 1 {
		log.Warnf("Crawl markdown article [%s] failed (cause: %s)  ... skip  %d/%d  !\n", blog.Title, "content is empty!", spider.blog_finished_counter, spider.blog_total)
		spider.blog_finished_counter += 1
		return
	}
	markdown_filename := fmt.Sprintf("[%s]-%s.md", blog.CreateTime.Format("2006-01-02"), strings.Replace(blog.Title, "/", "-", -1))
	markdown_path := path.Join(spider.outputDir, markdown_filename)
	markdown_dir := path.Dir(markdown_path)
	_, err := os.Stat(markdown_dir)
	if err != nil {
		err = os.MkdirAll(markdown_dir, 0777)
		if err != nil {
			cause := fmt.Sprintf("Failed to create directory  %s (caused by: %s)!", markdown_dir, err)
			log.Errorf("Crawl markdown article [%s] failed (cause: %s)  ... skip  %d/%d  !\n", blog.Title, cause, spider.blog_finished_counter, spider.blog_total)
			spider.blog_finished_counter += 1
			return
		}
	}
	f, err := os.Create(markdown_path)
	if err != nil {
		cause := fmt.Sprintf("Failed to create file %s (caused by: %s)!", markdown_path, err)
		log.Errorf("Crawl markdown article [%s] failed (cause: %s)  ... skip ( %d/%d ) !\n", blog.Title, cause, spider.blog_finished_counter, spider.blog_total)
		spider.blog_finished_counter += 1
		return
	}
	defer f.Close()
	_, err = f.WriteString(blog_md)
	if err != nil {
		cause := fmt.Sprintf("Failed to write file %s (caused by: %s)!", markdown_path, err)
		log.Errorf("Crawl markdown article [%s] failed (cause: %s)  ... skip ( %d/%d ) !\n", blog.Title, cause, spider.blog_finished_counter, spider.blog_total)
		spider.blog_finished_counter += 1
		return
	}
	log.Infof("Crawl markdown for blog: %s  ... done %d/%d ", blog.Title, spider.blog_finished_counter, spider.blog_total)
	spider.blog_finished_counter += 1
}

func (spider *CSDNSpider) Crawl() {
	blog_list_url := fmt.Sprintf(blog_list_api, spider.blog_page, spider.user)
	context := colly.NewContext()
	context.Put("user", spider.user)
	context.Put("page", "1")
	var blog_list_req, err = http.NewRequest("GET", blog_list_url, nil)
	if err != nil {
		log.Infof("Crawl user %s blogs ... at page %d failed (cause: %s)!", spider.user, spider.blog_page, err)
		return
	}
	log.Infof("Crawl user %s blogs ... at page %d", spider.user, spider.blog_page)
	spider.proxy(blog_list_req)
	spider.list_collector.Request(blog_list_req.Method, blog_list_req.URL.String(), blog_list_req.Body, context, blog_list_req.Header)
	spider.list_collector.Wait()
	spider.markdown_collector.Wait()
}

func init() {
	RegisterSpider("CSDN", (*CSDNSpider)(nil))
}
