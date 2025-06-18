package main

import (
	"fmt"
	"github.com/ddkwork/golibrary/std/mylog"
	"github.com/ddkwork/golibrary/std/stream"
	"github.com/ddkwork/golibrary/std/stream/net/httpClient"
	"os"
	"strings"
)

func detectAction() {
	fmt.Println("GITHUB_ACTIONS =", os.Getenv("GITHUB_ACTIONS"))
	fmt.Println("GITHUB_WORKSPACE =", os.Getenv("GITHUB_WORKSPACE"))
	fmt.Println("GITHUB_RUN_ID =", os.Getenv("GITHUB_RUN_ID"))

	// 检查GitHub运行器目录是否存在
	if _, err := os.Stat("/opt/hostedtoolcache"); err == nil {
		fmt.Println("✅ Detected GitHub Actions runner directory")
	} else {
		fmt.Println("❌ GitHub Actions runner directory not found")
	}
}

func main() {
	//detectAction()
	panic(stream.IsRunningOnGitHubActions())
	return
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
		g := stream.NewGeneratedFile()
		g.P()
		g.P("EWDK_ISO_URL", "=", iso)
		stream.WriteAppend(githubEnv, g.String())

		mylog.Trace("env path", githubEnv)
		mylog.Json("github_env", string(mylog.Check2(os.ReadFile(githubEnv))))

		//filepath.Walk(filepath.Dir(githubEnv), func(path string, info fs.FileInfo, err error) error {
		//	println(path)
		//	return err
		//})

	} else {
		mylog.Check(os.Setenv("EWDK_ISO_URL", iso))
	}
	return iso
}
