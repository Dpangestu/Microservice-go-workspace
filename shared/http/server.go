package http

import (
	"net"
	"net/http"
	"time"
)

type ServerOptions struct {
	Addr         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	Handler      http.Handler
}

func NewServer(opt ServerOptions) *http.Server {
	return &http.Server{
		Addr:         normalizeAddr(opt.Addr),
		Handler:      opt.Handler,
		ReadTimeout:  opt.ReadTimeout,
		WriteTimeout: opt.WriteTimeout,
		IdleTimeout:  opt.IdleTimeout,
	}
}

func normalizeAddr(addr string) string {
	if _, _, err := net.SplitHostPort(addr); err != nil {
		return ":" + addr
	}
	return addr
}
