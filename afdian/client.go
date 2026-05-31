package afdian

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"os"
	"time"

	retry "github.com/avast/retry-go/v4"
	"github.com/carlmjohnson/requests"
	"golang.org/x/exp/slog"
)

const (
	DelayMs         = 150
	ChromeUserAgent = `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/147.0.0.0 Safari/537.36`
)

// 单次 HTTP 请求的超时与失败重试参数（可在测试中覆盖）
var (
	RequestTimeout      = 30 * time.Second
	RetryAttempts  uint = 3
	RetryDelay          = 1 * time.Second
)

// MediaDownloadDelay 媒体下载间的随机等待（5~15s + 抖动），避免触发限流
func MediaDownloadDelay() {
	baseMs := 5000 + rand.IntN(10001)
	jitterMs := 500 + rand.IntN(1001)
	if rand.IntN(2) == 0 {
		jitterMs = -jitterMs
	}
	delay := time.Duration(baseMs+jitterMs) * time.Millisecond
	slog.Info("Waiting before next download", "delay", delay)
	time.Sleep(delay)
}

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
		"sec-ch-ua":          {`"Google Chrome";v="147", "Chromium";v="147", "Not_A Brand";v="24"`},
		"sec-ch-ua-mobile":   {"?0"},
		"sec-ch-ua-platform": {`"Windows"`},
		"sec-fetch-dest":     {"empty"},
		"sec-fetch-mode":     {"cors"},
		"sec-fetch-site":     {"same-origin"},
		"sec-gpc":            {"1"},
		"user-agent":         {ChromeUserAgent},
	}
}

// NewRequestGet 发送GET请求，带超时控制与失败退避重试
func NewRequestGet(host string, Url string, cookieString string, referer string) ([]byte, error) {
	var body bytes.Buffer
	err := retry.Do(
		func() error {
			body.Reset()
			ctx, cancel := context.WithTimeout(context.Background(), RequestTimeout)
			defer cancel()
			return requests.
				URL(Url).
				Headers(buildAfdianHeaders(host, cookieString, referer)).
				ToBytesBuffer(&body).
				Fetch(ctx)
		},
		retry.Attempts(RetryAttempts),
		retry.Delay(RetryDelay),
		retry.DelayType(retry.BackOffDelay),
		retry.LastErrorOnly(true),
		retry.RetryIf(func(err error) bool {
			// 仅对 5xx 与 429 重试；其余 4xx 客户端错误视为永久失败，立即放弃
			var re *requests.ResponseError
			if errors.As(err, &re) {
				return re.StatusCode >= 500 || re.StatusCode == http.StatusTooManyRequests
			}
			// 传输层错误（超时、连接失败等）：重试
			return true
		}),
		retry.OnRetry(func(n uint, err error) {
			slog.Warn("request failed, retrying", "attempt", n+1, "url", Url, "err", err)
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("GET %s failed: %w", Url, err)
	}
	return body.Bytes(), nil
}
