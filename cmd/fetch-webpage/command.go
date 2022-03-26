package main

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/hashicorp/go-multierror"
	fetchwebpage "github.com/izumin5210-sandbox/fetch-webpage"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	rootCmd *cobra.Command
	verbose bool
	debug   bool
)

func init() {
	rootCmd = &cobra.Command{
		Use: "fetch-webpage",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := initializeLogger()
			if err != nil {
				return fmt.Errorf("failed to initialize logger: %w", err)
			}

			return nil
		},
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
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "prints logs")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "prints more logs")
}

func initializeLogger() error {
	switch {
	case debug:
		cfg := zap.NewProductionConfig()
		cfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
		cfg.DisableStacktrace = true
		cfg.EncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Local().Format("2006-01-02 15:04:05 MST"))
		}
		zapLogger, err := cfg.Build()
		if err != nil {
			return fmt.Errorf("failed to build logger: %w", err)
		}
		zap.ReplaceGlobals(zapLogger)
	case verbose:
		cfg := zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		cfg.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
		zapLogger, err := cfg.Build()
		if err != nil {
			return fmt.Errorf("failed to build logger: %w", err)
		}
		zap.ReplaceGlobals(zapLogger)
	default:
		zap.ReplaceGlobals(zap.NewNop())
	}
	return nil
}
