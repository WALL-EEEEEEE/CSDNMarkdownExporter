package main

import (
	"net/http"
	"net/url"
	. "net/url"

	"github.com/gocolly/colly"
	log "github.com/sirupsen/logrus"
)

func main() {
	collector := colly.NewCollector()
	url, _ := url.Parse("https://blog.csdn.net/community/home-api/v1/get-business-list?page=1&size=20&businessType=blog&orderby=&noMore=false&username=duandianR")

	collector.OnResponse(func(r *colly.Response) {
		log.Info("Response Status: ", r.StatusCode)
		log.Info("Response Content: ", string(r.Body))
	})
	/*
		collector.OnRequest(func(r *colly.Request) {
			var proxy_port string
			if len(r.URL.Port()) > 0 {
				proxy_port = r.URL.Port()
			} else {
				if r.URL.Scheme == "https" {
					proxy_port = "443"
				} else {
					proxy_port = "80"
				}
			}
			r.Headers.Set("User-Agent", "colly - https://github.com/gocolly/colly")
			r.Headers.Set("X-Caddy-Upstream-Host", r.URL.Host)
			r.Headers.Set("X-Caddy-Upstream-Port", ":"+proxy_port)
			proxy_url, err := Parse("https://localhost")
			r.URL.Host = proxy_url.Host
			if err != nil {
				panic(err)
			}
			log.Info("Request URL: ", r.URL)
			log.Info("Request Headers: ", r.Headers)
		})
	*/
	var proxy_port string
	if len(url.Port()) > 0 {
		proxy_port = url.Port()
	} else {
		if url.Scheme == "https" {
			proxy_port = "443"
		} else {
			proxy_port = "80"
		}
	}
	header := &http.Header{}
	header.Set("User-Agent", "colly - https://github.com/gocolly/colly")
	header.Set("X-Caddy-Upstream-Host", url.Host)
	header.Set("X-Caddy-Upstream-Port", ":"+proxy_port)
	proxy_url, _ := Parse("https://localhost")
	url.Host = proxy_url.Host
	collector.Request("GET", url.String(), nil, nil, *header)
	collector.Wait()

}
