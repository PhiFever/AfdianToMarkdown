package shop

import (
	"AfdianToMarkdown/afdian"
	"AfdianToMarkdown/config"
	"AfdianToMarkdown/utils"
	"context"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/carlmjohnson/requests"
	"golang.org/x/exp/slog"
)

const (
	shopDir = "shop"
)

// GetShopProducts 获取作者电铺的所有商品
func GetShopProducts(cfg *config.Config, authorUrlSlug string, cookieString string, tagId string, quickUpdate bool) error {
	authorHost, _ := url.JoinPath(cfg.HostUrl, "a", authorUrlSlug, "?tab=shop")
	//创建存储文件夹
	subDir := tagId
	if subDir == "" {
		subDir = "default"
	}
	saveDir := path.Join(cfg.DataDir, authorUrlSlug, shopDir, subDir)

	if err := os.MkdirAll(saveDir, os.ModePerm); err != nil {
		return fmt.Errorf("create shop dir error: %v", err)
	}
	slog.Info("开始获取电铺内容", "authorHostUrl", authorHost, "tagId", tagId, "saveDir", saveDir)

	products, err := afdian.GetProductList(cfg, authorUrlSlug, cookieString, tagId)
	if err != nil {
		return err
	}

	converter := md.NewConverter("", true, nil)
	for _, product := range products {
		timePrefix := product.UpdateTime.Format("2006-01-02_15_04_05")
		fileName := fmt.Sprintf("%s_%s.md", timePrefix, product.Name)
		filePath := path.Join(saveDir, fileName)

		if _, err := os.Stat(filePath); err == nil {
			if quickUpdate {
				slog.Info("Quick update: 检测到已存在文件，停止更新电铺", "product", product.Name)
				return nil
			}
			slog.Info("商品已存在，跳过", "product", product.Name)
			continue
		}

		slog.Info("正在保存商品", "name", product.Name, "price", product.Price)

		// 转换描述为 Markdown
		markdownDesc, err := converter.ConvertString(product.Desc)
		if err != nil {
			slog.Error("转换 Markdown 失败", "err", err, "product", product.Name)
			markdownDesc = product.Desc // 回退到原始 HTML
		}

		// 处理封面图片
		picContent := ""
		if product.Pic != "" {
			assetsDir := filepath.Join(saveDir, utils.ImgDir)
			_ = os.MkdirAll(assetsDir, os.ModePerm)

			ext := filepath.Ext(product.Pic)
			if ext == "" {
				ext = ".jpg"
			}
			imgName := fmt.Sprintf("%s_cover%s", product.ID, ext)
			imgPath := filepath.Join(assetsDir, imgName)

			err := requests.
				URL(product.Pic).
				Header("user-agent", afdian.ChromeUserAgent).
				ToFile(imgPath).
				Fetch(context.Background())

			if err == nil {
				picContent = fmt.Sprintf("![cover](%s)\n\n", filepath.Join(utils.ImgDir, imgName))
			} else {
				picContent = fmt.Sprintf("![cover](%s)\n\n", product.Pic)
			}
		}

		tagDisplay := product.TagID
		if tagDisplay == "" {
			tagDisplay = "全部"
		}

		content := fmt.Sprintf("## %s\n\n**价格**: %s\n\n**标签**: %s\n\n**链接**: %s\n\n**更新时间**: %s\n\n---\n\n%s%s",
			product.Name, product.Price, tagDisplay, product.Url, product.UpdateTime.Format("2006-01-02 15:04:05"), picContent, markdownDesc)

		if err := os.WriteFile(filePath, []byte(content), os.ModePerm); err != nil {
			slog.Error("保存商品文件失败", "err", err, "path", filePath)
		}

		time.Sleep(time.Millisecond * time.Duration(afdian.DelayMs))
	}

	slog.Info("电铺内容获取完成", "count", len(products))
	return nil
}
