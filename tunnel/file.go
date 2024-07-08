package tunnel

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-gost/core/auth"
	"github.com/go-gost/core/chain"
	"github.com/go-gost/core/handler"
	"github.com/go-gost/core/listener"
	"github.com/go-gost/core/logger"
	"github.com/go-gost/core/observer/stats"
	"github.com/go-gost/core/service"
	cfg "github.com/go-gost/gost.plus/config"
	xauth "github.com/go-gost/x/auth"
	xchain "github.com/go-gost/x/chain"
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
	stats    cfg.ServiceStats

	cclose chan struct{}

	err error
	mu  sync.RWMutex
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
		options.Name = endpoint
	}
	if options.CreatedAt.IsZero() {
		options.CreatedAt = time.Now()
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

func (s *fileTunnel) Entrypoint() string {
	return fmt.Sprintf("https://%s.%s", s.endpoint, EndpointAddr)
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
		Chains:   []*config.ChainConfig{ChainConfig(s.opts.ID, s.opts.Name)},
	}

	return nil
}

func (s *fileTunnel) Run() (err error) {
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
		cfg := s.config.Services[0]
		ln := tcp.NewListener(
			listener.LoggerOption(log.WithFields(map[string]any{"kind": "listener", "listener": "tcp"})),
		)
		if err = ln.Init(nil); err != nil {
			return
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
		if err = h.Init(mdx.NewMetadata(cfg.Handler.Metadata)); err != nil {
			return
		}
		s.file = xservice.NewService(s.opts.Name, ln, h, xservice.LoggerOption(log))
	}

	{
		var ch chain.Chainer
		ch, err = chain_parser.ParseChain(s.config.Chains[0], log)
		if err != nil {
			log.Error(err)
			return
		}

		pStats := &stats.Stats{}
		{
			pStats.Add(stats.KindCurrentConns, int64(s.stats.CurrentConns))
			pStats.Add(stats.KindInputBytes, int64(s.stats.InputBytes))
			pStats.Add(stats.KindOutputBytes, int64(s.stats.OutputBytes))
			pStats.Add(stats.KindTotalConns, int64(s.stats.TotalConns))
			pStats.Add(stats.KindTotalErrs, int64(s.stats.TotalErrs))
		}

		listenerLogger := log.WithFields(map[string]any{"kind": "listener", "listener": "rtcp"})
		cfg := s.config.Services[1]
		ln := rtcp.NewListener(
			listener.AddrOption(cfg.Addr),
			listener.RouterOption(xchain.NewRouter(chain.ChainRouterOption(ch), chain.LoggerRouterOption(listenerLogger))),
			listener.LoggerOption(listenerLogger),
			listener.StatsOption(pStats),
		)
		if err = ln.Init(mdx.NewMetadata(cfg.Listener.Metadata)); err != nil {
			return
		}

		handlerLogger := log.WithFields(map[string]any{"kind": "handler", "handler": "rtcp"})
		h := remote.NewHandler(
			handler.RouterOption(xchain.NewRouter(chain.LoggerRouterOption(handlerLogger))),
			handler.LoggerOption(handlerLogger),
		)
		if err = h.Init(mdx.NewMetadata(cfg.Handler.Metadata)); err != nil {
			return
		}
		if forwarder, ok := h.(handler.Forwarder); ok {
			forwarder.Forward(hop.NewHop(
				hop.NodeOption(chain.NewNode(s.opts.Name, s.file.Addr().String())),
				hop.LoggerOption(log.WithFields(map[string]any{"kind": "hop"})),
			))
		}
		s.forward = xservice.NewService(s.opts.Name, ln, h,
			xservice.LoggerOption(log),
			xservice.StatsOption(pStats),
		)
		log.Infof("service listen on %s", s.file.Addr())
	}

	go s.file.Serve()
	go func() {
		s.setErr(s.forward.Serve())
	}()

	log.Infof("file service run at %s", s.file.Addr())
	return nil
}

func (s *fileTunnel) Status() *xservice.Status {
	if ss, _ := s.forward.(ServiceStatus); ss != nil {
		return ss.Status()
	}
	return nil
}

func (s *fileTunnel) Stats() cfg.ServiceStats {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.stats
}

func (s *fileTunnel) SetStats(stats cfg.ServiceStats) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stats = stats
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

func (s *fileTunnel) setErr(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.err = err
}

func (s *fileTunnel) Err() error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.err
}
