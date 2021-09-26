package main

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
	Url "net/url"
	"os"
	"path"
	"time"

	"github.com/gocolly/colly"
)

const (
	blog_list_api          = "https://bizapi.csdn.net/blog-console-api/v1/article/list?page=%d&status=enable&pageSize=20"
	blog_markdown_api      = "https://bizapi.csdn.net/blog-console-api/v3/editor/getArticle?id=%s&model_type="
	blog_size              = 20
	x_ca_key               = "203803574"
	x_ca_signature_headers = "x-ca-key,x-ca-nonce"
)

type Blog struct {
	desc       string
	url        string
	view       int
	id         string
	comment    int
	title      string
	createTime time.Time
}

var blog_counter int = 1
var blog_markdown_counter int = 1
var blog_finished_counter int = 1
var blog_total int = -1
var blog_page int = 1
var blogs []*Blog
var default_header map[string]string = map[string]string{
	"accept":                 "*/*",           //需要指定该头，不然就会报签名错误
	"accept-encoding":        "gzip, deflate", //需要指定该头，不然就会报签名错误
	"x-ca-key":               x_ca_key,
	"x-ca-signature-headers": x_ca_signature_headers,
	"x-ca-nonce":             "",
	"x-ca-signature":         "",
	"cookie":                 `uuid_tt_dd=10_9893907410-1624680270736-675324; _ga=GA1.2.1118997374.1624680273; __gads=ID=7a576859ebe6ada2-22410a550dca005f:T=1625065315:RT=1625065315:S=ALNI_MYPhoDssv7CbcLXRA-zvPxr7pWhaA; ssxmod_itna=QuDtGIqGxjg3i=DXtG7maKiKYYvbP0=+335bQq7Dla=xWKGkD6DWP0WbuN3b83CWWYn3Y3WarTxPLFRQgRaQ3W7Puz8mDAoDhx7QDoPqD1XD0KPo+dkKD3Dm4i3DDxQDgDmKGg8qG36nx0r92PD0xD8Bgpzu2iPpjR8vaHQzxGd/C038C4FPY2Dkb+KsjnvCDhxkAhKQGreVWRDiCRPZ7dDi2DF7YD==; ssxmod_itna2=QuDtGIqGxjg3i=DXtG7maKiKYYvbP0=+335bQqD6EfEWcx0vux03q1icU023cD8Eb64/4fEFPP2K4CiRiQq8Lb2IWhYP7YA8ccMvumOk/WmrhoOGz7K66wKKU2svVZucRS1XuNg7wIkNM/MjniR6/Yj6cbL6+WQUVniqo54R4N0YfFMvX+3=KluQQ=LukWdiViWbFwo=T=Ra7T1iQrQLxx66WNQE+CnX9R37qg63Sf601dlnStZ2m=UIvpFRstAqaD7jx7ihx7SvWSAIqAlIXqX+hYyCnOHxDFqD2WiD; UserName=duandianR; UserInfo=3dbc6260694043afaed303ac9ae236ce; UserToken=3dbc6260694043afaed303ac9ae236ce; UserNick=断桥bian; AU=4F0; UN=duandianR; BT=1625981704006; p_uid=U010000; Hm_up_6bcd52f51e9b3dce32bec4a3997715ac={"islogin":{"value":"1","scope":1},"isonline":{"value":"1","scope":1},"isvip":{"value":"0","scope":1},"uid_":{"value":"duandianR","scope":1}}; Hm_ct_6bcd52f51e9b3dce32bec4a3997715ac=6525*1*10_9893907410-1624680270736-675324!5744*1*duandianR; c_first_ref=www.google.com.hk; c_first_page=https://blog.csdn.net/qq_42956179/article/details/118576680; c_segment=12; dc_sid=323881ccbc1207c5088fc72b1112977d; Hm_lvt_6bcd52f51e9b3dce32bec4a3997715ac=1632410118,1632410150; dc_session_id=10_1632549981861.981546; c_page_id=default; c_hasSub=true; is_advert=1; log_Id_view=319; c_pref=https://blog.csdn.net/qq_42956179/category_11058451.html; c_ref=https://editor.csdn.net/; Hm_lpvt_6bcd52f51e9b3dce32bec4a3997715ac=1632550021; dc_tos=qzz7nq; log_Id_pv=86; log_Id_click=24`,
}
var header *http.Header = &http.Header{}

var transport *http.Transport = &http.Transport{
	TLSClientConfig: &tls.Config{
		InsecureSkipVerify: true,
	},
}
var list_collector *colly.Collector
var markdown_collector *colly.Collector

