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

// DefaultCookiePath 返回默认的 cookie 文件路径
func DefaultCookiePath() string {
	return path.Join(GetAppDataPath(), `cookies.json`)
}

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

// CheckAndListAuthors 通过检查 dataDir 下是否有二级文件夹 motions 来获取所有的作者名
// 如果有，则返回所有一级文件夹名
func CheckAndListAuthors(dataDir string) ([]string, error) {
	var folders []string

	// 读取数据目录下的所有文件和文件夹
	files, err := os.ReadDir(dataDir)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		//fmt.Println("file: ", file.Name())
		if file.IsDir() {
			// 检查是否存在二级文件夹 motions
			motionPath := filepath.Join(dataDir, file.Name(), "motions")
			if _, err := os.Stat(motionPath); err == nil {
				folders = append(folders, file.Name())
			}
		}
	}

	//fmt.Println("folders: ", folders)
	return folders, nil
}

// GetAppDataPath 根据 cookies.json 获取程序的实际执行目录
// 若以 build 方式运行，找到可执行文件所在目录
// 若以 go run 方式运行，找到当前工作目录
func GetAppDataPath() string {
	// 1. 尝试查找可执行文件目录
	ex, err := os.Executable()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(-1)
	}
	execDir := filepath.Dir(ex)
	if err == nil {
		execCookie := filepath.Join(execDir, "cookies.json")
		if _, err := os.Stat(execCookie); err == nil {
			return execDir
		}
	}
	// 2. 尝试查找当前工作目录
	wd, err := os.Getwd()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(-1)
	}
	if err == nil {
		wdCookie := filepath.Join(wd, "cookies.json")
		if _, err := os.Stat(wdCookie); err == nil {
			return wd
		}
	}
	slog.Error("Failed to find cookies.json both in", "execDir", execDir, "workDir", filepath.Dir(wd))
	os.Exit(-1)
	return ""
}
