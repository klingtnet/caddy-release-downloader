package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
)

func errExit(e error) {
	if e != nil {
		panic(e)
	}
}

func download(u string, filename string, wg *sync.WaitGroup) {
	defer wg.Done()

	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0644)
	errExit(err)
	defer file.Close()

	resp, err := http.Get(u)
	errExit(err)
	defer resp.Body.Close()

	_, err = io.Copy(file, resp.Body)
	errExit(err)
}

func main() {
	ver := "0.9.3"
	archs := []string{"386", "amd64", "arm"}
	rootURL := "https://caddyserver.com/download/build"
	features := []string{"awslambda", "cors", "filemanager", "git", "hugo", "ipfilter", "jwt", "locale", "mailout", "minify", "multipass", "prometheus", "ratelimit", "realip", "search", "upload", "cloudflare", "digitalocean", "dnsimple", "dyn", "gandi", "googlecloud", "namecheap", "rfc2136", "route53", "vultr"}

	featureList := url.QueryEscape(strings.Join(features, ","))

	var wg sync.WaitGroup
	for _, arch := range archs {
		filename := fmt.Sprintf("caddy-all-plugins-%s-%s.tar.gz", ver, arch)
		queryString := fmt.Sprintf("os=linux&arch=%s&features=%s", arch, featureList)
		rawURL := fmt.Sprintf("%s?%s", rootURL, queryString)
		u, err := url.Parse(rawURL)
		errExit(err)
		wg.Add(1)
		fmt.Println("Starting download for caddy", arch)
		go download(u.String(), filename, &wg)
	}
	wg.Wait()
}
