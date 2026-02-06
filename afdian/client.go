package afdian

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/carlmjohnson/requests"
)

const (
	DelayMs         = 150
	ChromeUserAgent = `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36`
)

// ReadCookiesFromFile 从文件中读取 Cookies
func ReadCookiesFromFile(filePath string) ([]Cookie, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var cookies []Cookie
	if err := json.Unmarshal(data, &cookies); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cookies: %w", err)
	}

	return cookies, nil
}

func GetCookiesString(cookies []Cookie) (cookiesString string) {
	for _, cookie := range cookies {
		cookiesString += cookie.Name + "=" + cookie.Value + ";"
	}
	return cookiesString
}

func GetAuthTokenString(cookies []Cookie) (authTokenString string) {
	for _, cookie := range cookies {
		if cookie.Name == "auth_token" {
			authTokenString = fmt.Sprintf("auth_token=%s", cookie.Value)
		}
	}
	return authTokenString
}

func GetCookies(cookiePath string) (cookieString string, authToken string, err error) {
	cookies, err := ReadCookiesFromFile(cookiePath)
	if err != nil {
		return "", "", fmt.Errorf("failed to read cookies from file: %w", err)
	}
	cookieString = GetCookiesString(cookies)
	authToken = GetAuthTokenString(cookies)
	return cookieString, authToken, nil
}

func buildAfdianHeaders(host string, cookieString string, referer string) http.Header {
	return http.Header{
		"authority":          {host},
		"accept":             {"accept", "application/json, text/plain, */*"},
		"accept-language":    {"zh-CN,zh;q=0.9,en;q=0.8"},
		"cache-control":      {"no-cache"},
		"cookie":             {cookieString},
		"dnt":                {"1"},
		"locale-lang":        {"zh-CN"},
		"pragma":             {"no-cache"},
		"referer":            {referer},
		"sec-ch-ua":          {`"Google Chrome";v="131", "Chromium";v="131", "Not_A Brand";v="24"`},
		"sec-ch-ua-mobile":   {"?0"},
		"sec-ch-ua-platform": {`"Windows"`},
		"sec-fetch-dest":     {"empty"},
		"sec-fetch-mode":     {"cors"},
		"sec-fetch-site":     {"same-origin"},
		"sec-gpc":            {"1"},
		"user-agent":         {ChromeUserAgent},
	}
}

// NewRequestGet 发送GET请求
func NewRequestGet(host string, Url string, cookieString string, referer string) ([]byte, error) {
	var body bytes.Buffer
	err := requests.
		URL(Url).
		Headers(buildAfdianHeaders(host, cookieString, referer)).
		ToBytesBuffer(&body).
		Fetch(context.Background())
	if err != nil {
		return nil, fmt.Errorf("GET %s failed: %w", Url, err)
	}
	return body.Bytes(), nil
}
