package tunnel

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"sync/atomic"

	"github.com/go-gost/core/auth"
	"github.com/go-gost/core/chain"
	"github.com/go-gost/core/handler"
	"github.com/go-gost/core/listener"
	"github.com/go-gost/core/logger"
	"github.com/go-gost/core/service"
	xauth "github.com/go-gost/x/auth"
	"github.com/go-gost/x/config"
	chain_parser "github.com/go-gost/x/config/parsing/chain"
	"github.com/go-gost/x/handler/file"
	"github.com/go-gost/x/handler/forward/remote"
	"github.com/go-gost/x/hop"
	"github.com/go-gost/x/listener/rtcp"
	"github.com/go-gost/x/listener/tcp"
	mdx "github.com/go-gost/x/metadata"
	xservice "github.com/go-gost/x/service"
	"github.com/google/uuid"
)

type fileTunnel struct {
	endpoint string
	opts     Options
	config   *config.Config
	file     service.Service
	forward  service.Service
	favorite atomic.Bool

	cclose chan struct{}
}

func NewFileTunnel(opts ...Option) Tunnel {
	var options Options
	for _, opt := range opts {
		opt(&options)
	}

	if options.ID == "" {
		options.ID = uuid.NewString()
	}

	v := md5.Sum([]byte(options.ID))
	endpoint := hex.EncodeToString(v[:8])

	if options.Endpoint == "" {
		options.Endpoint, _ = os.Getwd()
	}

	if options.Name == "" {
		options.Name = fmt.Sprintf("FILE-%s", endpoint)
	}

	s := &fileTunnel{
		endpoint: endpoint,
		opts:     options,
		cclose:   make(chan struct{}),
	}

	return s
}

func (s *fileTunnel) ID() string {
	return s.opts.ID
}

func (s *fileTunnel) Type() string {
	return FileTunnel
}

func (s *fileTunnel) Name() string {
	return s.opts.Name
}

func (s *fileTunnel) Endpoint() string {
	return s.opts.Endpoint
}

func (f *fileTunnel) Entrypoint() string {
	return fmt.Sprintf("https://%s.%s", f.endpoint, endpointAddr)
}

func (s *fileTunnel) Options() Options {
	return s.opts
}

func (s *fileTunnel) Favorite(b bool) {
	s.favorite.Store(b)
}

func (s *fileTunnel) IsFavorite() bool {
	return s.favorite.Load()
}

func (s *fileTunnel) init() error {
	file := &config.ServiceConfig{
		Name: s.opts.Name,
		Addr: ":0",
		Handler: &config.HandlerConfig{
			Type:     "file",
			Metadata: map[string]any{"file.dir": s.opts.Endpoint},
		},
		Listener: &config.ListenerConfig{
			Type: "tcp",
		},
	}
	if s.opts.Username != "" {
		file.Handler.Auth = &config.AuthConfig{
			Username: s.opts.Username,
			Password: s.opts.Password,
		}
	}

	rtcp := &config.ServiceConfig{
		Name: s.opts.Name,
		Addr: s.opts.Hostname,
		Handler: &config.HandlerConfig{
			Type: "rtcp",
		},
		Listener: &config.ListenerConfig{
			Type:  "rtcp",
			Chain: s.opts.Name,
		},
	}

	s.config = &config.Config{
		Services: []*config.ServiceConfig{file, rtcp},
		Chains:   []*config.ChainConfig{chainConfig(s.opts.ID, s.opts.Name)},
	}

	return nil
}

func (s *fileTunnel) Run() error {
	if s.IsClosed() {
		return ErrTunnelClosed
	}

	if err := s.init(); err != nil {
		return err
	}

	log := logger.Default().WithFields(map[string]any{
		"kind":    "service",
		"service": s.opts.Name,
	})

	{
		cfg := s.config.Services[0]
		ln := tcp.NewListener(
			listener.LoggerOption(log.WithFields(map[string]any{"kind": "listener", "listener": "tcp"})),
		)
		if err := ln.Init(nil); err != nil {
			return err
		}
		log.Infof("listen on %s", ln.Addr())

		var auther auth.Authenticator
		if auth := cfg.Handler.Auth; auth != nil {
			auther = xauth.NewAuthenticator(xauth.AuthsOption(map[string]string{auth.Username: auth.Password}))
		}
		h := file.NewHandler(
			handler.LoggerOption(log.WithFields(map[string]any{"kind": "handler", "handler": "file"})),
			handler.AutherOption(auther),
		)
		if err := h.Init(mdx.NewMetadata(cfg.Handler.Metadata)); err != nil {
			return err
		}
		s.file = xservice.NewService(s.opts.Name, ln, h, xservice.LoggerOption(log))
	}

	{
		ch, err := chain_parser.ParseChain(s.config.Chains[0], log)
		if err != nil {
			log.Error(err)
			return err
		}

		cfg := s.config.Services[1]
		ln := rtcp.NewListener(
			listener.AddrOption(cfg.Addr),
			listener.ChainOption(ch),
			listener.LoggerOption(log.WithFields(map[string]any{"kind": "listener", "listener": "rtcp"})),
		)
		if err := ln.Init(mdx.NewMetadata(cfg.Listener.Metadata)); err != nil {
			return err
		}

		h := remote.NewHandler(
			handler.LoggerOption(log.WithFields(map[string]any{"kind": "handler", "handler": "rtcp"})),
		)
		if err := h.Init(mdx.NewMetadata(cfg.Handler.Metadata)); err != nil {
			return err
		}
		if forwarder, ok := h.(handler.Forwarder); ok {
			forwarder.Forward(hop.NewHop(
				hop.NodeOption(chain.NewNode(s.opts.Name, s.file.Addr().String())),
				hop.LoggerOption(log.WithFields(map[string]any{"kind": "hop"})),
			))
		}
		s.forward = xservice.NewService(s.opts.Name, ln, h, xservice.LoggerOption(log))
		log.Infof("service listen on %s", s.file.Addr())
	}

	go s.file.Serve()
	go s.forward.Serve()

	log.Infof("file service run at %s", s.file.Addr())
	return nil
}

func (s *fileTunnel) Close() error {
	defer func() {
		select {
		case <-s.cclose:
		default:
			close(s.cclose)
		}
	}()

	if s.forward != nil {
		s.forward.Close()
	}
	if s.file != nil {
		return s.file.Close()
	}
	return nil
}

func (s *fileTunnel) IsClosed() bool {
	select {
	case <-s.cclose:
		return true
	default:
		return false
	}
}
