package motion

import (
	"AfdianToMarkdown/afdian"
	"AfdianToMarkdown/utils"
	"fmt"
	"log"
	"net/url"
	"os"
	"path"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/spf13/cast"
)

const (
	authorDir = "motions"
)

// GetMotions 获取作者的所有作品
func GetMotions(authorUrlSlug string, cookieString string, authToken string, disableComment bool) error {
	authorHost, _ := url.JoinPath(afdian.HostUrl, "a", authorUrlSlug)
	//创建作者文件夹
	if err := os.MkdirAll(path.Join(authorUrlSlug, authorDir), os.ModePerm); err != nil {
		return fmt.Errorf("create author dir error: %v", err)
	}
	log.Println("authorHost:", authorHost)

	//获取作者作品列表
	prevPublishSn := ""
	var postList []afdian.Post
	for {
		//获取作者作品列表
		subArticleList, publishSn := afdian.GetMotionUrlList(authorUrlSlug, cookieString, prevPublishSn)
		postList = append(postList, subArticleList...)
		prevPublishSn = publishSn
		if publishSn == "" {
			break
		}
		time.Sleep(time.Millisecond * time.Duration(30))
	}
	//log.Println("postList:", utils.ToJSON(postList))
	log.Println("postList length:", len(postList))

	converter := md.NewConverter("", true, nil)
	for i, article := range postList {
		filePath := path.Join(utils.GetExecutionPath(), authorUrlSlug, authorDir, cast.ToString(i)+"_"+article.Name+".md")
		if err := afdian.SavePostIfNotExist(filePath, article, authToken, disableComment, converter); err != nil {
			return err
		}
	}
	return nil
}
