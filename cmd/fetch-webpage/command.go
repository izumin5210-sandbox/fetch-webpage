package main

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/hashicorp/go-multierror"
	fetchwebpage "github.com/izumin5210-sandbox/fetch-webpage"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use: "fetch-webpage",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		downloader := fetchwebpage.NewDownloader(new(http.Client))
		fetcher := fetchwebpage.NewFetcher(downloader, afero.NewOsFs())

		errCh := make(chan error, len(args))

		var wg sync.WaitGroup

		for _, arg := range args {
			url := arg
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := fetcher.Fetch(ctx, url)
				if err != nil {
					errCh <- fmt.Errorf("failed to fetch web page: %w", err)
				}
			}()
		}

		wg.Wait()
		close(errCh)

		var combinedErr error
		for err := range errCh {
			combinedErr = multierror.Append(combinedErr, err)
		}

		return combinedErr
	},
}
