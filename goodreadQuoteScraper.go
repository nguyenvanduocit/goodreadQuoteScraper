package goodreadQuoteScraper

import (
	"net/http"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"strconv"
	"github.com/gin-gonic/gin/json"
	"io/ioutil"
	"regexp"
)

type Quote struct {
	Author string `json:"author"`
	AuthorAvatar string `json:"author_avatar"`
	Content string `json:"content"`
	Tags []string `json:"tags"`
}

type Crawler struct {
	BaseUrl string
	Quotes []*Quote `json:"quotes"`
}

func (c *Crawler)fetchDocument(url string)(*goquery.Document, error){
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("status code error: %d %s", res.StatusCode, res.Status))
	}
	return goquery.NewDocumentFromReader(res.Body)
}

func (c *Crawler)worker(id int, jobs <-chan int, results chan<- []*Quote) {
	for j := range jobs {
		log.Println(fmt.Sprintf("Worker %d is take page %d", id, j))
		results <- c.getQuoteInPage(j)
	}
}

func (c *Crawler)Crawl(){
	totalPage := c.getTotalPage()
	jobs := make(chan int, totalPage)
	results := make(chan []*Quote, totalPage)

	for workerIndex := 1; workerIndex <= 3; workerIndex++ {
		go c.worker(workerIndex, jobs, results)
	}

	for j := 1; j <= totalPage; j++ {
		jobs <- j
	}
	close(jobs)

	var quotes []*Quote
	for a := 1; a <= totalPage; a++ {
		quotes = append(quotes, <-results...)
	}
	bQuotes, _ := json.Marshal(quotes)
	err := ioutil.WriteFile("./quote.json", bQuotes, 0644)
	if err != nil {
		log.Fatalln(err)
	}

}

func (c *Crawler)getTotalPage()int{
	doc, err := c.fetchDocument(c.BaseUrl)
	if err != nil {
		log.Fatalln(err)
	}
	docPageList := doc.Find(".next_page").Parent().Find("a")
	docLastPage := docPageList.Eq(docPageList.Length() - 2)
	totalPage, err := strconv.Atoi(docLastPage.Text())
	if err != nil {
		log.Fatalln(err)
	}
	return totalPage
}

func (c *Crawler)getQuoteInPage(pageIndex int)([]*Quote){
	doc, err := c.fetchDocument(fmt.Sprintf("%s?page=%d", c.BaseUrl, pageIndex))
	if err != nil {
		log.Fatalln(err)
	}
	var quotes []*Quote

	quoteDiv := doc.Find("div.quote")
	quoteDiv.Each(func(i int, selection *goquery.Selection) {
		quote, err := c.parseQuote(selection)
		if err != nil {
			log.Println(err)
		} else {
			quotes = append(quotes, quote)
		}
	})
	return quotes
}

func (c *Crawler)parseQuote(selection *goquery.Selection)(*Quote, error){
	quote := &Quote{
		Author: selection.Find(".authorOrTitle").Text(),
		Tags: []string{},
	}
	avatarUrl, ok := selection.Find(".leftAlignedImage img").Attr("src")
	if ok {
		quote.AuthorAvatar = avatarUrl
	}

	var re = regexp.MustCompile(`(?m)“([^”]+)”`)
	quote.Content = re.FindStringSubmatch(selection.Find(".quoteText").Text())[1]

	selection.Find(".quoteFooter .greyText a").Each(func(i int, selection *goquery.Selection) {
		quote.Tags = append(quote.Tags, selection.Text())
	})
	return quote, nil
}