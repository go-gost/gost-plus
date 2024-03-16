package tunnel

import (
	"errors"
	"sync"

	"github.com/go-gost/core/logger"
	"github.com/go-gost/gost.plus/config"
	xconfig "github.com/go-gost/x/config"
	_ "github.com/go-gost/x/connector/tunnel"
	_ "github.com/go-gost/x/dialer/ws"
	xservice "github.com/go-gost/x/service"
)

const (
	EndpointAddr = "gost.plus"
	ServerName   = "tunnel.gost.plus"
	ServerAddr   = ServerName + ":443"
)

const (
	FileTunnel = "file"
	HTTPTunnel = "http"
	TCPTunnel  = "tcp"
	UDPTunnel  = "udp"
)

var (
	ErrTunnelClosed = errors.New("tunnel closed")
)

type Options struct {
	ID        string
	Name      string
	Endpoint  string
	Hostname  string
	Username  string
	Password  string
	EnableTLS bool
	Keepalive bool
	TTL       int
}

type Option func(opts *Options)

func IDOption(id string) Option {
	return func(opts *Options) {
		opts.ID = id
	}
}

func NameOption(name string) Option {
	return func(opts *Options) {
		opts.Name = name
	}
}

func EndpointOption(endpoint string) Option {
	return func(opts *Options) {
		opts.Endpoint = endpoint
	}
}

func HostnameOption(hostname string) Option {
	return func(opts *Options) {
		opts.Hostname = hostname
	}
}

func UsernameOption(username string) Option {
	return func(opts *Options) {
		opts.Username = username
	}
}

func PasswordOption(password string) Option {
	return func(opts *Options) {
		opts.Password = password
	}
}

func EnableTLSOption(b bool) Option {
	return func(opts *Options) {
		opts.EnableTLS = b
	}
}

func KeepaliveOption(b bool) Option {
	return func(opts *Options) {
		opts.Keepalive = b
	}
}

func TTLOption(ttl int) Option {
	return func(opts *Options) {
		opts.TTL = ttl
	}
}

type ServiceStatus interface {
	Status() *xservice.Status
}

type Tunnel interface {
	ID() string
	Type() string
	Name() string
	Endpoint() string
	Entrypoint() string
	Options() Options
	Run() error
	Status() *xservice.Status
	Favorite(b bool)
	IsFavorite() bool
	Close() error
	IsClosed() bool
	Err() error
}

type tunnelList struct {
	list []Tunnel
	mux  sync.RWMutex
}

var (
	tunnels tunnelList
)

func Count() int {
	tunnels.mux.RLock()
	defer tunnels.mux.RUnlock()
	return len(tunnels.list)
}

func Add(s Tunnel) {
	tunnels.mux.Lock()
	defer tunnels.mux.Unlock()
	tunnels.list = append(tunnels.list, s)
}

func Set(s Tunnel) {
	if s == nil {
		return
	}
	t := Get(s.ID())
	if t == nil {
		return
	}
	s.Favorite(t.IsFavorite())

	tunnels.mux.Lock()
	defer tunnels.mux.Unlock()

	for i, sv := range tunnels.list {
		if sv != nil && sv.ID() == s.ID() {
			tunnels.list[i] = s
		}
	}
}

func GetIndex(index int) Tunnel {
	tunnels.mux.RLock()
	defer tunnels.mux.RUnlock()
	if index < 0 || index >= len(tunnels.list) {
		return nil
	}
	return tunnels.list[index]
}

func Get(id string) Tunnel {
	tunnels.mux.RLock()
	defer tunnels.mux.RUnlock()

	for _, s := range tunnels.list {
		if s != nil && s.ID() == id {
			return s
		}
	}
	return nil
}

func Delete(id string) {
	tunnels.mux.Lock()
	defer tunnels.mux.Unlock()

	for i, s := range tunnels.list {
		if s != nil && s.ID() == id {
			s.Close()
			tunnels.list[i] = nil
			return
		}
	}
}

func ChainConfig(id string, name string) *xconfig.ChainConfig {
	return &xconfig.ChainConfig{
		Name: name,
		Hops: []*xconfig.HopConfig{
			{
				Name: name,
				Nodes: []*xconfig.NodeConfig{
					{
						Name: name,
						Addr: ServerAddr,
						Connector: &xconfig.ConnectorConfig{
							Type:     "tunnel",
							Metadata: map[string]any{"tunnel.id": id},
						},
						Dialer: &xconfig.DialerConfig{
							Type: "wss",
							TLS: &xconfig.TLSConfig{
								Secure:     true,
								ServerName: ServerName,
							},
						},
					},
				},
			},
		},
	}
}

func LoadConfig() {
	for _, cfg := range config.Get().Tunnels {
		if cfg == nil {
			continue
		}

		tun := createTunnel(cfg.Type, Options{
			ID:        cfg.ID,
			Name:      cfg.Name,
			Endpoint:  cfg.Endpoint,
			Hostname:  cfg.Hostname,
			Username:  cfg.Username,
			Password:  cfg.Password,
			EnableTLS: cfg.EnableTLS,
		})
		if tun == nil {
			continue
		}

		if cfg.Closed {
			tun.Close()
		} else {
			tun.Run()
		}

		tun.Favorite(cfg.Favorite)
		Add(tun)
	}
}

func SaveConfig() error {
	cfg := config.Get()
	cfg.Tunnels = nil

	for i := 0; i < Count(); i++ {
		tun := GetIndex(i)
		if tun == nil {
			continue
		}

		opts := tun.Options()

		cfg.Tunnels = append(cfg.Tunnels, &config.Tunnel{
			ID:        tun.ID(),
			Name:      tun.Name(),
			Type:      tun.Type(),
			Endpoint:  tun.Endpoint(),
			Hostname:  opts.Hostname,
			Username:  opts.Username,
			Password:  opts.Password,
			EnableTLS: opts.EnableTLS,
			Favorite:  tun.IsFavorite(),
			Closed:    tun.IsClosed(),
		})
	}

	config.Set(cfg)

	if err := cfg.Write(); err != nil {
		logger.Default().Error(err)
		return err
	}
	return nil
}

func createTunnel(st string, opts Options) Tunnel {
	options := []Option{
		IDOption(opts.ID),
		NameOption(opts.Name),
		EndpointOption(opts.Endpoint),
		HostnameOption(opts.Hostname),
		UsernameOption(opts.Username),
		PasswordOption(opts.Password),
		EnableTLSOption(opts.EnableTLS),
	}
	switch st {
	case FileTunnel:
		return NewFileTunnel(options...)
	case HTTPTunnel:
		return NewHTTPTunnel(options...)
	case TCPTunnel:
		return NewTCPTunnel(options...)
	case UDPTunnel:
		return NewUDPTunnel(options...)
	default:
		return nil
	}
}
