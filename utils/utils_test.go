package utils

import (
	"os"
	"reflect"
	"testing"
)

func TestFileExists(t *testing.T) {
	type args struct {
		filePath string
	}
	tests := []struct {
		name  string
		args  args
		want  os.FileInfo
		want1 bool
	}{
		{
			name: "TestFileExists", // Aim to fail
			args: args{
				filePath: "D:\\MyProject\\Golang\\WorkSpace\\AifadianCrawler\\cookies.json",
			},
			want:  nil,
			want1: true,
		},
		{
			name: "TestFileNotExist",
			args: args{
				filePath: "cookies1.json",
			},
			want:  nil,
			want1: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := FileExists(tt.args.filePath)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FileExists() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("FileExists() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