func init() {
	for key, value := range default_header {
		header.Add(key, value)
	}
	collector := colly.NewCollector()
	collector.WithTransport(transport)
	list_collector = collector.Clone()
	markdown_collector = list_collector.Clone()
	list_collector.OnResponse(parse_blog_list)
	list_collector.OnError(parse_blog_list_error)
	markdown_collector.OnResponse(parse_blog_markdown)
	markdown_collector.OnError(parse_blog_markdown_error)
}

func IntRange(start int, end int, step int) []int {
	var seq_cap int = (end - start) / step
	seq := make([]int, 0, seq_cap)
	for i := 0; i < seq_cap; i++ {
		seq = append(seq, start)
		start = start + step
	}
	return seq
}

func createUuid() string {
	var text string = ""
	var char_list []byte
	const uuid_template string = "xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx"
	seq := append(IntRange(49, 58, 1), IntRange(97, 97+6, 1)...)
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

func getSign(uuid string, url string) string {
	u_url, err := Url.Parse(url)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse url %s (%s)", url, err))
	}
	ekey := []byte("9znpamsyl2c7cdrr9sas0le9vbc3r6ba")
	url_query := u_url.RawQuery
	url_compnent := u_url.Path + "?" + url_query[0:len(url_query)-1]
	url_str := fmt.Sprintf("GET\n*/*\n\n\n\nx-ca-key:203803574\nx-ca-nonce:%s\n%s", uuid, url_compnent)
	to_enc := []byte(url_str)
	hmac_encoder := hmac.New(sha256.New, ekey)
	hmac_encoder.Write(to_enc)
	hmac_digest := hmac_encoder.Sum(nil)
	sign := base64.StdEncoding.EncodeToString(hmac_digest)
	return sign
}

func crawl_blog_markdown(blog *Blog) {
	blog_id := blog.id
	blog_markdown_url := fmt.Sprintf(blog_markdown_api, blog_id)
	uuid := createUuid()
	sign := getSign(uuid, blog_markdown_url)
	context := colly.NewContext()
	context.Put("blog", blog)
	context.Put("counter", blog_markdown_counter)
	header.Set("x-ca-nonce", uuid)
	header.Set("x-ca-signature", sign)
	fmt.Println(fmt.Sprintf("Crawl markdown for blog: %s  ... %d/%d ", blog.title, blog_markdown_counter, blog_total))
	markdown_collector.Request("GET", blog_markdown_url, nil, context, *header)
	blog_markdown_counter += 1
}

func parse_blog_list(resp *colly.Response) {
	user := resp.Ctx.Get("user")
	if resp.StatusCode != 200 {
		tip := fmt.Sprintf("%s's blog list  can't get!", user)
		panic(tip)
	}
	var resp_json map[string]interface{}
	json.Unmarshal(resp.Body, &resp_json)
	total_info, ok := resp_json["data"].(map[string]interface{})["total"]
	if ok && (blog_total == -1) {
		blog_total = int(total_info.(float64))
	}
	if blog_total <= 0 {
		fmt.Println(fmt.Sprintf("No blogs found on user %s.", user))
	}
	blog_list, _ := resp_json["data"].(map[string]interface{})["list"].([]interface{})
	for _, blog := range blog_list {
		vblog := blog.(map[string]interface{})
		zblog := &Blog{}
		id, _ := vblog["articleId"].(float64)
		title, _ := vblog["title"].(string)
		postTime, err := time.Parse("2006-01-02 15:04:05", vblog["postTime"].(string))
		url, _ := vblog["url"].(string)
		fmt.Println(fmt.Sprintf("Crawl blog %s ( %s ) ...  %d/%d", title, url, int(blog_counter), int(blog_total)))
		if err != nil {
			fmt.Printf("Failed to parse blog date (%s)", err)
			return
		}
		zblog.desc = vblog["description"].(string)
		zblog.id = fmt.Sprintf("%d", int(id))
		zblog.url = url
		zblog.title = title
		zblog.comment = int(vblog["commentCount"].(float64))
		zblog.view = int(vblog["viewCount"].(float64))
		zblog.createTime = postTime
		crawl_blog_markdown(zblog)
		blogs = append(blogs, zblog)
		blog_counter += 1
	}
	if blog_counter < int(blog_total) {
		blog_page += 1
		total_blog_page := int(math.Ceil(float64(blog_total) / blog_size))
		blog_list_url := fmt.Sprintf(blog_list_api, blog_page)
		uuid := createUuid()
		sign := getSign(uuid, blog_list_url)
		header.Set("x-ca-nonce", uuid)
		header.Set("x-ca-signature", sign)
		fmt.Println(fmt.Sprintf("Crawl user %s blogs ... at page %d/%d", user, blog_page, total_blog_page))
		resp.Ctx.Put("user", user)
		resp.Ctx.Put("page", string(blog_page))
		list_collector.Request("GET", blog_list_url, nil, resp.Ctx, *header)
	}
}
func parse_blog_list_error(resp *colly.Response, err error) {
	user := resp.Ctx.GetAny("user")
	page := resp.Ctx.GetAny("page")
	fmt.Println(resp.Request.URL)
	fmt.Println(resp.Request.Headers)
	fmt.Println(resp.Headers)
	fmt.Println(fmt.Sprintf("Crawl user %s blogs failed (cause: %s) ... at page %s ", user, err, page))
}
func parse_blog_markdown_error(resp *colly.Response, err error) {
	blog := resp.Ctx.GetAny("blog").(*Blog)
	counter := resp.Ctx.GetAny("counter")
	fmt.Println(fmt.Sprintf("Crawl markdown for blog: %s  ... failed (cause: %s) %d/%d ", blog.title, err, counter, blog_total))
}

