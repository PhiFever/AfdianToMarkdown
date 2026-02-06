package afdian

import (
	"AfdianToMarkdown/config"
	"AfdianToMarkdown/logger"
	"fmt"
	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"golang.org/x/exp/slog"
	"testing"
)

const q9adgId = "3f49234e3e8f11eb8f6152540025c377"

var (
	cookieString, authToken string
	cfg                     *config.Config
)

func init() {
	//localPath, _ := os.Getwd()
	//执行测试前，先设置cookie路径为实际本地路径
	slog.SetDefault(logger.SetupLogger())
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

func TestGetAuthorMotionUrlList(t *testing.T) {
	type args struct {
		userName      string
		cookieString  string
		prevPublishSn string
	}
	tests := []struct {
		name              string
		args              args
		wantArticleList   []Post
		wantNextPublishSn string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authorArticleList, nextPublishSn, err := GetMotionUrlList(cfg, tt.args.userName, tt.args.cookieString, tt.args.prevPublishSn)
			assert.NoError(t, err)
			assert.Equalf(t, tt.wantArticleList, authorArticleList, "GetMotionUrlList(%v, %v, %v)", tt.args.userName, tt.args.cookieString, tt.args.prevPublishSn)
			assert.Equalf(t, tt.wantNextPublishSn, nextPublishSn, "GetMotionUrlList(%v, %v, %v)", tt.args.userName, tt.args.cookieString, tt.args.prevPublishSn)
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

func TestGetAlbumArticleList(t *testing.T) {
	type args struct {
		albumId   string
		authToken string
	}
	tests := []struct {
		name          string
		args          args
		wantUrlSlug   string
		wantAlbumName string
		wantPost      []Post
	}{

		{
			name: "阿牟AMFIG",
			args: args{
				albumId:   "aa82d81c88f711edb68f52540025c377",
				authToken: authToken,
			},
			wantUrlSlug:   "AMFIG",
			wantAlbumName: "版本更新记录",
			wantPost: []Post{
				{
					Name:     "罗德里[2.1.10]已发布(2025-06-01)",
					Url:      "https://afdian.com/album/aa82d81c88f711edb68f52540025c377/527d91b23e9311f0bddb52540025c377",
					Pictures: []string(nil),
				},
				{
					Name:     "罗德里[2.1.9]已发布(2025-05-26)",
					Url:      "https://afdian.com/album/aa82d81c88f711edb68f52540025c377/b07905ee3a0611f0ba0d52540025c377",
					Pictures: []string{"https://pic1.afdiancdn.com/user/104acec41aa411edbf6152540025c377/common/6cdc5f7c0733550298ad0bb650cf029e_w1280_h1280_s14.png"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotUrlSlug, gotAlbumName, gotPost, err := GetAlbumPostList(cfg, tt.args.albumId, tt.args.authToken)
			assert.NoError(t, err)
			assert.Equalf(t, tt.wantUrlSlug, gotUrlSlug, "GetAlbumArticleList(%v, %v)", tt.args.albumId, tt.args.authToken)
			assert.Equalf(t, tt.wantAlbumName, gotAlbumName, "GetAlbumArticleList(%v, %v)", tt.args.albumId, tt.args.authToken)
			assert.Subset(t, gotPost, tt.wantPost, "GetAlbumArticleList(%v, %v)", tt.args.albumId, tt.args.authToken)
		})
	}
}

func TestGetArticleContent(t *testing.T) {
	type args struct {
		articleUrl string
		authToken  string
		converter  *md.Converter
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetPostContent(cfg, tt.args.articleUrl, tt.args.authToken, tt.args.converter)
			assert.NoError(t, err)
			assert.Equalf(t, tt.want, got, "getPostContent(%v, %v, %v)", tt.args.articleUrl, tt.args.authToken, tt.args.converter)
		})
	}
}

func TestGetArticleComment(t *testing.T) {
	type args struct {
		articleUrl   string
		cookieString string
	}
	tests := []struct {
		name                 string
		args                 args
		wantCommentString    string
		wantHotCommentString string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCommentString, gotHotCommentString, err := GetPostComment(cfg, tt.args.articleUrl, tt.args.cookieString)
			assert.NoError(t, err)
			assert.Equalf(t, tt.wantCommentString, gotCommentString, "GetPostComment(%v, %v)", tt.args.articleUrl, tt.args.cookieString)
			assert.Equalf(t, tt.wantHotCommentString, gotHotCommentString, "GetPostComment(%v, %v)", tt.args.articleUrl, tt.args.cookieString)
		})
	}
}

func Test_getCommentString(t *testing.T) {
	type args struct {
		commentJson gjson.Result
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, getCommentString(tt.args.commentJson), "getCommentString(%v)", tt.args.commentJson)
		})
	}
}
