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
	xchain "github.com/go-gost/x/chain"
	"github.com/go-gost/x/config"
	chain_parser "github.com/go-gost/x/config/parsing/chain"
	"github.com/go-gost/x/handler/forward/local"
	"github.com/go-gost/x/hop"
	"github.com/go-gost/x/listener/udp"
	mdx "github.com/go-gost/x/metadata"
	xstats "github.com/go-gost/x/observer/stats"
	xservice "github.com/go-gost/x/service"
	"github.com/google/uuid"
)

type udpEntryPoint struct {
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

func NewUDPEntryPoint(opts ...tunnel.Option) EntryPoint {
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

	s := &udpEntryPoint{
		endpoint: endpoint,
		opts:     options,
		cclose:   make(chan struct{}),
	}

	return s
}

func (s *udpEntryPoint) ID() string {
	return s.opts.ID
}

func (s *udpEntryPoint) Type() string {
	return UDPEntryPoint
}

func (s *udpEntryPoint) Name() string {
	return s.opts.Name
}

func (s *udpEntryPoint) Endpoint() string {
	return fmt.Sprintf("%s.%s", s.endpoint, tunnel.EndpointAddr)
}

func (s *udpEntryPoint) Entrypoint() string {
	return s.opts.Endpoint
}

func (s *udpEntryPoint) Options() tunnel.Options {
	return s.opts
}

func (s *udpEntryPoint) Favorite(b bool) {
	s.favorite.Store(b)
}

func (s *udpEntryPoint) IsFavorite() bool {
	return s.favorite.Load()
}

func (s *udpEntryPoint) init() error {
	tcp := &config.ServiceConfig{
		Name: s.opts.Name,
		Addr: s.opts.Endpoint,
		Handler: &config.HandlerConfig{
			Type:  "udp",
			Chain: s.opts.Name,
		},
		Listener: &config.ListenerConfig{
			Type: "udp",
			Metadata: map[string]any{
				"keepalive": s.opts.Keepalive,
				"ttl":       time.Duration(s.opts.TTL) * time.Second,
			},
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

func (s *udpEntryPoint) Run() (err error) {
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

		stats := xstats.NewStats(false)

		cfg := s.config.Services[0]
		ln := udp.NewListener(
			listener.AddrOption(cfg.Addr),
			listener.LoggerOption(log.WithFields(map[string]any{"kind": "listener", "listener": "udp"})),
			listener.StatsOption(stats),
		)
		if err = ln.Init(mdx.NewMetadata(cfg.Listener.Metadata)); err != nil {
			return
		}

		handlerLogger := log.WithFields(map[string]any{"kind": "handler", "handler": "udp"})
		h := local.NewHandler(
			handler.RouterOption(xchain.NewRouter(
				chain.ChainRouterOption(ch),
				chain.LoggerRouterOption(handlerLogger),
			)),
			handler.LoggerOption(handlerLogger),
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

func (s *udpEntryPoint) Status() *xservice.Status {
	if ss, _ := s.forward.(tunnel.ServiceStatus); ss != nil {
		return ss.Status()
	}
	return nil
}

func (s *udpEntryPoint) Stats() cfg.ServiceStats {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.stats
}

func (s *udpEntryPoint) SetStats(stats cfg.ServiceStats) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stats = stats
}

func (s *udpEntryPoint) Close() error {
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

func (s *udpEntryPoint) IsClosed() bool {
	select {
	case <-s.cclose:
		return true
	default:
		return false
	}
}

func (s *udpEntryPoint) setErr(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.err = err
}

func (s *udpEntryPoint) Err() error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.err
}
