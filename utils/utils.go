package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	ImgDir = ".assets"
)

// ResolveAppDir 解析程序所在目录（可执行文件目录或工作目录）
// 用于推断默认的数据目录和 cookie 路径
func ResolveAppDir() (string, error) {
	// 1. 尝试可执行文件目录
	ex, err := os.Executable()
	if err == nil {
		execDir := filepath.Dir(ex)
		// 排除 go run 产生的临时目录
		if !strings.Contains(execDir, "go-build") {
			return execDir, nil
		}
	}
	// 2. 回退到当前工作目录
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}
	return wd, nil
}

// DefaultDataDir 返回默认的数据目录（appDir/data）
func DefaultDataDir(appDir string) string {
	return filepath.Join(appDir, "data")
}

// DefaultCookiePath 返回默认的 cookie 文件路径（appDir/cookies.json）
func DefaultCookiePath(appDir string) string {
	return filepath.Join(appDir, "cookies.json")
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
