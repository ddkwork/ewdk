package main

import (
	"github.com/ddkwork/golibrary/std/mylog"
	"github.com/ddkwork/golibrary/std/stream"
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
