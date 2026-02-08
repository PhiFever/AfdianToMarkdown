package afdian

import (
	"AfdianToMarkdown/config"
	"AfdianToMarkdown/logger"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"golang.org/x/exp/slog"
)

const q9adgId = "3f49234e3e8f11eb8f6152540025c377"

var (
	cookieString, authToken string
	cfg                     *config.Config
)

func init() {
	//localPath, _ := os.Getwd()
	//执行测试前，先设置cookie路径为实际本地路径
	slog.SetDefault(logger.SetupLogger(slog.LevelInfo))
	cfg = config.NewConfig("afdian.com", `D:\MyProject\Golang\AfdianToMarkdown\data`, `D:\MyProject\Golang\AfdianToMarkdown\cookies.json`)
	slog.Info("cookiePath:", "path", cfg.CookiePath)
	var err error
	cookieString, authToken, err = GetCookies(cfg.CookiePath)
	if err != nil {
		panic(fmt.Sprintf("failed to load cookies: %v", err))
	}
}

func getAlbumUrl(AlbumId string) string {
	return fmt.Sprintf("%s/album/%s", cfg.HostUrl, AlbumId)
}

func TestGetAuthorId(t *testing.T) {
	type args struct {
		authorUrlSlug string
		referer       string
		cookieString  string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "q9adg",
			args: args{
				authorUrlSlug: "q9adg",
				referer:       cfg.HostUrl,
				cookieString:  cookieString,
			},
			want: q9adgId,
		},
		{name: "深海巨狗",
			args: args{
				authorUrlSlug: "Arabian_nights",
				referer:       cfg.HostUrl,
				cookieString:  cookieString,
			},
			want: "d7c0ebe2c83911ea8ad552540025c377",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetAuthorId(cfg, tt.args.authorUrlSlug, tt.args.referer, tt.args.cookieString)
			assert.NoError(t, err)
			assert.Equalf(t, tt.want, got, "GetAuthorId(%v, %v, %v)", tt.args.authorUrlSlug, tt.args.referer, tt.args.cookieString)
		})
	}
}

func TestGetAlbumList(t *testing.T) {
	type args struct {
		userId       string
		referer      string
		cookieString string
	}
	tests := []struct {
		name string
		args args
		want []Album
	}{
		{
			name: "q9adg",
			args: args{
				userId:       q9adgId,
				referer:      cfg.HostUrl,
				cookieString: cookieString,
			},
			want: []Album{
				{AlbumName: "个人成长", AlbumUrl: getAlbumUrl("38dcf1ee3b1a11efba7c52540025c377")},
				{AlbumName: "亲子教育", AlbumUrl: getAlbumUrl("831821ee3b1911efbe3c52540025c377")},
				{AlbumName: "职业伦理", AlbumUrl: getAlbumUrl("9ff4fff83b1911efbb2b52540025c377")},
				{AlbumName: "钩沉", AlbumUrl: getAlbumUrl("9eb1469496c711eda7395254001e7c00")},
				{AlbumName: "亲密关系", AlbumUrl: getAlbumUrl("4feb06ca3b1811ef8b9c52540025c377")},
				{AlbumName: "社区互动", AlbumUrl: getAlbumUrl("72d0d32a4f7e11efb83152540025c377")},
				{AlbumName: "开放版权内容", AlbumUrl: getAlbumUrl("9bf7e3084f7c11ef9b6452540025c377")},
				{AlbumName: "会员专享", AlbumUrl: getAlbumUrl("6f4b70763eb511eb957d52540025c377")},
				{AlbumName: "善哉集", AlbumUrl: getAlbumUrl("3c92a37470e911efbb4752540025c377")},
				//《南斗集》已被作者删除，迁移至《亲密关系》
				//{AlbumName: "南斗集", AlbumUrl: getAlbumUrl("c2624006a35111eeaebb52540025c377")},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotList, err := GetAlbumList(cfg, tt.args.userId, tt.args.referer, tt.args.cookieString)
			assert.NoError(t, err)
			for _, wantAlbum := range tt.want {
				assert.Contains(t, gotList, wantAlbum, "GetAlbumList(%v, %v, %v) not contains want album: %s", tt.args.userId, tt.args.referer, tt.args.cookieString, wantAlbum)
			}
		})
	}
}

func TestGetAlbumInfo(t *testing.T) {
	tests := []struct {
		name          string
		albumId       string
		wantUrlSlug   string
		wantAlbumName string
	}{
		{
			name:          "阿牟AMFIG",
			albumId:       "aa82d81c88f711edb68f52540025c377",
			wantUrlSlug:   "AMFIG",
			wantAlbumName: "版本更新记录",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := GetAlbumInfo(cfg, tt.albumId, cookieString)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantUrlSlug, info.AuthorUrlSlug)
			assert.Equal(t, tt.wantAlbumName, info.AlbumName)
			assert.Greater(t, info.PostCount, int64(0))
		})
	}
}

func TestGetAlbumPostPage(t *testing.T) {
	tests := []struct {
		name     string
		albumId  string
		wantPost []Post
	}{
		{
			name:    "阿牟AMFIG",
			albumId: "aa82d81c88f711edb68f52540025c377",
			wantPost: []Post{
				{
					Name:     "游戏更新日志 2022-12-30",
					Url:      "https://afdian.com/album/aa82d81c88f711edb68f52540025c377/cf0064c088f711edb42a52540025c377",
					Pictures: []string(nil), PublishTime: time.Date(2022, time.December, 31, 18, 42, 22, 0, time.Local),
				},
				{
					Name:     "版本更新1.0.1 2023-01-13 (附游戏安装包(安卓版))",
					Url:      "https://afdian.com/album/aa82d81c88f711edb68f52540025c377/77d2cdd8935a11ed9ebb5254001e7c00",
					Pictures: []string(nil), PublishTime: time.Date(2023, time.January, 13, 23, 53, 48, 0, time.Local),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPost, err := GetAlbumPostPage(cfg, tt.albumId, cookieString, 0, "asc")
			assert.NoError(t, err)
			assert.Subset(t, gotPost, tt.wantPost)
		})
	}
}

func Test_getCommentString(t *testing.T) {
	t.Run("单条评论", func(t *testing.T) {
		json := `[{"user":{"name":"张三"},"publish_time":"1700000000","content":"好文章"}]`
		got := getCommentString(gjson.Parse(json))
		assert.Contains(t, got, "[0]")
		assert.Contains(t, got, "by 张三")
		assert.Contains(t, got, "好文章")
		assert.Contains(t, got, "2023-11-1")
		assert.NotContains(t, got, "回复")
	})

	t.Run("带回复的评论", func(t *testing.T) {
		json := `[{"user":{"name":"李四"},"publish_time":"1700000000","content":"同意","reply_user":{"name":"张三"}}]`
		got := getCommentString(gjson.Parse(json))
		assert.Contains(t, got, "by 李四")
		assert.Contains(t, got, "> 回复 张三: ")
		assert.Contains(t, got, "同意")
	})

	t.Run("多条评论序号递增", func(t *testing.T) {
		json := `[{"user":{"name":"A"},"publish_time":"1700000000","content":"第一条"},{"user":{"name":"B"},"publish_time":"1700000100","content":"第二条"}]`
		got := getCommentString(gjson.Parse(json))
		assert.Contains(t, got, "[0]")
		assert.Contains(t, got, "[1]")
		assert.Contains(t, got, "by A")
		assert.Contains(t, got, "by B")
	})

	t.Run("空列表", func(t *testing.T) {
		got := getCommentString(gjson.Parse(`[]`))
		assert.Empty(t, got)
	})
}
