package album

import (
	"AfdianToMarkdown/afdian"
	"AfdianToMarkdown/config"
	"AfdianToMarkdown/storage"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"golang.org/x/exp/slog"
)

func GetAlbums(cfg *config.Config, authorUrlSlug string, cookieString string, authToken string, disableComment bool, quickUpdate bool) error {
	albumHost, _ := url.JoinPath(cfg.HostUrl, "a", authorUrlSlug, "album")
	slog.Info("album列表页:", "albumHostUrl", albumHost)
	userId, err := afdian.GetAuthorId(cfg, authorUrlSlug, albumHost, cookieString)
	if err != nil {
		return err
	}
	albumList, err := afdian.GetAlbumList(cfg, userId, albumHost, cookieString)
	if err != nil {
		return err
	}
	converter := md.NewConverter("", true, nil)
	for _, album := range albumList {
		slog.Info("Find album: ", "albumName", album.AlbumName)
		err := GetAlbum(cfg, cookieString, authToken, album, disableComment, quickUpdate, converter, "", "")
		if err != nil {
			return err
		}

	}
	return nil
}

func GetAlbum(cfg *config.Config, cookieString string, authToken string, album afdian.Album, disableComment bool, quickUpdate bool, converter *md.Converter, fromDate string, toDate string) error {
	// 解析时间范围
	var fromTime, toTime time.Time
	var hasFrom, hasTo bool
	if fromDate != "" {
		t, err := time.Parse("2006-01-02", fromDate)
		if err != nil {
			return fmt.Errorf("--from 日期格式错误，需要 YYYY-MM-DD: %w", err)
		}
		fromTime = t
		hasFrom = true
	}
	if toDate != "" {
		t, err := time.Parse("2006-01-02", toDate)
		if err != nil {
			return fmt.Errorf("--to 日期格式错误，需要 YYYY-MM-DD: %w", err)
		}
		// 将结束日期设置为当天结束 (23:59:59)
		toTime = t.Add(24*time.Hour - time.Second)
		hasTo = true
	}
	if hasFrom || hasTo {
		rangeDesc := ""
		if hasFrom {
			rangeDesc += "从 " + fromTime.Format("2006-01-02")
		}
		if hasTo {
			rangeDesc += " 到 " + toTime.Format("2006-01-02")
		}
		slog.Info("时间范围过滤已启用", "range", rangeDesc)
	}

	//获取作品集的所有文章
	//album.AlbumUrl会类似于 https://afdian.com/album/xyz
	re := regexp.MustCompile("^.*/album/")
	albumId := re.ReplaceAllString(album.AlbumUrl, "")

	albumInfo, err := afdian.GetAlbumInfo(cfg, albumId, cookieString)
	if err != nil {
		return err
	}
	time.Sleep(time.Millisecond * time.Duration(afdian.DelayMs))

	albumSaveDir := filepath.Join(cfg.DataDir, albumInfo.AuthorUrlSlug, albumInfo.AlbumName)
	if err := os.MkdirAll(albumSaveDir, os.ModePerm); err != nil {
		return fmt.Errorf("create album dir <%s> error: %v", albumSaveDir, err)
	}

	//边获取边下载
	var i int64
	downloadedCount := 0
	skippedCount := 0
	outOfRangeCount := 0
	for i = 0; i < albumInfo.PostCount; i += 10 {
		postList, err := afdian.GetAlbumPostPage(cfg, albumId, cookieString, i, "desc")
		if err != nil {
			return err
		}

		for _, post := range postList {
			// 时间范围过滤
			if hasFrom && post.PublishTime.Before(fromTime) {
				slog.Debug("跳过（早于起始日期）", "title", post.Name, "publishTime", post.PublishTime.Format("2006-01-02"))
				outOfRangeCount++
				continue
			}
			if hasTo && post.PublishTime.After(toTime) {
				slog.Debug("跳过（晚于结束日期）", "title", post.Name, "publishTime", post.PublishTime.Format("2006-01-02"))
				outOfRangeCount++
				continue
			}

			timePrefix := post.PublishTime.Format("2006-01-02_15_04_05")
			filePath := filepath.Join(albumSaveDir, timePrefix+"_"+post.Name+".md")

			skipped, err := storage.SavePostIfNotExist(cfg, filePath, post, authToken, disableComment, converter)
			if err != nil {
				if cfg.SkipFailed {
					slog.Error("下载失败，跳过", "title", post.Name, "url", post.Url, "err", err)
					continue
				}
				return err
			}
			if skipped {
				skippedCount++
			} else {
				downloadedCount++
			}
			if quickUpdate && skipped {
				slog.Info("Quick update: 检测到已存在文件，跳过剩余作品集文章", "album", albumInfo.AlbumName)
				goto done
			}
		}
		time.Sleep(time.Millisecond * time.Duration(afdian.DelayMs))
	}

done:
	if hasFrom || hasTo {
		slog.Info("时间范围过滤统计", "下载", downloadedCount, "已存在跳过", skippedCount, "超出范围跳过", outOfRangeCount)
	}
	return nil
}