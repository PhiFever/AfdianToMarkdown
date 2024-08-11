package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	ImgDir     = ".assets"
	CookiePath = `cookies.json`
)

// ReadListFile 用于按行读取列表文件，返回一个字符串切片
func ReadListFile(filePath string) ([]string, error) {
	var contentList []string
	file, err := os.Open(filePath)
	if err != nil {
		return contentList, err
	}
	defer func(file *os.File) {
		if err := file.Close(); err != nil {
			panic(err)
		}
	}(file)

	var fileLine string
	for {
		_, err := fmt.Fscanln(file, &fileLine)
		if err != nil {
			break
		}
		contentList = append(contentList, fileLine)
	}
	return contentList, nil
}

func GetExecutionTime(startTime time.Time, endTime time.Time) string {
	//按时:分:秒格式输出
	duration := endTime.Sub(startTime)
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	seconds := int(duration.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%d时%d分%d秒", hours, minutes, seconds)
	} else if minutes > 0 {
		return fmt.Sprintf("%d分%d秒", minutes, seconds)
	} else {
		return fmt.Sprintf("%d秒", seconds)
	}
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

func ToJSON(JSONString interface{}) string {
	b, err := json.Marshal(JSONString)
	if err != nil {
		return fmt.Sprintf("%+v", JSONString)
	}
	var out bytes.Buffer
	err = json.Indent(&out, b, "", "    ")
	if err != nil {
		return fmt.Sprintf("%+v", JSONString)
	}
	return out.String()
}

// CheckAndListAuthors 通过检查当前目录下是否有二级文件夹 motion 来获取所有的作者名
// 如果有，则返回所有一级文件夹名
func CheckAndListAuthors() ([]string, error) {
	var folders []string

	// 获取当前目录路径
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

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
