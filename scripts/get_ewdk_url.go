package main

import (
	"fmt"
	"github.com/ddkwork/golibrary/std/stream/net/httpClient"
	"os"
	"strings"
)

func main() {
	GetIsoLink()
}

func GetIsoLink() string {
	c := httpClient.New() //.SetDebug(true)

	b := c.Get("https://learn.microsoft.com/en-us/legal/windows/hardware/enterprise-wdk-license-2022").Request().Buffer
	latestUrl := ""
	for s := range strings.Lines(b.String()) {
		if strings.Contains(s, "Accept license terms") {
			before, after, found := strings.Cut(s, `" data-linktype`)
			if found {
				before, after, found = strings.Cut(before, `href="`)
				latestUrl = after
				break
			}
		}
	}
	c.Get(latestUrl).StopCode(302).Request()
	iso := c.Response.Header.Get("Location")

	if githubEnv := os.Getenv("GITHUB_ENV"); githubEnv != "" {
		// 在GitHub Actions中运行 - 写入到GITHUB_ENV文件
		envFile, err := os.OpenFile(githubEnv, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ 无法打开GITHUB_ENV文件: %v\n", err)
			os.Exit(1)
		}
		defer envFile.Close()

		if _, err := fmt.Fprintf(envFile, "EWDK_ISO_URL=%s\n", iso); err != nil {
			fmt.Fprintf(os.Stderr, "❌ 写入GITHUB_ENV失败: %v\n", err)
			os.Exit(1)
		}
	} else {
		// 本地运行 - 直接设置环境变量
		if err := os.Setenv("EWDK_ISO_URL", iso); err != nil {
			fmt.Fprintf(os.Stderr, "❌ 设置环境变量失败: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Print(iso)
	return iso
}
