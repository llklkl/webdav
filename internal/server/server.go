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

package server

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/webdav"

	"github.com/llklkl/webdav/conf"
	"github.com/llklkl/webdav/internal/fs"
	"github.com/llklkl/webdav/internal/middleware"
	"github.com/llklkl/webdav/internal/pkg"
)

type Server struct {
	cfg      *conf.Conf
	httpSvr  *http.Server
	httpsSvr *http.Server

	mux         *http.ServeMux
	middleWares []middleware.MiddleWare
}

func NewServer(cfg *conf.Conf) (*Server, error) {
	s := &Server{
		cfg: cfg,
	}
	err := s.buildServer()
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Server) buildServer() error {
	s.buildWebdavHandler()

	if err := s.buildHttpServer(); err != nil {
		return err
	}
	if err := s.buildHttpsServer(); err != nil {
		return err
	}

	return nil
}

func (s *Server) buildWebdavHandler() {
	for i := 0; i < len(s.cfg.Library); i++ {
		for j := i + 1; j < len(s.cfg.Library); j++ {
			pi := filepath.Clean(s.cfg.Library[i].MountPoint)
			pj := filepath.Clean(s.cfg.Library[j].MountPoint)
			if strings.HasPrefix(pi, pj) || strings.HasPrefix(pj, pi) {
				slog.Warn("mount point should not be prefix of other mount point",
					slog.String("a", pj),
					slog.String("b", pi),
				)
			}
		}
	}

	s.mux = http.NewServeMux()
	for _, lib := range s.cfg.Library {
		prefix := cleanPrefix(lib.Prefix)
		s.mux.Handle(prefix, &webdav.Handler{
			Prefix:     "",
			FileSystem: fs.NewFs(s.cfg, lib),
			LockSystem: webdav.NewMemLS(),
			Logger:     nil,
		})
	}

	s.middleWares = middleware.NewMiddleWares(s.cfg)
}

func (s *Server) buildHttpsServer() error {
	if !s.cfg.HttpsEnable {
		return nil
	}

	certs := make([]tls.Certificate, 1)
	if s.cfg.TlsKeyPem != "" && s.cfg.TlsCertPem != "" {
		keyPem, err := base64.StdEncoding.DecodeString(s.cfg.TlsKeyPem)
		if err != nil {
			return fmt.Errorf("the private key is not in the correct base64 encoding: %w", err)
		}
		certPem, err := base64.StdEncoding.DecodeString(s.cfg.TlsCertPem)
		if err != nil {
			return fmt.Errorf("the cert is not in the correct base64 encoding: %w", err)
		}
		cert, err := tls.X509KeyPair(certPem, keyPem)
		if err != nil {
			return fmt.Errorf("load tls cert and key error: %w", err)
		}
		certs[0] = cert
	} else if s.cfg.TlsCertPemPath != "" && s.cfg.TlsKeyPemPath != "" {
		cert, err := tls.LoadX509KeyPair(s.cfg.TlsCertPemPath, s.cfg.TlsKeyPemPath)
		if err != nil {
			return fmt.Errorf("load tls cert and key error: %w", err)
		}
		certs[0] = cert
	} else {
		return errors.New("the TLS certificate is not specified")
	}

	s.httpsSvr = &http.Server{
		Addr: s.cfg.HttpsListen,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var h http.Handler
			h = s.mux
			for i := range s.middleWares {
				h = s.middleWares[i].Serve(h)
			}
			h.ServeHTTP(w, r)
		}),
		TLSConfig: &tls.Config{
			Certificates: certs,
		},
	}

	return nil
}

func (s *Server) buildHttpServer() error {
	if !s.cfg.HttpEnable {
		return nil
	}

	s.httpSvr = &http.Server{
		Addr: s.cfg.HttpListen,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var h http.Handler
			h = s.mux
			for i := len(s.middleWares) - 1; i >= 0; i-- {
				h = s.middleWares[i].Serve(h)
			}
			h.ServeHTTP(w, r)
		}),
	}

	return nil
}

func cleanPrefix(prefix string) string {
	prefix = strings.Trim(prefix, "/")
	prefix = "/" + prefix
	if prefix == "/" {
		return prefix
	}

	return prefix + "/"
}

func (s *Server) Start() {
	if s.httpSvr != nil {
		pkg.SafeG(func() {
			slog.Info("starting webdav server",
				slog.String("addr", fmt.Sprintf("http://%s", s.cfg.HttpListen)))
			err := s.httpSvr.ListenAndServe()
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				slog.Error("http server exited unexpectedly", slog.Any("err", err))
			}
		})
	}

	if s.httpsSvr != nil {
		pkg.SafeG(func() {
			slog.Info("starting webdav server",
				slog.String("addr", fmt.Sprintf("https://%s", s.cfg.HttpsListen)))
			err := s.httpsSvr.ListenAndServeTLS("", "")
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				slog.Error("https server exited unexpectedly", slog.Any("err", err))
			}
		})
	}
}

func (s *Server) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	wg := &sync.WaitGroup{}
	wg.Add(2)
	pkg.SafeG(func() {
		defer wg.Done()
		if s.httpSvr != nil {
			_ = s.httpSvr.Shutdown(ctx)
		}
	})
	pkg.SafeG(func() {
		defer wg.Done()
		if s.httpsSvr != nil {
			_ = s.httpsSvr.Shutdown(ctx)
		}
	})

	wg.Wait()
}
