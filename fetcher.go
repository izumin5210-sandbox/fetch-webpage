package fetchwebpage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/spf13/afero"
)

type Fetcher interface {
	Fetch(ctx context.Context, url string) error
}

func NewFetcher(fs afero.Fs, httpClient *http.Client) Fetcher {
	return &fetcherImpl{
		httpClient: httpClient,
		fs:         fs,
	}
}

type fetcherImpl struct {
	httpClient *http.Client
	fs         afero.Fs
}

func (f *fetcherImpl) Fetch(ctx context.Context, givenURL string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, givenURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create new http request: %w", err)
	}
	resp, err := f.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send http request: %w", err)
	}
	defer resp.Body.Close()

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

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
