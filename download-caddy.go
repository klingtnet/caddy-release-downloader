package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func errExit(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	ver := "0.9.3"
	archs := []string{"386", "amd64", "arm"}
	rootURL := "https://caddyserver.com/download/build"
	features := []string{"awslambda", "cors", "filemanager", "git", "hugo", "ipfilter", "jwt", "locale", "mailout", "minify", "multipass", "prometheus", "ratelimit", "realip", "search", "upload", "cloudflare", "digitalocean", "dnsimple", "dyn", "gandi", "googlecloud", "namecheap", "rfc2136", "route53", "vultr"}

	featureList := url.QueryEscape(strings.Join(features, ","))

	for _, arch := range archs {
		filename := fmt.Sprintf("caddy-all-plugins-%s-%s.tar.gz", ver, arch)
		file, err := os.Create(filename)
		errExit(err)
		defer file.Close()

		queryString := fmt.Sprintf("os=linux&arch=%s&features=%s", arch, featureList)
		rawURL := fmt.Sprintf("%s?%s", rootURL, queryString)
		fmt.Println(rawURL)
		u, err := url.Parse(rawURL)
		errExit(err)

		resp, err := http.Get(u.String())
		errExit(err)
		defer resp.Body.Close()

		_, err = io.Copy(file, resp.Body)
		errExit(err)
	}
}
