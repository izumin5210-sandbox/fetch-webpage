package fetchwebpage

import (
	"bytes"
	"context"
	"fmt"
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

	filename := fmt.Sprintf("%s.html", parsedURL.Host)

	var buf bytes.Buffer
	err = f.downloader.Download(ctx, givenURL, &buf)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}

	assetDirname := fmt.Sprintf("%s_assets", filename)

	downloadAsset := func(path *url.URL) error {
		assetURL := parsedURL.ResolveReference(path)

		filename := filepath.Join(assetDirname, assetURL.RequestURI())
		dir := filepath.Dir(filename)

		err = f.fs.MkdirAll(dir, os.FileMode(0755))
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
			err := downloadAsset(path)
			if err != nil {
				errCh <- fmt.Errorf("failed to download asset: %w", err)
			}
		}
	}()

	doc, err := goquery.NewDocumentFromReader(&buf)
	if err != nil {
		return fmt.Errorf("failed to parse html file: %w", err)
	}
	doc.Find("img").Each(func(i int, s *goquery.Selection) {
		if src, ok := s.Attr("src"); ok {
			if u, err := url.Parse(src); err == nil && !u.IsAbs() {
				assetPathCh <- u
				newPath := filepath.Join(assetDirname, u.RequestURI())
				s.SetAttr("src", newPath)
			}
		}
	})

	close(assetPathCh)
	close(errCh)
	wg.Wait()

	html, err := doc.Html()
	if err != nil {
		return fmt.Errorf("failed to generate manipulated html: %w", err)
	}

	err = afero.WriteFile(f.fs, filename, []byte(html), os.FileMode(0664))
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