func parse_blog_markdown(resp *colly.Response) {
	blog := resp.Ctx.GetAny("blog").(*Blog)
	if resp.StatusCode != 200 {
		blog_finished_counter += 1
		fmt.Printf("Crawl markdown article [%s] failed!\n", blog.title)
		return
	}
	var resp_json map[string]interface{}
	json.Unmarshal(resp.Body, &resp_json)
	status := int(resp_json["code"].(float64))
	if status != 200 {
		fmt.Printf("Crawl markdown article [%s] failed (cause: %s) ... skip %d/%d \n", blog.title, "network error!", blog_finished_counter, blog_total)
		blog_finished_counter += 1
		return
	}
	blog_md := resp_json["data"].(map[string]interface{})["markdowncontent"].(string)
	if len(blog_md) < 1 {
		fmt.Printf("Crawl markdown article [%s] failed (cause: %s)  ... skip  %d/%d  !\n", blog.title, "content is empty!", blog_finished_counter, blog_total)
		blog_finished_counter += 1
		return
	}
	markdown_path := fmt.Sprintf("blogs/[%s]-%s.md", blog.createTime.Format("2006-01-02"), blog.title)
	markdown_dir := path.Dir(markdown_path)
	_, err := os.Stat(markdown_dir)
	if err != nil {
		err = os.MkdirAll(markdown_dir, 0777)
		if err != nil {
			cause := fmt.Sprintf("Failed to create directory  %s (caused by: %s)!", markdown_dir, err)
			fmt.Printf("Crawl markdown article [%s] failed (cause: %s)  ... skip  %d/%d  !\n", blog.title, cause, blog_finished_counter, blog_total)
			blog_finished_counter += 1
			return
		}
	}
	f, err := os.Create(markdown_path)
	if err != nil {
		cause := fmt.Sprintf("Failed to create file %s (caused by: %s)!", markdown_path, err)
		fmt.Printf("Crawl markdown article [%s] failed (cause: %s)  ... skip ( %d/%d ) !\n", blog.title, cause, blog_finished_counter, blog_total)
		blog_finished_counter += 1
		return
	}
	defer f.Close()
	_, err = f.WriteString(blog_md)
	if err != nil {
		cause := fmt.Sprintf("Failed to write file %s (caused by: %s)!", markdown_path, err)
		fmt.Printf("Crawl markdown article [%s] failed (cause: %s)  ... skip ( %d/%d ) !\n", blog.title, cause, blog_finished_counter, blog_total)
		blog_finished_counter += 1
		return
	}
	fmt.Println(fmt.Sprintf("Crawl markdown for blog: %s  ... done %d/%d ", blog.title, blog_finished_counter, blog_total))
	blog_finished_counter += 1
}

func crawl_blog(user string) []*Blog {
	blog_list_url := fmt.Sprintf(blog_list_api, blog_page)
	uuid := createUuid()
	list_header := header.Clone()
	sign := getSign(uuid, blog_list_url)
	list_header.Set("x-ca-nonce", uuid)
	list_header.Set("x-ca-signature", sign)
	context := colly.NewContext()
	context.Put("user", user)
	context.Put("page", "1")
	fmt.Println(fmt.Sprintf("Crawl user %s blogs ... at page %d", user, blog_page))
	list_collector.Request("GET", blog_list_url, nil, context, list_header)
	list_collector.Wait()
	fmt.Println("list_collector exit.")
	markdown_collector.Wait()
	fmt.Println("markdown_collector exit.")
	return blogs

}

func main() {
	blog_user := "duandianR"
	crawl_blog(blog_user)

}
