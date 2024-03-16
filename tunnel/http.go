package tunnel

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/go-gost/core/chain"
	"github.com/go-gost/core/handler"
	"github.com/go-gost/core/listener"
	"github.com/go-gost/core/logger"
	"github.com/go-gost/core/service"
	xauth "github.com/go-gost/x/auth"
	"github.com/go-gost/x/config"
	chain_parser "github.com/go-gost/x/config/parsing/chain"
	"github.com/go-gost/x/handler/forward/remote"
	"github.com/go-gost/x/hop"
	"github.com/go-gost/x/listener/rtcp"
	mdx "github.com/go-gost/x/metadata"
	xservice "github.com/go-gost/x/service"
	"github.com/google/uuid"
)

type httpTunnel struct {
	endpoint string
	opts     Options
	config   *config.Config
	forward  service.Service
	favorite atomic.Bool

	cclose chan struct{}

	err error
	mu  sync.RWMutex
}

func NewHTTPTunnel(opts ...Option) Tunnel {
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
		options.Name = endpoint
	}

	s := &httpTunnel{
		endpoint: endpoint,
		opts:     options,
		cclose:   make(chan struct{}),
	}

	return s
}

func (s *httpTunnel) ID() string {
	return s.opts.ID
}

func (s *httpTunnel) Type() string {
	return HTTPTunnel
}

func (s *httpTunnel) Name() string {
	return s.opts.Name
}

func (s *httpTunnel) Endpoint() string {
	return s.opts.Endpoint
}

func (s *httpTunnel) Entrypoint() string {
	return fmt.Sprintf("https://%s.%s", s.endpoint, EndpointAddr)
}

func (s *httpTunnel) Options() Options {
	return s.opts
}

func (s *httpTunnel) Favorite(b bool) {
	s.favorite.Store(b)
}

func (s *httpTunnel) IsFavorite() bool {
	return s.favorite.Load()
}

func (s *httpTunnel) init() error {
	node := &config.ForwardNodeConfig{
		Name: s.opts.Name,
		Addr: s.opts.Endpoint,
	}
	if s.opts.Username != "" {
		node.Auth = &config.AuthConfig{
			Username: s.opts.Username,
			Password: s.opts.Password,
		}
	}
	if s.opts.Hostname != "" {
		node.HTTP = &config.HTTPNodeConfig{
			Host: s.opts.Hostname,
		}
	}
	if s.opts.EnableTLS {
		node.TLS = &config.TLSNodeConfig{}
	}

	rtcp := &config.ServiceConfig{
		Name: s.opts.Name,
		Addr: "",
		Handler: &config.HandlerConfig{
			Type: "rtcp",
			Metadata: map[string]any{
				"sniffing": true,
			},
		},
		Listener: &config.ListenerConfig{
			Type:  "rtcp",
			Chain: s.opts.Name,
		},
		Forwarder: &config.ForwarderConfig{
			Nodes: []*config.ForwardNodeConfig{node},
		},
	}

	s.config = &config.Config{
		Services: []*config.ServiceConfig{rtcp},
		Chains:   []*config.ChainConfig{ChainConfig(s.opts.ID, s.opts.Name)},
	}
	return nil
}

func (s *httpTunnel) Run() (err error) {
	if s.IsClosed() {
		return ErrTunnelClosed
	}

	defer func() {
		s.setErr(err)
	}()

	if err = s.init(); err != nil {
		return
	}

	log := logger.Default().WithFields(map[string]any{
		"kind":    "service",
		"service": s.opts.Name,
	})

	{
		var ch chain.Chainer
		ch, err = chain_parser.ParseChain(s.config.Chains[0], log)
		if err != nil {
			log.Error(err)
			return
		}

		cfg := s.config.Services[0]
		ln := rtcp.NewListener(
			listener.AddrOption(cfg.Addr),
			listener.ChainOption(ch),
			listener.LoggerOption(log.WithFields(map[string]any{"kind": "listener", "listener": "rtcp"})),
		)
		if err = ln.Init(mdx.NewMetadata(cfg.Listener.Metadata)); err != nil {
			return
		}

		h := remote.NewHandler(
			handler.LoggerOption(log.WithFields(map[string]any{"kind": "handler", "handler": "rtcp"})),
		)
		if err = h.Init(mdx.NewMetadata(cfg.Handler.Metadata)); err != nil {
			return
		}

		node := cfg.Forwarder.Nodes[0]
		var nodeOpts []chain.NodeOption
		if node.HTTP != nil {
			httpNodeSettings := &chain.HTTPNodeSettings{
				Host:   node.HTTP.Host,
				Header: node.HTTP.Header,
			}
			if node.HTTP.Auth != nil {
				httpNodeSettings.Auther = xauth.NewAuthenticator(xauth.AuthsOption(map[string]string{node.HTTP.Auth.Username: node.HTTP.Auth.Password}))
			}
			nodeOpts = append(nodeOpts, chain.HTTPNodeOption(httpNodeSettings))
		}
		if node.TLS != nil {
			nodeOpts = append(nodeOpts, chain.TLSNodeOption(&chain.TLSNodeSettings{
				ServerName: node.TLS.ServerName,
				Secure:     node.TLS.Secure,
			}))
		}
		if forwarder, ok := h.(handler.Forwarder); ok {
			forwarder.Forward(hop.NewHop(hop.NodeOption(chain.NewNode(node.Name, node.Addr, nodeOpts...)),
				hop.LoggerOption(log.WithFields(map[string]any{"kind": "hop"})),
			))
		}
		s.forward = xservice.NewService(s.opts.Name, ln, h, xservice.LoggerOption(log))
	}

	go func() {
		s.setErr(s.forward.Serve())
	}()

	return nil
}

func (s *httpTunnel) Status() *xservice.Status {
	if ss, _ := s.forward.(ServiceStatus); ss != nil {
		return ss.Status()
	}
	return nil
}

func (s *httpTunnel) Close() error {
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

func (s *httpTunnel) IsClosed() bool {
	select {
	case <-s.cclose:
		return true
	default:
		return false
	}
}

func (s *httpTunnel) setErr(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.err = err
}

func (s *httpTunnel) Err() error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.err
}
