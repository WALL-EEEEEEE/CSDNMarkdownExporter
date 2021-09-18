package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	Url "net/url"
	"regexp"
	"time"

	"github.com/gocolly/colly"
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

const (
	blog_api          = "https://blog.csdn.net/community/home-api/v1/get-business-list?page=%d&size=20&businessType=blog&orderby=&noMore=false&username=%s"
	blog_markdown_api = "https://bizapi.csdn.net/blog-console-api/v3/editor/getArticle?id=%d&model_type="
	blog_size         = 20
)

var blog_counter int = 1
var blogs []*Blog
var user string

var collector *colly.Collector = colly.NewCollector()

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
	url_compnent := u_url.Path + "?" + url_query[0:len(url_query)]
	fmt.Println(url_compnent)
	to_enc := []byte(fmt.Sprintf("GET\n*/*\n\n\n\nx-ca-key:203803574\nx-ca-nonce:%s\n%s", uuid, url_compnent))
	hmac_encoder := hmac.New(sha256.New, ekey)
	hmac_encoder.Write(to_enc)
	sign := base64.StdEncoding.EncodeToString([]byte(hex.EncodeToString(hmac_encoder.Sum(nil))))
	return sign
}

func crawl_blog_markdown(article_id string) {
}

func parse_blog_list(resp *colly.Response) {
	if resp.StatusCode != 200 {
		tip := fmt.Sprintf("%s's blog list  can't get!", user)
		panic(tip)
	}
	var resp_json map[string]interface{}
	json.Unmarshal(resp.Body, &resp_json)
	blog_total, _ := resp_json["data"].(map[string]interface{})["total"].(float64)
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
		blogs = append(blogs, zblog)
		blog_counter += 1
	}
	if blog_counter < int(blog_total) {
		blog_page := int(math.Ceil(float64(blog_counter) / blog_size))
		total_blog_page := int(math.Ceil(float64(blog_total) / blog_size))
		blog_url := fmt.Sprintf(blog_api, blog_page, user)
		fmt.Println(fmt.Sprintf("Crawl user %s blogs ... at page %d/%d", user, blog_page, total_blog_page))
		collector.Visit(blog_url)
	}
}

func parse_blog_markdown(resp *colly.Response) {

}

func parseResponse(resp *colly.Response) error {
	blog_md_reg, blog_md_reg_err := regexp.Compile(`https?:\/\/blog\.csdn\.net\/community\/home-api\/v1\/get-business-list\?.*`)
	blog_list_reg, blog_list_reg_err := regexp.Compile(`https?:\/\/bizapi\.csdn\.net\/blog-console-api\/v3\/editor\/getArticle\?.*`)
	url := resp.Request.URL.String()
	if blog_md_reg_err != nil {
		return fmt.Errorf("Failed to compilre regex of blog markdown url regex.")
	} else if blog_list_reg_err != nil {
		return fmt.Errorf("Failed to compilre regex of blog list url regex.")
	}
	switch url = url; {
	case blog_list_reg.MatchString(url):
		parse_blog_list(resp)
	case blog_md_reg.MatchString(url):
		parse_blog_markdown(resp)
	}
	return nil
}

func crawl_blog(user string) []*Blog {
	var blog_counter int = 1
	blog_url := fmt.Sprintf(blog_api, blog_counter, user)
	var blogs []*Blog
	collector.Visit(blog_url)
	collector.Wait()
	return blogs
}

