package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"syscall"
	"time"

	"github.com/USA-RedDragon/rtz-server/internal/config"
	"github.com/USA-RedDragon/rtz-server/internal/db"
	"github.com/USA-RedDragon/rtz-server/internal/server"
	"github.com/spf13/cobra"
	"github.com/ztrue/shutdown"
	"golang.org/x/sync/errgroup"
)

func NewCommand(version, commit string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "rtz-server",
		Version: fmt.Sprintf("%s - %s", version, commit),
		Annotations: map[string]string{
			"version": version,
			"commit":  commit,
		},
		RunE:          run,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	config.RegisterFlags(cmd)
	return cmd
}

func run(cmd *cobra.Command, _ []string) error {
	slog.Info("rtz-server", "version", cmd.Annotations["version"], "commit", cmd.Annotations["commit"])

	config, err := config.LoadConfig(cmd)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	err = config.Validate()
	if err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	// Check access to the upload directory
	err = os.MkdirAll(config.Persistence.Uploads, 0755)
	if err != nil {
		return fmt.Errorf("failed to create uploads directory: %w", err)
	}

	db, err := db.MakeDB(config)
	if err != nil {
		return fmt.Errorf("failed to make database: %w", err)
	}
	slog.Info("Database connection established")

	slog.Info("Starting HTTP server")
	server := server.NewServer(config, db)
	err = server.Start()
	if err != nil {
		return fmt.Errorf("failed to start HTTP server: %w", err)
	}

	stop := func(sig os.Signal) {
		slog.Info("Shutting down")

		errGrp := errgroup.Group{}

		errGrp.Go(func() error {
			return server.Stop()
		})

		err := errGrp.Wait()
		if err != nil {
			slog.Error("Shutdown error", "error", err.Error())
		}
		slog.Info("Shutdown complete")
	}

	if cmd.Annotations["version"] == "testing" {
		doneChannel := make(chan struct{})
		go func() {
			slog.Info("Sleeping for 5 seconds")
			time.Sleep(5 * time.Second)
			slog.Info("Sending SIGTERM")
			stop(syscall.SIGTERM)
			doneChannel <- struct{}{}
		}()
		<-doneChannel
	} else {
		shutdown.AddWithParam(stop)
		shutdown.Listen(syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM, syscall.SIGQUIT)
	}

	return nil
}
