package fetchwebpage

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"go.uber.org/zap"
	"golang.org/x/sync/semaphore"
)

type Downloader interface {
	Download(ctx context.Context, url string, w io.Writer) error
}

func NewDownloader(httpClient *http.Client, parallelism int) Downloader {
	return &downloaderImpl{
		httpClient: httpClient,
		sem:        semaphore.NewWeighted(int64(parallelism)),
	}
}

type downloaderImpl struct {
	httpClient *http.Client
	sem        *semaphore.Weighted
}

func (d *downloaderImpl) Download(ctx context.Context, url string, w io.Writer) error {
	err := d.sem.Acquire(ctx, 1)
	if err != nil {
		return fmt.Errorf("failed to acquire semaphore: %w", err)
	}
	defer d.sem.Release(1)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create new http request: %w", err)
	}

	zap.L().Debug("Download is started", zap.String("url", url))
	resp, err := d.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send http request: %w", err)
	}
	defer resp.Body.Close()

	_, err = io.Copy(w, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	zap.L().Debug("Download is finished", zap.String("url", url))

	return nil
}
