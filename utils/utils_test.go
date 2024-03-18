package utils

import (
	"github.com/stretchr/testify/assert"
	"os"
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
				filePath: "C:\\SysLog.ini",
			},
			want:  nil, // Aim to fail
			want1: true,
		},
		{
			name: "TestFileNotExist",
			args: args{
				filePath: "C:\\1234567890.txt",
			},
			want:  nil,
			want1: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := FileExists(tt.args.filePath)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.want1, got1)
		})
	}
}
