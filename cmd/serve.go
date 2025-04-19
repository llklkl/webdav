/*
 * Copyright (c) 2024 llklkl
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *
 */

package cmd

import (
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/llklkl/webdav/conf"
	"github.com/llklkl/webdav/internal/pkg"
	"github.com/llklkl/webdav/internal/server"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "start a webdav server",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		confPath, _ := cmd.Flags().GetString("conf")

		setupLog(cmd)

		cfg, err := conf.Parse(confPath)
		if err != nil {
			slog.Error("failed to parse configure file", slog.Any("err", err))
			return
		}

		svr, err := server.NewServer(cfg)
		if err != nil {
			slog.Error("failed to start webdav server", slog.Any("err", err))
			return
		}
		svr.Start()

		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigs
		slog.Warn("received signal, quit", slog.String("signal", sig.String()))
		svr.Stop()
		time.Sleep(time.Millisecond * 100)
	},
	PostRun: func(cmd *cobra.Command, args []string) {
		for i := range postRunFunc {
			pkg.SafeRun(postRunFunc[i])
		}
	},
}

func setupLog(cmd *cobra.Command) {
	level := slog.LevelError
	if levels, err := cmd.Flags().GetString("log"); err == nil {
		levels = strings.ToUpper(levels)
		switch levels {
		case "DEBUG":
			level = slog.LevelDebug
		case "INFO":
			level = slog.LevelInfo
		case "WARN":
			level = slog.LevelWarn
		}
	}

	var fp = os.Stdout
	if file, err := cmd.Flags().GetString("log-file"); err == nil {
		f, err := os.OpenFile(file, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err == nil {
			fp = f
			postRunFunc = append(postRunFunc, func() {
				_ = f.Sync()
				_ = f.Close()
			})
		}
	}

	handler := slog.NewJSONHandler(fp, &slog.HandlerOptions{Level: level, AddSource: true})
	slog.SetDefault(slog.New(handler))
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().StringP("conf", "c", "", "path to configure file")
	serveCmd.Flags().StringP("log", "", "ERROR", "log level, support: DEBUG, INFO, WARN, ERROR")
	serveCmd.Flags().StringP("log-file", "", "", "path to log file, default: stdout")

	serveCmd.MarkFlagRequired("conf")
}
