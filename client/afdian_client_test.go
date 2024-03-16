package client

import (
	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	cookies      = ReadCookiesFromFile("../cookies.json")
	cookieString = GetCookiesString(cookies)
	authToken    = GetAuthTokenCookieString(cookies)
	converter    = md.NewConverter("", true, nil)
)

func TestGetAuthorId(t *testing.T) {
	type args struct {
		authorName string
		referer    string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "case: q9adg",
			args: args{
				authorName: "q9adg",
				referer:    "https://afdian.net/a/q9adg",
			},
			want: "3f49234e3e8f11eb8f6152540025c377",
		},
		{
			name: "case: sunx1983",
			args: args{
				authorName: "sunx1983",
				referer:    "https://afdian.net/a/sunx1983",
			},
			want: "6de4661e386611e8955c52540025c377",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, GetAuthorId(tt.args.authorName, tt.args.referer, cookieString))
		})
	}
}

func TestGetAlbumListByInterface(t *testing.T) {
	type args struct {
		userId  string
		referer string
	}
	tests := []struct {
		name string
		args args
		want []Album
	}{
		{
			name: "case: q9adg",
			args: args{
				userId:  GetAuthorId("q9adg", "https://afdian.net/a/q9adg", cookieString),
				referer: "https://afdian.net/a/q9adg/album",
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
				userId:  GetAuthorId("sunx1983", "https://afdian.net/a/sunx1983", cookieString),
				referer: "https://afdian.net/a/sunx1983/album",
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
			assert.Equal(t, tt.want, GetAlbumListByInterface(tt.args.userId, tt.args.referer, cookieString))
		})
	}
}

func TestGetAlbumArticleListByInterface(t *testing.T) {
	type args struct {
		albumId string
	}
	tests := []struct {
		name string
		args args
		want []Article
	}{
		{
			name: "case: c2624006a35111eeaebb52540025c377 (南斗集)",
			args: args{
				albumId: "c2624006a35111eeaebb52540025c377",
			},
			want: []Article{
				{ArticleName: "早恋 1", ArticleUrl: "https://afdian.net/album/c2624006a35111eeaebb52540025c377/0c26f170a4ea11eea1de52540025c377"},
				{ArticleName: "早恋 2", ArticleUrl: "https://afdian.net/album/c2624006a35111eeaebb52540025c377/f2352a16a5bb11ee92325254001e7c00"},
				{ArticleName: "爱的基本约定 1", ArticleUrl: "https://afdian.net/album/c2624006a35111eeaebb52540025c377/8b9bf5fac34d11eeb4db52540025c377"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetAlbumArticleListByInterface(tt.args.albumId, authToken)
			//log.Println("got:", utils.ToJSON(got))
			assert.Equalf(t, tt.want, got, "GetAlbumArticleListByInterface(%v)", tt.args.albumId)
		})
	}
}

func TestGetArticleContentByInterface(t *testing.T) {
	type args struct {
		articleUrl string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "土地换和平真的有错吗？",
			args: args{
				articleUrl: "https://afdian.net/album/9eb1469496c711eda7395254001e7c00/e18ec048a36211ed867452540025c377",
			},
			want: "去跟挨了两枚原子弹的日本讲讲看。\n土地换的哪里是和平，换的是免除无意义的死亡。\n有的失败已经足够显然，再为了某些野心家和中二学生的愚蠢牺牲大量的人命，显然是不道德的。\n如果土地换和平是绝对邪恶的，那么你是在告诉台湾人面对绝对军事优势也要弃绝和平统一，在告诉日本人挨了原子弹和东京大轰炸也要咬死一亿总玉碎。因为按照你的“土地换和平绝对不道德/不明智论”，这是ta们逻辑上唯一能令你满意的选择。\n来，你再说一遍看看，这为什么是道德的？\n为什么为了让你在知乎占上风，这几亿人就应该死绝？\n\n再次提醒，作为不灭之国，我们中国的很多固有信念和传统，对其它国家、其它民族是不适用的。把自己独有禀赋的要求强往人家身上套，要求别的国家拿命满足自己的道德审美，是一种不道德的行为。\n大象非逼着老鼠学自己对猫勇敢，就是在逼老鼠去死。\n如果只要坚持抵抗就能胜利，世界上哪来这么多灭绝的民族和灭亡的国家？难道每一个民族灭亡的原因都是因为没抵抗？不勇敢？\n\n小学生都知道的道理，为什么你读完大学都不懂？",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, GetArticleContentByInterface(tt.args.articleUrl, authToken, converter), "GetArticleContentByInterface(%v)", tt.args.articleUrl)
		})
	}
}
