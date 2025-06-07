package utils

import (
	"fmt"
	"golang.org/x/exp/slog"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

const (
	ImgDir = ".assets"
)

var CookiePath = path.Join(GetExecutionPath(), `cookies.json`)

func GetExecutionTime(startTime, endTime time.Time) string {
	duration := endTime.Sub(startTime)
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	seconds := duration.Seconds() - float64(hours*3600+minutes*60)

	result := ""
	if hours > 0 {
		result += fmt.Sprintf("%dh", hours)
	}
	if minutes > 0 {
		result += fmt.Sprintf("%dmin", minutes)
	}
	result += fmt.Sprintf("%.2fs", seconds)
	return result
}

func ToSafeFilename(in string) string {
	//https://stackoverflow.com/questions/1976007/what-characters-are-forbidden-in-windows-and-linux-directory-names
	//全部替换为_
	rp := strings.NewReplacer(
		"/", "_",
		`\`, "_",
		"<", "_",
		">", "_",
		":", "_",
		`"`, "_",
		"|", "_",
		"?", "_",
		"*", "_",
	)
	rt := rp.Replace(in)
	return rt
}

// CheckAndListAuthors 通过检查程序目录下是否有二级文件夹 motions 来获取所有的作者名
// 如果有，则返回所有一级文件夹名
func CheckAndListAuthors() ([]string, error) {
	var folders []string

	// 获取当前目录路径
	currentDir := GetExecutionPath()
	//fmt.Println("CurrentDir: ", currentDir)

	// 读取当前目录下的所有文件和文件夹
	files, err := os.ReadDir(currentDir)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		//fmt.Println("file: ", file.Name())
		if file.IsDir() {
			// 检查是否存在二级文件夹 motion
			motionPath := filepath.Join(currentDir, file.Name(), "motions")
			if _, err := os.Stat(motionPath); err == nil {
				folders = append(folders, file.Name())
			}
		}
	}

	//fmt.Println("folders: ", folders)
	return folders, nil
}

// GetExecutionPath 获取程序的实际执行目录
func GetExecutionPath() string {
	ex, err := os.Executable()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(-1)
	}
	return filepath.Dir(ex)
}
