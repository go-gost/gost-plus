package entrypoint

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-gost/core/chain"
	"github.com/go-gost/core/handler"
	"github.com/go-gost/core/listener"
	"github.com/go-gost/core/logger"
	"github.com/go-gost/core/service"
	cfg "github.com/go-gost/gost.plus/config"
	"github.com/go-gost/gost.plus/tunnel"
	"github.com/go-gost/x/config"
	chain_parser "github.com/go-gost/x/config/parsing/chain"
	"github.com/go-gost/x/handler/forward/local"
	"github.com/go-gost/x/hop"
	"github.com/go-gost/x/listener/tcp"
	mdx "github.com/go-gost/x/metadata"
	xservice "github.com/go-gost/x/service"
	"github.com/go-gost/x/stats"
	"github.com/google/uuid"
)

type tcpEntryPoint struct {
	endpoint string
	opts     tunnel.Options
	config   *config.Config
	forward  service.Service
	favorite atomic.Bool
	stats    cfg.ServiceStats

	cclose chan struct{}

	err error
	mu  sync.RWMutex
}

func NewTCPEntryPoint(opts ...tunnel.Option) EntryPoint {
	var options tunnel.Options
	for _, opt := range opts {
		opt(&options)
	}

	if options.ID == "" {
		options.ID = uuid.NewString()
	}

	v := md5.Sum([]byte(options.ID))
	endpoint := hex.EncodeToString(v[:8])

	if options.Endpoint == "" {
		options.Endpoint = "localhost:8000"
	}

	if options.Name == "" {
		options.Name = endpoint
	}
	if options.CreatedAt.IsZero() {
		options.CreatedAt = time.Now()
	}

	s := &tcpEntryPoint{
		endpoint: endpoint,
		opts:     options,
		cclose:   make(chan struct{}),
	}

	return s
}

func (s *tcpEntryPoint) ID() string {
	return s.opts.ID
}

func (s *tcpEntryPoint) Type() string {
	return TCPEntryPoint
}

func (s *tcpEntryPoint) Name() string {
	return s.opts.Name
}

func (s *tcpEntryPoint) Endpoint() string {
	return fmt.Sprintf("%s.%s", s.endpoint, tunnel.EndpointAddr)
}

func (s *tcpEntryPoint) Entrypoint() string {
	return s.opts.Endpoint
}

func (s *tcpEntryPoint) Options() tunnel.Options {
	return s.opts
}

func (s *tcpEntryPoint) Favorite(b bool) {
	s.favorite.Store(b)
}

func (s *tcpEntryPoint) IsFavorite() bool {
	return s.favorite.Load()
}

func (s *tcpEntryPoint) init() error {
	tcp := &config.ServiceConfig{
		Name: s.opts.Name,
		Addr: s.opts.Endpoint,
		Handler: &config.HandlerConfig{
			Type:  "tcp",
			Chain: s.opts.Name,
		},
		Listener: &config.ListenerConfig{
			Type: "tcp",
		},
		Forwarder: &config.ForwarderConfig{
			Nodes: []*config.ForwardNodeConfig{
				{
					Name: s.opts.Name,
					Addr: s.Endpoint(),
				},
			},
		},
	}

	s.config = &config.Config{
		Services: []*config.ServiceConfig{tcp},
		Chains:   []*config.ChainConfig{tunnel.ChainConfig(s.opts.ID, s.opts.Name)},
	}
	return nil
}

func (s *tcpEntryPoint) Run() (err error) {
	if s.IsClosed() {
		return ErrEntryPointClosed
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

		stats := &stats.Stats{}

		cfg := s.config.Services[0]
		ln := tcp.NewListener(
			listener.AddrOption(cfg.Addr),
			listener.LoggerOption(log.WithFields(map[string]any{"kind": "listener", "listener": "tcp"})),
			listener.StatsOption(stats),
		)
		if err = ln.Init(mdx.NewMetadata(cfg.Listener.Metadata)); err != nil {
			return
		}

		h := local.NewHandler(
			handler.RouterOption(chain.NewRouter(
				chain.ChainRouterOption(ch),
				chain.LoggerRouterOption(log),
			)),
			handler.LoggerOption(log.WithFields(map[string]any{"kind": "handler", "handler": "tcp"})),
		)
		if err = h.Init(mdx.NewMetadata(cfg.Handler.Metadata)); err != nil {
			return
		}

		node := cfg.Forwarder.Nodes[0]
		if forwarder, ok := h.(handler.Forwarder); ok {
			forwarder.Forward(hop.NewHop(
				hop.NodeOption(chain.NewNode(node.Name, node.Addr)),
				hop.LoggerOption(log.WithFields(map[string]any{"kind": "hop"})),
			))
		}
		s.forward = xservice.NewService(s.opts.Name, ln, h,
			xservice.LoggerOption(log),
			xservice.StatsOption(stats),
		)
	}

	go func() {
		s.setErr(s.forward.Serve())
	}()

	return nil
}

func (s *tcpEntryPoint) Status() *xservice.Status {
	if ss, _ := s.forward.(tunnel.ServiceStatus); ss != nil {
		return ss.Status()
	}
	return nil
}

func (s *tcpEntryPoint) Stats() cfg.ServiceStats {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.stats
}

func (s *tcpEntryPoint) SetStats(stats cfg.ServiceStats) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stats = stats
}

func (s *tcpEntryPoint) Close() error {
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

func (s *tcpEntryPoint) IsClosed() bool {
	select {
	case <-s.cclose:
		return true
	default:
		return false
	}
}

func (s *tcpEntryPoint) setErr(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.err = err
}

func (s *tcpEntryPoint) Err() error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.err
}
