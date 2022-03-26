package main

import (
	"context"
	"fmt"
	"net/http"

	fetchwebpage "github.com/izumin5210-sandbox/fetch-webpage"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use: "fetch-webpage",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		fetcher := fetchwebpage.NewFetcher(afero.NewOsFs(), new(http.Client))
		err := fetcher.Fetch(ctx, args[0])
		if err != nil {
			return fmt.Errorf("failed to fetch web page: %w", err)
		}

		return nil
	},
}