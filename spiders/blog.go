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
	URL "net/url"
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
	u_url, err := URL.Parse(url)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse url %s (%s)", url, err))
	}
	ekey := []byte("9znpamsyl2c7cdrr9sas0le9vbc3r6ba")
	url_compnent := u_url.Path + "?" + u_url.RawQuery
	fmt.Println(url_compnent)
	to_enc := []byte(fmt.Sprintf("GET\n*/*\n\n\n\nx-ca-key:203803574\nx-ca-nonce:%s\n%s", uuid, url_compnent))
	hmac_encoder := hmac.New(sha256.New, ekey)
	hmac_encoder.Write(to_enc)
	sign := base64.StdEncoding.EncodeToString([]byte(hex.EncodeToString(hmac_encoder.Sum(nil))))
	return sign
}

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
	//blog_user := "duandianR"
	//crawl_blog(blog_user)
	url := fmt.Sprintf("https://bizapi.csdn.net/blog-console-api/v3/editor/getArticle?id=%s&model_type=", "80341748")
	var headers map[string]string = map[string]string{
		"x-ca-key":               "203803574",
		"x-ca-signature-headers": "x-ca-key,x-ca-nonce",
		"x-ca-nonce":             "",
		"x-ca-signature":         "",
		"Cookies":                "uuid_tt_dd=10_17437689420-1631256337980-379222; _ga=GA1.2.1512928437.1631256339; c_segment=6; dc_sid=1d2eeb271d79ce4df7f3d5d60d965d2b; __gads=ID=86303851e4c317da-220fe49495cb004d:T=1631274921:RT=1631274921:S=ALNI_MawIU-8Q9PrA1j7DED107s72QPnWA; _gid=GA1.2.805216307.1631678588; UN=duandianR; p_uid=U010000; Hm_lvt_e5ef47b9f471504959267fd614d579cd=1631759515,1631759540; Hm_ct_6bcd52f51e9b3dce32bec4a3997715ac=6525*1*10_17437689420-1631256337980-379222!5744*1*duandianR; Hm_lpvt_e5ef47b9f471504959267fd614d579cd=1631759547; c_first_ref=www.google.com.hk; SESSION=c08358b9-f15f-43c2-a3df-b133467d8a2d; UserName=duandianR; UserInfo=773258e6c6db4f95ab9a214b1bb4d220; UserToken=773258e6c6db4f95ab9a214b1bb4d220; UserNick=%E6%96%AD%E6%A1%A5bian; AU=4F0; BT=1631790356838; ssxmod_itna=eui=Dv4UxjxhCDzxAOe5Q0QqcqDTFhHuWeGQQpqDlOexWKGkD6DWP0WruT5zvzIe1i23rO/hEcGQTFORTRoP8SOpNTrDCPGn+pxMhYD44GTDt4DTD34DYDixib2DiydDjxGP9RLky=DEDYP9DDoDY+=uDitD4qDBDD66D7QDIw==lG=qyb2HICUQQmD9D0U3xBL4iaT5T1aNnuiro7KTWDDHz4yxl0qCN0mz4CPuDB=wxBQMzOX7IeyyBMUDNEoT3gf0R0otDShqSoQP09xAeh4AK74P87wd+05dKDDAS1zs3iD=; ssxmod_itna2=eui=Dv4UxjxhCDzxAOe5Q0QqcqDTFhHuWeGQQTD6ER2D0yGo503IPOnOhIXznqA1BPGFQxWANmqoKly0sVKm2wbEOj4eLWhTWznmysgRnwdKpxhfdgXwA1TM2H804IkzCIR4bix6o/AAoAgExdjU+3DrIFni+Yl2P=3o+REE3oeTIUW6+t8TPEoWY5ZYI77WNo8RTNmnWsWuC8m2Na6RtErh+RgvgjnWg3RE1oKHttKu0W4Tt3YKCvL7nlcBw/ZPj=9ShnnTLvdbw1CB75zCzVM5zNsGft5Iqj6lhIH3B6Zxm8sEHbcTMMYZuaVhFXBNxrClqhjOhjMsjOYkktgYY07CanKqM=CuFZooPkY+Q=Df3Pue43tOGHChx0GdTGH7R=+2snnKz7D5T3iqvT7NY+OBCAmevoqvP0v+WPD7QTG4MYPTBxxRD0ec7n5f8cwlDXltG7+=Dmqmfojq63FRtDu4eNnw0R541wWaGXf9g8w4m3l3HNDH1ADlmhr95DseF82xQ=E0xq/aDoY4Dm5LwYVU5DGcDG7axeuD8K059O44D===; Hm_up_6bcd52f51e9b3dce32bec4a3997715ac=%7B%22islogin%22%3A%7B%22value%22%3A%221%22%2C%22scope%22%3A1%7D%2C%22isonline%22%3A%7B%22value%22%3A%221%22%2C%22scope%22%3A1%7D%2C%22isvip%22%3A%7B%22value%22%3A%220%22%2C%22scope%22%3A1%7D%2C%22uid_%22%3A%7B%22value%22%3A%22duandianR%22%2C%22scope%22%3A1%7D%7D; log_Id_click=29; c_first_page=https%3A//blog.csdn.net/cbmljs/article/details/84991453; Hm_lvt_6bcd52f51e9b3dce32bec4a3997715ac=1631790331,1631791457,1631791514,1631864548; Hm_lpvt_6bcd52f51e9b3dce32bec4a3997715ac=1631864548; c_hasSub=true; log_Id_view=100; dc_session_id=10_1631866566150.407857; c_pref=https%3A//www.google.com.hk/; c_ref=https%3A//blog.csdn.net/duandianR%3Fspm%3D1000.2115.3001.5343; c_page_id=default; dc_tos=qzkkau; log_Id_pv=54",
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
