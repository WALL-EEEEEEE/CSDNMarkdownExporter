package main

import (
	"fmt"
	"io"
	"net/http"
)

func main() {
	client := &http.Client{}

	url := "https://localhost/community/home-api/v1/get-business-list?page=1&size=20&businessType=blog&orderby=&noMore=false&username=duandianR"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Add("User-Agent", "colly - https://github.com/gocolly/colly")
	req.Header.Add("X-Caddy-Upstream-Host", "blog.csdn.net")
	req.Header.Add("X-Caddy-Upstream-Port", ":443")
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	fmt.Println(resp.Status)
	fmt.Println(resp.ContentLength)
	content, err := io.ReadAll(resp.Body)
	fmt.Println(string(content))

}
