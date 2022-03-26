package fetchwebpage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/hashicorp/go-multierror"
	"github.com/spf13/afero"
)

type Fetcher interface {
	Fetch(ctx context.Context, url string) error
}

func NewFetcher(downloader Downloader, fs afero.Fs) Fetcher {
	return &fetcherImpl{
		downloader: downloader,
		fs:         fs,
	}
}

type fetcherImpl struct {
	downloader Downloader
	fs         afero.Fs
}

func (f *fetcherImpl) Fetch(ctx context.Context, givenURL string) error {
	parsedURL, err := url.Parse(givenURL)
	if err != nil {
		return fmt.Errorf("failed to parse given url: %w", err)
	}

	var buf bytes.Buffer
	err = f.downloader.Download(ctx, givenURL, &buf)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}

	filename := fmt.Sprintf("%s.html", parsedURL.Host)
	assetDirname := fmt.Sprintf("%s_assets", filename)

	var wg sync.WaitGroup

	errCh := make(chan error)
	var combinedErr error
	wg.Add(1)
	go func() {
		defer wg.Done()
		for err := range errCh {
			combinedErr = multierror.Append(combinedErr, err)
		}
	}()

	assetPathCh := make(chan *url.URL)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for path := range assetPathCh {
			assetURL := parsedURL.ResolveReference(path)
			err := f.downloadAndWriteFile(ctx, assetURL, assetDirname)
			if err != nil {
				errCh <- fmt.Errorf("failed to download asset: %w", err)
			}
		}
	}()

	html, err := findAndUpdateAssetPaths(&buf, assetPathCh, assetDirname)
	err = afero.WriteFile(f.fs, filename, []byte(html), os.FileMode(0664))
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	close(assetPathCh)
	close(errCh)
	wg.Wait()

	if combinedErr != nil {
		return combinedErr
	}

	return nil
}

func (f *fetcherImpl) downloadAndWriteFile(ctx context.Context, assetURL *url.URL, baseDir string) error {
	filename := filepath.Join(baseDir, assetURL.RequestURI())
	dir := filepath.Dir(filename)

	err := f.fs.MkdirAll(dir, os.FileMode(0755))
	if err != nil {
		return fmt.Errorf("failed to create assets directory: %w", err)
	}

	file, err := f.fs.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(0644))
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}

	err = f.downloader.Download(ctx, assetURL.String(), file)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}

	return nil
}

func findAndUpdateAssetPaths(reader io.Reader, assetPathCh chan<- *url.URL, assetDirname string) (string, error) {
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return "", fmt.Errorf("failed to parse html file: %w", err)
	}

	findAndUpdateAttr := func(s *goquery.Selection, attr string) {
		src, ok := s.Attr(attr)
		if !ok {
			return
		}
		u, err := url.Parse(src)
		if err != nil || u.IsAbs() {
			return
		}
		assetPathCh <- u
		newPath := filepath.Join(assetDirname, u.RequestURI())
		s.SetAttr(attr, newPath)
	}

	doc.Find("img").Each(func(i int, s *goquery.Selection) {
		findAndUpdateAttr(s, "src")
		// TODO: suport srcset
	})
	doc.Find("link").Each(func(i int, s *goquery.Selection) {
		findAndUpdateAttr(s, "href")
	})
	doc.Find("script").Each(func(i int, s *goquery.Selection) {
		findAndUpdateAttr(s, "src")
	})

	html, err := doc.Html()
	if err != nil {
		return "", fmt.Errorf("failed to generate manipulated html: %w", err)
	}

	return html, nil
}
