package tunnel

import (
	"errors"
	"sync"

	"github.com/go-gost/core/logger"
	"github.com/go-gost/gost-plus/config"
	xconfig "github.com/go-gost/x/config"
	_ "github.com/go-gost/x/connector/tunnel"
	_ "github.com/go-gost/x/dialer/ws"
)

const (
	endpointAddr = "gost.plus"
	serverName   = "tunnel.gost.plus"
	serverAddr   = serverName + ":443"
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

type Tunnel interface {
	ID() string
	Type() string
	Name() string
	Endpoint() string
	Entrypoint() string
	Options() Options
	Run() error
	Favorite(b bool)
	IsFavorite() bool
	Close() error
	IsClosed() bool
}

type tunnelList struct {
	tunnels []Tunnel
	mux     sync.RWMutex
}

func (sl *tunnelList) Count() int {
	sl.mux.RLock()
	defer sl.mux.RUnlock()
	return len(sl.tunnels)
}

func (sl *tunnelList) Add(s Tunnel) {
	sl.mux.Lock()
	defer sl.mux.Unlock()
	sl.tunnels = append(sl.tunnels, s)
}

func (sl *tunnelList) Set(s Tunnel) {
	if s == nil {
		return
	}

	sl.mux.Lock()
	defer sl.mux.Unlock()

	for i, sv := range sl.tunnels {
		if sv != nil && sv.ID() == s.ID() {
			sl.tunnels[i] = s
		}
	}
}

func (sl *tunnelList) Get(index int) Tunnel {
	sl.mux.RLock()
	defer sl.mux.RUnlock()
	if index < 0 || index >= len(sl.tunnels) {
		return nil
	}
	return sl.tunnels[index]
}

func (sl *tunnelList) GetID(id string) Tunnel {
	sl.mux.RLock()
	defer sl.mux.RUnlock()

	for _, s := range sl.tunnels {
		if s != nil && s.ID() == id {
			return s
		}
	}
	return nil
}

func (sl *tunnelList) DeleteID(id string) {
	sl.mux.Lock()
	defer sl.mux.Unlock()

	for i, s := range sl.tunnels {
		if s != nil && s.ID() == id {
			s.Close()
			sl.tunnels[i] = nil
			return
		}
	}
}

var (
	tunnels tunnelList
)

func TunnelCount() int {
	return tunnels.Count()
}

func AddTunnel(s Tunnel) {
	tunnels.Add(s)
}

func SetTunnel(s Tunnel) {
	t := tunnels.GetID(s.ID())
	if t == nil {
		return
	}
	s.Favorite(t.IsFavorite())
	tunnels.Set(s)
}

func GetTunnel(index int) Tunnel {
	return tunnels.Get(index)
}

func GetTunnelID(id string) Tunnel {
	return tunnels.GetID(id)
}

func DeleteTunnel(id string) {
	tunnels.DeleteID(id)
}

func chainConfig(id string, name string) *xconfig.ChainConfig {
	return &xconfig.ChainConfig{
		Name: name,
		Hops: []*xconfig.HopConfig{
			{
				Name: name,
				Nodes: []*xconfig.NodeConfig{
					{
						Name: name,
						Addr: serverAddr,
						Connector: &xconfig.ConnectorConfig{
							Type:     "tunnel",
							Metadata: map[string]any{"tunnel.id": id},
						},
						Dialer: &xconfig.DialerConfig{
							Type: "wss",
							TLS: &xconfig.TLSConfig{
								Secure:     true,
								ServerName: serverName,
							},
						},
					},
				},
			},
		},
	}
}

func LoadConfig() {
	for _, tun := range config.Global().Tunnels {
		if tun == nil {
			continue
		}

		s := createTunnel(tun.Type, Options{
			ID:        tun.ID,
			Name:      tun.Name,
			Endpoint:  tun.Endpoint,
			Hostname:  tun.Hostname,
			Username:  tun.Username,
			Password:  tun.Password,
			EnableTLS: tun.EnableTLS,
		})
		if s == nil {
			continue
		}

		if tun.Closed {
			s.Close()
		} else {
			s.Run()
		}

		s.Favorite(tun.Favorite)
		tunnels.Add(s)
	}
}

func SaveTunnel() error {
	cfg := config.Global()
	cfg.Tunnels = nil

	for i := 0; i < tunnels.Count(); i++ {
		tun := tunnels.Get(i)
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
