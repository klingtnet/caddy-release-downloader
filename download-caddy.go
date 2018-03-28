package main

import (
	"encoding/json"
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

type Plugin struct {
	Name string `json:"Name"`
}

type DownloadMetadata struct {
	Plugins []Plugin `json:"plugins"`
}

func getFeatures() []string {
	featureURL := "https://caddyserver.com/api/download-page"
	res, err := http.Get(featureURL)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	decoder := json.NewDecoder(res.Body)
	var data DownloadMetadata
	err = decoder.Decode(&data)
	if err != nil {
		panic(err)
	}
	plugins := make([]string, len(data.Plugins))
	for i, v := range data.Plugins {
		plugins[i] = v.Name
	}
	return plugins
}

func filterFeatures(features []string) []string {
	filtered := make([]string, 0)
	for _, feature := range features {
		switch feature {
		case "hook.pluginloader", "hook.service", "http.grpc":
			continue
		}
		if strings.Contains(feature, "dns") {
			continue
		}
		filtered = append(filtered, feature)
	}
	return filtered
}

func main() {
	if len(os.Args) < 2 {
		binName := os.Args[0]
		fmt.Printf("Usage: %s <VERSION>\nExample: %s 0.9.3\n", binName, binName)
		os.Exit(1)
	}

	failed := false
	ver := os.Args[1]
	archs := []string{"386", "amd64", "arm7", "arm6"}
	rootURL := "https://caddyserver.com/download/linux"
	features := filterFeatures(getFeatures())
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

		queryString := fmt.Sprintf("license=personal&plugins=%s", featureList)
		rawURL := fmt.Sprintf("%s/%s?%s", rootURL, arch, queryString)
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
