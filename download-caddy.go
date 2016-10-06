package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"
)

func download(u string, filepath string, wg *sync.WaitGroup, errCh chan<- error) {
	defer wg.Done()

	file, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil && os.IsExist(err) {
		fmt.Println(filepath, "already downloaded, skipping ...")
		return
	}
	if err != nil {
		errCh <- err
		return
	}

	defer file.Close()

	resp, err := http.Get(u)
	if resp.StatusCode != 200 {
		errCh <- fmt.Errorf("Download '%s' failed: %d", u, resp.StatusCode)
		return
	}
	if err != nil {
		errCh <- err
		return
	}

	defer resp.Body.Close()

	fmt.Println("Saving to", filepath)
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		errCh <- err
	}
}

func main() {
	if len(os.Args) < 2 {
		binName := os.Args[0]
		fmt.Printf("Usage: %s <VERSION>\nExample: %s 0.9.3\n", binName, binName)
		os.Exit(1)
	}

	failed := false
	ver := os.Args[1]
	archs := []string{"386", "amd64", "arm"}
	rootURL := "https://caddyserver.com/download/build"
	features := []string{"awslambda", "cors", "filemanager", "git", "hugo", "ipfilter", "jwt", "locale", "mailout", "minify", "multipass", "prometheus", "ratelimit", "realip", "search", "upload", "cloudflare", "digitalocean", "dnsimple", "dyn", "gandi", "googlecloud", "namecheap", "rfc2136", "route53", "vultr"}

	featureList := url.QueryEscape(strings.Join(features, ","))

	var wg sync.WaitGroup
	errCh := make(chan error, 1)
	defer close(errCh)
	for _, arch := range archs {
		filename := fmt.Sprintf("caddy-all-plugins-%s-%s.tar.gz", ver, arch)
		wd, err := os.Getwd()
		if err != nil {
			errCh <- err
			continue
		}
		filepath := path.Join(wd, filename)

		queryString := fmt.Sprintf("os=linux&arch=%s&features=%s", arch, featureList)
		rawURL := fmt.Sprintf("%s?%s", rootURL, queryString)
		u, err := url.Parse(rawURL)
		if err != nil {
			errCh <- err
			continue
		}

		wg.Add(1)
		fmt.Println("Starting download for caddy", arch)
		go download(u.String(), filepath, &wg, errCh)
	}
	go func() {
		for {
			err := <-errCh
			failed = true
			fmt.Fprintln(os.Stderr, err)
		}
	}()
	wg.Wait()

	if failed {
		os.Exit(2)
	}
}
