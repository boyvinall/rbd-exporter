// Package main implements the main entrypoint for the rbd-exporter
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	cli "github.com/urfave/cli/v2"

	"github.com/boyvinall/rbd-exporter/pkg/collector"
)

func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	_, _ = w.Write([]byte(`<html>
		<head><title>RBD Exporter</title></head>
		<body>
		<h1>RBD Exporter</h1>
		<p><a href="/metrics">Metrics</a></p>
		</body>
		</html>`))
}

func serve(listenAddress string, pools []string) error {

	collector := collector.New(pools, &collector.RBDMirrorPoolStatus{})
	err := prometheus.Register(collector)
	if err != nil {
		return err
	}

	// serve http

	server := &http.Server{
		Addr:              listenAddress,
		ReadHeaderTimeout: 3 * time.Second,
	}
	errCh := make(chan error)
	go func() {
		http.HandleFunc("/", rootHandler)
		http.Handle("/metrics", promhttp.Handler())
		slog.Info("serving", "address", listenAddress)
		errCh <- server.ListenAndServe()
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// wait for error or signal

	slog.Info("waiting for error or signal")
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs,
		os.Interrupt,    // CTRL-C
		syscall.SIGTERM, // e.g. docker graceful shutdown
	)

	select {
	case err = <-errCh:
	case <-ctx.Done():
		err = ctx.Err()
	case s := <-sigs:
		switch s {
		case os.Interrupt:
			slog.Info("Interrupt: CTRL-C")
		case syscall.SIGTERM:
			slog.Info("SIGTERM")
		default:
			slog.Info("Received signal", "signal", s.String())
		}
	}

	_ = server.Shutdown(ctx)
	cancel()

	return err
}

func main() {
	app := cli.NewApp()
	app.Name = "rbd-exporter"
	app.Usage = "Prometheus exporter for Ceph RBD"
	app.Description = strings.Join([]string{}, "\n")
	app.EnableBashCompletion = true
	app.CommandNotFound = func(c *cli.Context, cmd string) {
		fmt.Printf("ERROR: Unknown command '%s'\n", cmd)
	}
	app.Commands = []*cli.Command{
		{
			Name:        "serve",
			Usage:       "Start the exporter",
			Description: strings.Join([]string{}, "\n"),
			Action: func(c *cli.Context) error {
				programLevel := slog.LevelInfo
				if c.Bool("verbose") {
					programLevel = slog.LevelDebug
				}
				h := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: programLevel})
				slog.SetDefault(slog.New(h))
				return serve(c.String("listen-address"), c.StringSlice("pool"))
			},
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "listen-address",
					Usage: "Address to listen on for web interface and telemetry",
					Value: ":9876",
				},
				&cli.StringSliceFlag{
					Name:  "pool",
					Usage: "Ceph pools to monitor",
					Value: cli.NewStringSlice(),
				},
			},
		},
		// {
		// 	Name:        "once",
		// 	Usage:       "run the checks once and exit",
		// 	Description: strings.Join([]string{}, "\n"),
		// 	Action: func(c *cli.Context) error {
		// 		managers, err := createManagers(c.String("settings-file"), c.StringSlice("cloud")...)
		// 		if err != nil {
		// 			return err
		// 		}
		// 		return once(managers, c.Args().Slice())
		// 	},
		// },
	}
	app.Flags = []cli.Flag{
		&cli.BoolFlag{
			Name:    "verbose",
			Aliases: []string{"v"},
			Usage:   "Increase verbosity",
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		panic(err)
	}
}
