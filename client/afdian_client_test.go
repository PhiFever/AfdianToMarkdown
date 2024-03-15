package client

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	cookies      = ReadCookiesFromFile("../cookies.json")
	cookieString = GetCookieString(cookies)
)

func TestGetAuthorId(t *testing.T) {
	type args struct {
		authorName   string
		referer      string
		cookieString string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "case: q9adg",
			args: args{
				authorName:   "q9adg",
				referer:      "https://afdian.net/a/q9adg",
				cookieString: cookieString,
			},
			want: "3f49234e3e8f11eb8f6152540025c377",
		},
		{
			name: "case: sunx1983",
			args: args{
				authorName:   "sunx1983",
				referer:      "https://afdian.net/a/sunx1983",
				cookieString: cookieString,
			},
			want: "6de4661e386611e8955c52540025c377",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, GetAuthorId(tt.args.authorName, tt.args.referer, tt.args.cookieString))
		})
	}
}

func TestGetAlbumListByInterface(t *testing.T) {
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
			name: "case: q9adg",
			args: args{
				userId:       GetAuthorId("q9adg", "https://afdian.net/a/q9adg", cookieString),
				referer:      "https://afdian.net/a/q9adg/album",
				cookieString: cookieString,
			},
			want: []Album{
				{AlbumName: "会员专享", AlbumUrl: "https://afdian.net/album/6f4b70763eb511eb957d52540025c377"},
				{AlbumName: "钩沉", AlbumUrl: "https://afdian.net/album/9eb1469496c711eda7395254001e7c00"},
				{AlbumName: "南斗集", AlbumUrl: "https://afdian.net/album/c2624006a35111eeaebb52540025c377"},
			},
		},
		{
			name: "case: sunx1983",
			args: args{
				userId:       GetAuthorId("sunx1983", "https://afdian.net/a/sunx1983", cookieString),
				referer:      "https://afdian.net/a/sunx1983/album",
				cookieString: cookieString,
			},
			want: []Album{
				{AlbumName: "杂图", AlbumUrl: "https://afdian.net/album/ea8ef19cb1ae11eb8f6c52540025c377"},
				{AlbumName: "欢乐懒朋友", AlbumUrl: "https://afdian.net/album/b55ba95cfedb11e9978452540025c377"},
				{AlbumName: "草稿短篇", AlbumUrl: "https://afdian.net/album/ee21ca0077b311ee956e5254001e7c00"},
				{AlbumName: "无敌勇者王", AlbumUrl: "https://afdian.net/album/6192d1924b4411eeabfa5254001e7c00"},
				{AlbumName: "超有病以及更早时期", AlbumUrl: "https://afdian.net/album/669ccf62a32511ec950552540025c377"},
				{AlbumName: "谷道修仙 全彩版", AlbumUrl: "https://afdian.net/album/f51825d0493211ed9ae552540025c377"},
				{AlbumName: "口袋畜生的复仇", AlbumUrl: "https://afdian.net/album/233329c8fff611ecb27552540025c377"},
				{AlbumName: "P眼修仙", AlbumUrl: "https://afdian.net/album/6eacd2a014e111ed86ac52540025c377"},
				{AlbumName: "设定、废弃漫画和废案", AlbumUrl: "https://afdian.net/album/8668cd58b1bf11ebb19d52540025c377"},
				{AlbumName: "《用漫画来玩环世界吧！》", AlbumUrl: "https://afdian.net/album/044ff41ad7f711ea898f52540025c377"},
				{AlbumName: "超脑洞", AlbumUrl: "https://afdian.net/album/ab167dda788411eca1ef52540025c377"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, GetAlbumListByInterface(tt.args.userId, tt.args.referer, tt.args.cookieString))
		})
	}
}
