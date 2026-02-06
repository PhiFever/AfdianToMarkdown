package storage

import (
	"AfdianToMarkdown/afdian"
	"AfdianToMarkdown/config"
	"AfdianToMarkdown/utils"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/carlmjohnson/requests"
	"golang.org/x/exp/slog"
)

// SavePostIfNotExist 检查文件是否存在，不存在则下载并保存文章
func SavePostIfNotExist(cfg *config.Config, filePath string, article afdian.Post, authToken string, disableComment bool, converter *md.Converter) error {
	_, err := os.Stat(filePath)
	fileExists := err == nil || os.IsExist(err)
	if !fileExists {
		slog.Info("Saving file:", "path", filePath)
		content, err := afdian.GetPostContent(cfg, article.Url, authToken, converter)
		if err != nil {
			return err
		}
		//TODO:不支持图文混排
		picContent, err := getPictures(filePath, article)
		if err != nil {
			return err
		}

		referUrl := strings.Replace(article.Url, "post", "p", 1)
		articleContent := fmt.Sprintf("## %s\n\n### Refer\n\n%s\n\n### 正文\n\n%s\n\n%s",
			article.Name, referUrl, content, picContent)

		if !disableComment {
			commentString, hotCommentString, err := afdian.GetPostComment(cfg, article.Url, authToken)
			if err != nil {
				return err
			}
			articleContent = fmt.Sprintf("%s\n\n%s\n\n%s", articleContent, hotCommentString, commentString)
		}

		if err := os.WriteFile(filePath, []byte(articleContent), os.ModePerm); err != nil {
			return err
		}
	} else {
		log.Printf("File exists: %s", filePath)
	}
	return nil
}

func getPictures(filePath string, article afdian.Post) (string, error) {
	if len(article.Pictures) == 0 {
		return "", nil
	}
	assetsDir := filepath.Join(filepath.Dir(filePath), utils.ImgDir)
	if err := os.MkdirAll(assetsDir, os.ModePerm); err != nil {
		return "", fmt.Errorf("create assets directory error: %v", err)
	}
	picContent := ""
	// 下载并保存图片到本地
	for i, pictureUrl := range article.Pictures {
		// 生成本地图片文件名
		ext := filepath.Ext(pictureUrl)
		if ext == "" {
			ext = ".jpg" // 默认扩展名
		}
		localFileName := fmt.Sprintf("%s_%d%s", utils.ToSafeFilename(article.Name), i, ext)
		localFilePath := filepath.Join(assetsDir, localFileName)

		log.Printf("Downloading picture in article %s: %s", article.Name, pictureUrl)
		// 使用requests下载图片
		err := requests.
			URL(pictureUrl).
			Header("user-agent", afdian.ChromeUserAgent).
			ToFile(localFilePath).
			Fetch(context.Background())

		if err != nil {
			log.Printf("Failed to download image %s: %v", pictureUrl, err)
			// 如果下载失败，使用原始URL
			picContent += fmt.Sprintf("![image](%s)\n", pictureUrl)
			continue
		}

		// 使用相对路径引用本地图片
		relPath := filepath.Join(utils.ImgDir, localFileName)
		picContent += fmt.Sprintf("![image](%s)\n", relPath)
	}
	return picContent, nil
}
