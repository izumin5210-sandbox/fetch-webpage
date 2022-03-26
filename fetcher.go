package fetchwebpage

import (
	"context"
	"fmt"
	"net/url"
	"os"

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

	file, err := f.fs.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(0644))
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	err = f.downloader.Download(ctx, givenURL, file)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
