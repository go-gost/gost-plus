package entrypoint

import (
	"errors"
	"sync"

	"github.com/go-gost/core/logger"
	"github.com/go-gost/gost-plus/config"
	"github.com/go-gost/gost-plus/tunnel"
)

const (
	TCPEntryPoint = "tcp"
	UDPEntryPoint = "udp"
)

var (
	ErrEntryPointClosed = errors.New("entrypoint closed")
)

type EntryPoint = tunnel.Tunnel

type entryPointList struct {
	list []EntryPoint
	mux  sync.RWMutex
}

var (
	entryPoints entryPointList
)

func Count() int {
	entryPoints.mux.RLock()
	defer entryPoints.mux.RUnlock()
	return len(entryPoints.list)
}

func Add(s EntryPoint) {
	entryPoints.mux.Lock()
	defer entryPoints.mux.Unlock()
	entryPoints.list = append(entryPoints.list, s)
}

func Set(s EntryPoint) {
	if s == nil {
		return
	}

	old := Get(s.ID())
	if old == nil {
		return
	}
	s.Favorite(old.IsFavorite())

	entryPoints.mux.Lock()
	defer entryPoints.mux.Unlock()

	for i, ep := range entryPoints.list {
		if ep != nil && ep.ID() == s.ID() {
			entryPoints.list[i] = s
		}
	}
}

func GetIndex(index int) EntryPoint {
	entryPoints.mux.RLock()
	defer entryPoints.mux.RUnlock()
	if index < 0 || index >= len(entryPoints.list) {
		return nil
	}
	return entryPoints.list[index]
}

func Get(id string) EntryPoint {
	entryPoints.mux.RLock()
	defer entryPoints.mux.RUnlock()

	for _, s := range entryPoints.list {
		if s != nil && s.ID() == id {
			return s
		}
	}
	return nil
}

func Delete(id string) {
	entryPoints.mux.Lock()
	defer entryPoints.mux.Unlock()

	for i, s := range entryPoints.list {
		if s != nil && s.ID() == id {
			s.Close()
			entryPoints.list[i] = nil
			return
		}
	}
}

func LoadConfig() {
	for _, cfg := range config.Global().EntryPoints {
		if cfg == nil {
			continue
		}

		ep := createEntryPoint(cfg.Type, tunnel.Options{
			ID:        cfg.ID,
			Name:      cfg.Name,
			Endpoint:  cfg.Endpoint,
			Hostname:  cfg.Hostname,
			Username:  cfg.Username,
			Password:  cfg.Password,
			EnableTLS: cfg.EnableTLS,
			Keepalive: cfg.Keepalive,
			TTL:       cfg.TTL,
		})
		if ep == nil {
			continue
		}

		if cfg.Closed {
			ep.Close()
		} else {
			ep.Run()
		}

		ep.Favorite(cfg.Favorite)
		Add(ep)
	}
}

func SaveConfig() error {
	cfg := config.Global()
	cfg.EntryPoints = nil

	for i := 0; i < Count(); i++ {
		ep := GetIndex(i)
		if ep == nil {
			continue
		}

		opts := ep.Options()

		cfg.EntryPoints = append(cfg.EntryPoints, &config.Tunnel{
			ID:        ep.ID(),
			Name:      ep.Name(),
			Type:      ep.Type(),
			Endpoint:  ep.Entrypoint(),
			Hostname:  opts.Hostname,
			Username:  opts.Username,
			Password:  opts.Password,
			EnableTLS: opts.EnableTLS,
			Favorite:  ep.IsFavorite(),
			Closed:    ep.IsClosed(),
		})
	}

	config.Set(cfg)

	if err := cfg.Write(); err != nil {
		logger.Default().Error(err)
		return err
	}
	return nil
}

func createEntryPoint(st string, opts tunnel.Options) EntryPoint {
	options := []tunnel.Option{
		tunnel.IDOption(opts.ID),
		tunnel.NameOption(opts.Name),
		tunnel.EndpointOption(opts.Endpoint),
		tunnel.HostnameOption(opts.Hostname),
		tunnel.UsernameOption(opts.Username),
		tunnel.PasswordOption(opts.Password),
		tunnel.EnableTLSOption(opts.EnableTLS),
	}
	switch st {
	case TCPEntryPoint:
		return NewTCPEntryPoint(options...)
	case UDPEntryPoint:
		return NewUDPEntryPoint(options...)
	default:
		return nil
	}
}
