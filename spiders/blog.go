package main

import (
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/gocolly/colly"
)

const (
	blog_api          = "https://blog.csdn.net/community/home-api/v1/get-business-list?page=%d&size=20&businessType=blog&orderby=&noMore=false&username=%s"
	blog_markdown_api = "https://bizapi.csdn.net/blog-console-api/v3/editor/getArticle?id=%d&model_type="
	blog_size         = 20
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
func createUuid() {
    text = ""
    char_list = []
    for c in range(97,97+6):
        char_list.append(chr(c))
    for c in range(49,58):
        char_list.append(chr(c))
    for i in "xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx":
        if i == "4":
            text += "4"
        elif i == "-":
            text += "-"
        else:
            text += random.choice(char_list)
    return text
}

func get_sign(uuid,url):
    s = urlparse(url)
    ekey = "9znpamsyl2c7cdrr9sas0le9vbc3r6ba".encode()
    to_enc = f"GET\n*/*\n\n\n\nx-ca-key:203803574\nx-ca-nonce:{uuid}\n{s.path+'?'+s.query[:-1]}".encode()
    sign = b64encode(hmac.new(ekey, to_enc, digestmod=hashlib.sha256).digest()).decode()
    return sign

func crawl_blog_markdown(article_id string) {
}

func crawl_blog(user string) []*Blog {
	var blog_counter int = 1
	blog_url := fmt.Sprintf(blog_api, blog_counter, user)
	var blogs []*Blog
	c := colly.NewCollector()
	c.OnResponse(func(r *colly.Response) {
		if r.StatusCode != 200 {
			tip := fmt.Sprintf("%s's blog list  can't get!", user)
			panic(tip)
		}
		var resp_json map[string]interface{}
		json.Unmarshal(r.Body, &resp_json)
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
			c.Visit(blog_url)
		}
	})
	c.Visit(blog_url)
	c.Wait()
	return blogs
}

func main() {
	blog_user := "duandianR"
	crawl_blog(blog_user)

}
