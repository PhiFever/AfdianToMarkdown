package afdian

import (
	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"testing"
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
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, GetAuthorId(tt.args.authorName, tt.args.referer, tt.args.cookieString), "GetAuthorId(%v, %v, %v)", tt.args.authorName, tt.args.referer, tt.args.cookieString)
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
		name  string
		args  args
		want  []Article
		want1 string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := GetAuthorMotionUrlList(tt.args.userName, tt.args.cookieString, tt.args.prevPublishSn)
			assert.Equalf(t, tt.want, got, "GetAuthorMotionUrlList(%v, %v, %v)", tt.args.userName, tt.args.cookieString, tt.args.prevPublishSn)
			assert.Equalf(t, tt.want1, got1, "GetAuthorMotionUrlList(%v, %v, %v)", tt.args.userName, tt.args.cookieString, tt.args.prevPublishSn)
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
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, GetAlbumList(tt.args.userId, tt.args.referer, tt.args.cookieString), "GetAlbumList(%v, %v, %v)", tt.args.userId, tt.args.referer, tt.args.cookieString)
		})
	}
}

func TestGetAlbumArticleList(t *testing.T) {
	type args struct {
		albumId   string
		authToken string
	}
	tests := []struct {
		name string
		args args
		want []Article
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, GetAlbumArticleList(tt.args.albumId, tt.args.authToken), "GetAlbumArticleList(%v, %v)", tt.args.albumId, tt.args.authToken)
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
			assert.Equalf(t, tt.want, GetArticleContent(tt.args.articleUrl, tt.args.authToken, tt.args.converter), "GetArticleContent(%v, %v, %v)", tt.args.articleUrl, tt.args.authToken, tt.args.converter)
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
			gotCommentString, gotHotCommentString := GetArticleComment(tt.args.articleUrl, tt.args.cookieString)
			assert.Equalf(t, tt.wantCommentString, gotCommentString, "GetArticleComment(%v, %v)", tt.args.articleUrl, tt.args.cookieString)
			assert.Equalf(t, tt.wantHotCommentString, gotHotCommentString, "GetArticleComment(%v, %v)", tt.args.articleUrl, tt.args.cookieString)
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
