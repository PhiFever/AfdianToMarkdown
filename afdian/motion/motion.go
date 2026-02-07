package motion

import (
	"AfdianToMarkdown/afdian"
	"AfdianToMarkdown/config"
	"AfdianToMarkdown/storage"
	"fmt"
	"net/url"
	"os"
	"path"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"golang.org/x/exp/slog"
)

const (
	authorDir = "motions"
)

// GetMotions 获取作者的所有作品
func GetMotions(cfg *config.Config, authorUrlSlug string, cookieString string, authToken string, disableComment bool, quickUpdate bool) error {
	authorHost, _ := url.JoinPath(cfg.HostUrl, "a", authorUrlSlug)
	//创建作者文件夹
	if err := os.MkdirAll(path.Join(cfg.DataDir, authorUrlSlug, authorDir), os.ModePerm); err != nil {
		return fmt.Errorf("create author dir error: %v", err)
	}
	slog.Info("作者主页", "authorHostUrl", authorHost)

	//获取作者作品列表，边获取边下载
	converter := md.NewConverter("", true, nil)
	prevPublishSn := ""
	totalCount := 0
	for {
		subArticleList, publishSn, err := afdian.GetMotionUrlList(cfg, authorUrlSlug, cookieString, prevPublishSn)
		if err != nil {
			return err
		}

		for _, article := range subArticleList {
			timePrefix := article.PublishTime.Format("2006-01-02_15_04_05")
			filePath := path.Join(cfg.DataDir, authorUrlSlug, authorDir, timePrefix+"_"+article.Name+".md")
			skipped, err := storage.SavePostIfNotExist(cfg, filePath, article, authToken, disableComment, converter)
			if err != nil {
				return err
			}
			if quickUpdate && skipped {
				slog.Info("Quick update: 检测到已存在文件，跳过剩余动态", "author", authorUrlSlug)
				return nil
			}
		}

		totalCount += len(subArticleList)
		prevPublishSn = publishSn
		if publishSn == "" {
			break
		}
		time.Sleep(time.Millisecond * time.Duration(30))
	}
	slog.Info("postList length:", "count", totalCount)
	return nil
}
