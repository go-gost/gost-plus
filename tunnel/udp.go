package tunnel

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"sync/atomic"

	"github.com/go-gost/core/chain"
	"github.com/go-gost/core/handler"
	"github.com/go-gost/core/listener"
	"github.com/go-gost/core/logger"
	"github.com/go-gost/core/service"
	"github.com/go-gost/x/config"
	chain_parser "github.com/go-gost/x/config/parsing/chain"
	"github.com/go-gost/x/handler/forward/remote"
	"github.com/go-gost/x/hop"
	"github.com/go-gost/x/listener/rudp"
	mdx "github.com/go-gost/x/metadata"
	xservice "github.com/go-gost/x/service"
	"github.com/google/uuid"
)

type udpTunnel struct {
	endpoint string
	opts     Options
	config   *config.Config
	forward  service.Service
	favorite atomic.Bool

	cclose chan struct{}
}

func NewUDPTunnel(opts ...Option) Tunnel {
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
		options.Endpoint = "localhost:8080"
	}

	if options.Name == "" {
		options.Name = fmt.Sprintf("UDP-%s", endpoint)
	}

	s := &udpTunnel{
		endpoint: endpoint,
		opts:     options,
		cclose:   make(chan struct{}),
	}

	return s
}

func (s *udpTunnel) ID() string {
	return s.opts.ID
}

func (s *udpTunnel) Type() string {
	return UDPTunnel
}

func (s *udpTunnel) Name() string {
	return s.opts.Name
}

func (s *udpTunnel) Endpoint() string {
	return s.opts.Endpoint
}

func (f *udpTunnel) Entrypoint() string {
	return fmt.Sprintf("%s.%s", f.endpoint, endpointAddr)
}

func (s *udpTunnel) Options() Options {
	return s.opts
}

func (s *udpTunnel) Favorite(b bool) {
	s.favorite.Store(b)
}

func (s *udpTunnel) IsFavorite() bool {
	return s.favorite.Load()
}

func (s *udpTunnel) init() error {
	rudp := &config.ServiceConfig{
		Name: s.opts.Name,
		Addr: s.opts.Hostname,
		Handler: &config.HandlerConfig{
			Type: "rudp",
		},
		Listener: &config.ListenerConfig{
			Type:  "rudp",
			Chain: s.opts.Name,
		},
		Forwarder: &config.ForwarderConfig{
			Nodes: []*config.ForwardNodeConfig{
				{
					Name: s.opts.Name,
					Addr: s.opts.Endpoint,
				},
			},
		},
	}

	s.config = &config.Config{
		Services: []*config.ServiceConfig{rudp},
		Chains:   []*config.ChainConfig{chainConfig(s.opts.ID, s.opts.Name)},
	}
	return nil
}

func (s *udpTunnel) Run() error {
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
		ch, err := chain_parser.ParseChain(s.config.Chains[0], log)
		if err != nil {
			log.Error(err)
			return err
		}

		cfg := s.config.Services[0]
		ln := rudp.NewListener(
			listener.AddrOption(cfg.Addr),
			listener.ChainOption(ch),
			listener.LoggerOption(log.WithFields(map[string]any{"kind": "listener", "listener": "rudp"})),
		)
		if err := ln.Init(mdx.NewMetadata(cfg.Listener.Metadata)); err != nil {
			return err
		}

		h := remote.NewHandler(
			handler.LoggerOption(log.WithFields(map[string]any{"kind": "handler", "handler": "rudp"})),
		)
		if err := h.Init(mdx.NewMetadata(cfg.Handler.Metadata)); err != nil {
			return err
		}

		node := cfg.Forwarder.Nodes[0]
		if forwarder, ok := h.(handler.Forwarder); ok {
			forwarder.Forward(hop.NewHop(
				hop.NodeOption(chain.NewNode(node.Name, node.Addr)),
				hop.LoggerOption(log.WithFields(map[string]any{"kind": "hop"})),
			))
		}
		s.forward = xservice.NewService(s.opts.Name, ln, h, xservice.LoggerOption(log))
	}

	go s.forward.Serve()

	return nil
}

func (s *udpTunnel) Close() error {
	defer func() {
		select {
		case <-s.cclose:
		default:
			close(s.cclose)
		}
	}()

	if s.forward != nil {
		return s.forward.Close()
	}
	return nil
}

func (s *udpTunnel) IsClosed() bool {
	select {
	case <-s.cclose:
		return true
	default:
		return false
	}
}
