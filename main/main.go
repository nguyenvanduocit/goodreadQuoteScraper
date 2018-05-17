package main

import "github.com/nguyenvanduocit/goodreadQuoteScraper"

func main(){
	crawl := &goodreadQuoteScraper.Crawler{
		BaseUrl: "https://www.goodreads.com/quotes",
	}
	crawl.Crawl()
}
