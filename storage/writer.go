package storage

import (
	"AfdianToMarkdown/afdian"
	"AfdianToMarkdown/config"
	"AfdianToMarkdown/utils"
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/carlmjohnson/requests"
	"golang.org/x/exp/slog"
)

// SavePostIfNotExist 检查文件是否存在，不存在则下载并保存文章
func SavePostIfNotExist(cfg *config.Config, filePath string, article afdian.Post, authToken string, disableComment bool, converter *md.Converter) (skipped bool, err error) {
	_, err = os.Stat(filePath)
	fileExists := err == nil || os.IsExist(err)
	if !fileExists {
		slog.Info("Saving file:", "path", filePath)
		content, audio, video, err := afdian.GetPostContent(cfg, article.Url, authToken, converter)
		if err != nil {
			return false, err
		}
		//TODO:不支持图文混排
		picContent, err := getPictures(filePath, article)
		if err != nil {
			return false, err
		}

		audioContent := renderMedia(cfg, filePath, article.Name, audio, "audio")
		videoContent := renderMedia(cfg, filePath, article.Name, video, "video")
		mediaContent := ""
		if audioContent != "" || videoContent != "" {
			mediaContent = "### 媒体\n\n" + audioContent + videoContent
		}

		referUrl := strings.Replace(article.Url, "post", "p", 1)
		articleContent := fmt.Sprintf("## %s\n\n### Refer\n\n%s\n\n### 正文\n\n%s\n\n%s\n\n%s",
			article.Name, referUrl, content, picContent, mediaContent)

		if !disableComment {
			commentString, hotCommentString, err := afdian.GetPostComment(cfg, article.Url, authToken)
			if err != nil {
				return false, err
			}
			articleContent = fmt.Sprintf("%s\n\n%s\n\n%s", articleContent, hotCommentString, commentString)
		}

		if err := os.WriteFile(filePath, []byte(articleContent), os.ModePerm); err != nil {
			return false, err
		}
	} else {
		slog.Info("File exists:", "path", filePath)
		return true, nil
	}
	return false, nil
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
		ext := filepath.Ext(strings.SplitN(pictureUrl, "?", 2)[0])
		if ext == "" {
			ext = ".jpg" // 默认扩展名
		}
		localFileName := fmt.Sprintf("%s_%d%s", utils.ToSafeFilename(article.Name), i, ext)
		localFilePath := filepath.Join(assetsDir, localFileName)

		slog.Info("Downloading picture in:", "article", article.Name, "url", pictureUrl)
		// 使用requests下载图片
		err := requests.
			URL(pictureUrl).
			Header("user-agent", afdian.ChromeUserAgent).
			ToFile(localFilePath).
			Fetch(context.Background())

		if err != nil {
			slog.Error("Failed to download image:", "url", pictureUrl, "error", err)
			// 如果下载失败，使用原始URL
			picContent += fmt.Sprintf("![image](%s)\n", pictureUrl)
			continue
		}

		// 使用相对路径引用本地图片
		relPath := path.Join(utils.ImgDir, localFileName)
		picContent += fmt.Sprintf("![image](%s)\n", relPath)
	}
	return picContent, nil
}

// renderMedia 生成媒体标签：cfg.DownloadMedia 启用时尝试下载到本地，未启用或下载失败时回退到远程 URL
func renderMedia(cfg *config.Config, filePath, articleName, mediaUrl, label string) string {
	if mediaUrl == "" {
		return ""
	}

	src := mediaUrl
	if cfg.DownloadMedia {
		if local, err := downloadMediaFile(filePath, articleName, mediaUrl, label); err != nil {
			slog.Error("Failed to download media", "label", label, "url", mediaUrl, "error", err)
		} else {
			src = local
		}
		afdian.MediaDownloadDelay()
	}

	return fmt.Sprintf("<%s controls src=\"%s\"></%s>\n\n", label, src, label)
}

func downloadMediaFile(filePath, articleName, mediaUrl, label string) (string, error) {
	assetsDir := filepath.Join(filepath.Dir(filePath), utils.ImgDir)
	if err := os.MkdirAll(assetsDir, os.ModePerm); err != nil {
		return "", fmt.Errorf("create assets directory error: %v", err)
	}

	ext := filepath.Ext(strings.SplitN(mediaUrl, "?", 2)[0])
	if ext == "" {
		if label == "audio" {
			ext = ".mp3"
		} else {
			ext = ".mp4"
		}
	}
	localFileName := fmt.Sprintf("%s_%s%s", utils.ToSafeFilename(articleName), label, ext)
	localFilePath := filepath.Join(assetsDir, localFileName)

	slog.Info("Downloading media", "label", label, "article", articleName, "url", mediaUrl)
	if err := requests.
		URL(mediaUrl).
		Header("user-agent", afdian.ChromeUserAgent).
		ToFile(localFilePath).
		Fetch(context.Background()); err != nil {
		return "", err
	}

	return path.Join(utils.ImgDir, localFileName), nil
}
