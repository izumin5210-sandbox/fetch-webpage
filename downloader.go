package fetchwebpage

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

type Downloader interface {
	Download(ctx context.Context, url string, w io.Writer) error
}

func NewDownloader(httpClient *http.Client) Downloader {
	return &downloaderImpl{
		httpClient: httpClient,
	}
}

type downloaderImpl struct {
	httpClient *http.Client
}

func (d *downloaderImpl) Download(ctx context.Context, url string, w io.Writer) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create new http request: %w", err)
	}
	resp, err := d.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send http request: %w", err)
	}
	defer resp.Body.Close()

	_, err = io.Copy(w, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}