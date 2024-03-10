package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/spf13/cast"
	"io"
	"log"
	"os"
	"time"
)

const ChromeUserAgent = `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36`

// Cookie 以下是使用chromedp的相关代码
// Cookie 从 Chrome 中使用EditThisCookie导出的 Cookies
type Cookie struct {
	Domain     string  `json:"domain"`
	Expiration float64 `json:"expirationDate"`
	HostOnly   bool    `json:"hostOnly"`
	HTTPOnly   bool    `json:"httpOnly"`
	Name       string  `json:"name"`
	Path       string  `json:"path"`
	SameSite   string  `json:"sameSite"`
	Secure     bool    `json:"secure"`
	Session    bool    `json:"session"`
	StoreID    string  `json:"storeId"`
	Value      string  `json:"value"`
}

// ReadCookiesFromFile 从文件中读取 Cookies
func ReadCookiesFromFile(filePath string) []Cookie {
	var cookies []Cookie

	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(file)

	data, err := io.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(data, &cookies)
	if err != nil {
		log.Fatal(err)
	}

	return cookies
}

// ConvertCookies 将从文件中读取的 Cookies 转换为 chromedp 需要的格式
func ConvertCookies(cookies []Cookie) []*network.CookieParam {
	cookieParams := make([]*network.CookieParam, len(cookies))
	for i, cookie := range cookies {
		cookieParams[i] = &network.CookieParam{
			Name:     cookie.Name,
			Value:    cookie.Value,
			Domain:   cookie.Domain,
			Path:     cookie.Path,
			Secure:   cookie.Secure,
			HTTPOnly: cookie.HTTPOnly,
		}
	}
	return cookieParams
}

// InitChromedpContext 实际在每次调用时可以派生一个新的超时context，然后在这个新的context中执行任务，可以避免卡住
// timeoutCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
// defer cancel()
func InitChromedpContext(imageEnabled bool) (context.Context, context.CancelFunc) {
	log.Println("正在初始化 Chromedp 上下文")
	// 设置Chrome启动参数
	// chromedp默认使用的是本机上chrome的UA，但是会因为headerless而被识别为爬虫
	// 所以需要覆写UA，注意与本机上chrome的UA保持一致，否则可能会过不了一些网站的人机验证
	options := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", !cast.ToBool(DebugMode)),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("–disable-plugins", true),
		chromedp.Flag("blink-settings", "imagesEnabled="+fmt.Sprintf("%t", imageEnabled)),
		chromedp.UserAgent(ChromeUserAgent),
	)
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), options...)

	chromeCtx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	return chromeCtx, cancel
}

// GetScrolledRenderedPage 获取需要整个页面滚动到底部后经过JavaScript渲染的页面
// FIXME:也许应该加个chromedp.WaitVisible(scrollSelector, chromedp.ByQuery),，等待页面加载完毕
func GetScrolledRenderedPage(ctx context.Context, cookieParams []*network.CookieParam, url string) []byte {
	log.Println("正在渲染页面:", url)

	var htmlContent string
	// 具体任务放在这里
	var tasks = chromedp.Tasks{
		network.SetCookies(cookieParams),
		chromedp.Navigate(url),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var prevScrollHeight, currentScrollHeight int64
			for {
				// 执行JS脚本获取当前文档的滚动高度
				err := chromedp.Evaluate(`document.body.scrollHeight`, &prevScrollHeight).Do(ctx)
				if err != nil {
					return err
				}

				// 执行JS脚本向下滚动页面
				err = chromedp.Evaluate(`window.scrollTo(0, document.body.scrollHeight);`, nil).Do(ctx)
				if err != nil {
					return err
				}

				// 等待一段时间，以便页面加载
				time.Sleep(2 * time.Second)

				// 再次获取滚动高度以判断是否到底
				err = chromedp.Evaluate(`document.body.scrollHeight`, &currentScrollHeight).Do(ctx)
				if err != nil {
					return err
				}

				// 如果滚动高度没有变化，则表示已经到达页面底部
				if prevScrollHeight == currentScrollHeight {
					break
				}
			}

			return nil
		}),
		chromedp.OuterHTML("html", &htmlContent),
	}
	//开始执行任务
	err := chromedp.Run(ctx, tasks)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("渲染完毕", url)
	return []byte(htmlContent)
}

// GetHtmlDoc 从[]byte中读取html内容，返回goquery.Document
func GetHtmlDoc(htmlContent []byte) *goquery.Document {
	// 将 []byte 转换为 io.Reader
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(htmlContent))
	if err != nil {
		log.Fatal(err)
	}
	return doc
}

// ReadHtmlDoc 从文件中读取html内容，返回goquery.Document
func ReadHtmlDoc(filePath string) *goquery.Document {
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(file)

	doc, err := goquery.NewDocumentFromReader(file)
	if err != nil {
		log.Fatal(err)
	}
	return doc
}