func main() {
	//blog_user := "duandianR"
	//crawl_blog(blog_user)
	//
	//collector.OnResponse(parseResponse)
	url := fmt.Sprintf("https://bizapi.csdn.net/blog-console-api/v3/editor/getArticle?id=%s&model_type=", "80341748")
	var headers map[string]string = map[string]string{
		"x-ca-key":               "203803574",
		"x-ca-signature-headers": "x-ca-key,x-ca-nonce",
		"x-ca-nonce":             "",
		"x-ca-signature":         "",
		"Cookies":                `uuid_tt_dd=10_9893908260-1631422578939-282460; _ga=GA1.2.892729514.1631631538; c_first_ref=www.google.com.hk; c_first_page=https://blog.csdn.net/Fei20140908/article/details/114849593; c_segment=12; dc_sid=8d2579a401f7d91942159c1334ac3b08; is_advert=1; _gid=GA1.2.655835131.1631807398; Hm_lvt_6bcd52f51e9b3dce32bec4a3997715ac=1631631538,1631632180,1631807398; dc_session_id=10_1631889878007.754241; unlogin_scroll_step=1631889884892; SESSION=ba973ce6-d269-450c-b554-e6fb4127515e; ssxmod_itna=eqRxnDBDcDRD97DyDGxBpOW5t0Qqe4xBAwFehKDsq3aDpxBKidDaxQapC+PuewReerbxmwIteFguifo=sRL4GLDmKQtQUWxiiDC40rD74irDDxD3Db4KDSCxG=DjWz6MyvxYPGWAqitfDQyODiPD=2yHZxi7DD5DArPDwx7OtQ0eTWKDerasUeokCAgDqWKD9xYDsEifORAfjt2C833ERnqDUWGU3GL4qe2DSpRHDlF2DC9kUi5Qr+d8cxvza0DTFHxNYWx5E0D0/Y44xEr4eejq+sedKAD5tnGKOdVeDDcpOCOPrYD=; ssxmod_itna2=eqRxnDBDcDRD97DyDGxBpOW5t0Qqe4xBAwFexikAqK33trDlZODjR0Wn6merlW=Gs+xApxyUEB5RxRD4Q+w+5OWTxhxTL2gIW+accWka2K6zqr2fYbfzM=OIq5v8y20HbIziM42kFYeBidGU8ou0odGQwhZiCiuYRIRf4rvfrbM+WN6Y+L=tmuYvlbdNCbEFmuPVQaQNbjPh6Wf7Bi7viEv8kHsDf3KQ7iVK7ByKYd+Y7dABioKy277Z6KRQU3tj1HmcZcfp2KeZgYhKQCtoytUQ6RwW7+pt3f2Nwhn/6=7Lo6S7OI8FKA3ezD4E+vilRr7Tjtqq=YiFALz7qpYxL2ehaP542mbVUe0Wes7iyDTlfI9743L+xAKaCpiQizUx3EbQoKl05qjP+UrYPbYHacj7K=oqE+nIohfhqiGQbvtu9b/a4gqtixDKwPtB30q5MixdZTK7LeAL0fPh1I2nLsBTT7OYGxwqF4SKX1I1ehVnxMD5uYwM1I1PhG1D3BTCZX1E05z5LPhlZxlpqTY53zqr9xrzDQDDLxD20od+0dQLTi9QnPqAQV/GDD==; acw_sc__v2=6144aa7a2c1a201ae02ccf8558621979f165528e; UserName=duandianR; UserInfo=c8ec4cf1eaba4441b605b8dd54d7ec1e; UserToken=c8ec4cf1eaba4441b605b8dd54d7ec1e; UserNick=断桥bian; AU=4F0; UN=duandianR; BT=1631890051748; p_uid=U010000; c_page_id=default; Hm_up_6bcd52f51e9b3dce32bec4a3997715ac={"islogin":{"value":"1","scope":1},"isonline":{"value":"1","scope":1},"isvip":{"value":"0","scope":1},"uid_":{"value":"duandianR","scope":1}}; Hm_ct_6bcd52f51e9b3dce32bec4a3997715ac=6525*1*10_9893908260-1631422578939-282460!5744*1*duandianR; log_Id_view=16; log_Id_click=3; c_pref=https://blog.csdn.net/qq_35524157/article/details/117385786; c_ref=https://mp.csdn.net/mp_blog/manage/article?spm=1001.2101.3001.5448; dc_tos=qzl2v0; log_Id_pv=12; Hm_lpvt_6bcd52f51e9b3dce32bec4a3997715ac=1631890622`,
	}
	uuid := createUuid()
	headers["x-ca-nonce"] = uuid
	sign := getSign(uuid, url)
	fmt.Println(sign)
	/*
		headers["x-ca-signature"] = sign
		client := &http.Client{}
		fmt.Println(url)
		req, _ := http.NewRequest("GET", url, nil)
		for key, value := range headers {
			req.Header.Add(key, value)
		}
		fmt.Println(req.Header)
		resp, err := client.Do(req)
		fmt.Println(resp.ContentLength)
		fmt.Println(err)
	*/
}
