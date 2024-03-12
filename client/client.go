package client

import (
	"fmt"
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/extensions"
	"log"
	"math/rand"
	"time"
)

const (
	DelayMs = 330
	Host    = `https://afdian.net/`
)

func TrueRandFloat(min, max float64) float64 {
	// 使用当前时间作为种子值
	seed := time.Now().Unix()
	source := rand.NewSource(seed)
	randomGenerator := rand.New(source)

	// 生成范围在 [min, max) 内的随机浮点数
	randomFloat := min + randomGenerator.Float64()*(max-min)
	return randomFloat
}

func TrueRandInt(min, max int) int {
	// 使用当前时间作为种子值
	seed := time.Now().Unix()
	source := rand.NewSource(seed)
	randomGenerator := rand.New(source)

	// 生成范围在 [min, max) 内的随机整数
	randomInt := min + randomGenerator.Intn(max-min)
	return randomInt
}

func InitBaseCollector() *colly.Collector {
	c := colly.NewCollector(
		colly.AllowedDomains(Host),
	)
	//https://go-colly.org/docs/best_practices/extensions/
	extensions.RandomUserAgent(c)
	//https://github.com/gocolly/colly/blob/v1.2.0/extensions/referer.go#L10
	//extensions.Referer(c)

	//设置超时时间
	c.SetRequestTimeout(30 * time.Second)

	//cookies := `_ga=GA1.1.1610556398.1702557124; auth_token=a2f5931cb036de871664e0f0df9991ec_20231214203204; _ga_6STWKR7T9E=GS1.1.1702557123.1.1.1702557220.28.0.0`

	c.OnRequest(func(r *colly.Request) {
		//r.Headers.Set("Cookie", cookies)
		r.Headers.Set("Referer", Host)
		log.Println("Visiting", r.URL)
		log.Println("UserAgent:", r.Headers.Get("User-Agent"))
		log.Println("Referer:", r.Headers.Get("Referer"))
		log.Println("Cookies", c.Cookies(r.URL.String()))
	})

	c.OnError(func(_ *colly.Response, err error) {
		log.Println("Something went wrong:", err)
	})

	c.OnResponse(func(r *colly.Response) {
		log.Println("Visited", r.Request.URL)
		fmt.Println(string(r.Body))
	})
	return c
}
