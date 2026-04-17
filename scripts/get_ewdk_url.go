package main

import (
	"os"
	"strconv"
	"strings"

	"github.com/ddkwork/golibrary/std/mylog"
	"github.com/ddkwork/golibrary/std/safemap"
	"github.com/ddkwork/golibrary/std/stream"
	"github.com/ddkwork/golibrary/std/stream/net/httpClient"
)

func main() {
	GetIsoLink()
}

func GetIsoLink() string {
	c := httpClient.New() //.SetDebug(true)

	b := c.Get("https://learn.microsoft.com/en-us/legal/windows/hardware/enterprise-wdk-license-2022").Request().Buffer
	m := safemap.M[int64, string]{}
	for s := range strings.Lines(b.String()) {
		if strings.Contains(s, "Accept license terms") {
			before, after, found := strings.Cut(s, `" data-linktype`)
			if found {
				before, after, found = strings.Cut(before, `href="`)
				_, idStr, f := strings.Cut(after, "linkid=")
				if f {
					id := mylog.Check2(strconv.ParseInt(idStr, 10, 64))
					m.Set(id, after)
				}
			}
		}
	}
	latest := int64(0)
	latestUrl := ""
	keys := m.Keys()
	for _, id := range keys {
		url, _ := m.Get(id)
		if id > latest {
			latest = id
			latestUrl = url
		}
	}
	mylog.Struct(latest)
	c.Get(latestUrl).StopCode(302).Request()
	iso := c.Response.Header.Get("Location")
	mylog.Success(iso)
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
