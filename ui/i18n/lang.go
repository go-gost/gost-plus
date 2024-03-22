package i18n

import "sync"

type Lang struct {
	Name    Key
	Value   string
	content map[Key]string
}

var langs = []Lang{
	{
		Name:    English,
		Value:   "en_US",
		content: en_US,
	},
	{
		Name:    Chinese,
		Value:   "zh_CN",
		content: zh_CN,
	},
}

func Langs() []Lang {
	return langs
}

var (
	currentLang Lang = langs[0]
	mux         sync.RWMutex
)

func Current() Lang {
	mux.RLock()
	defer mux.RUnlock()

	return currentLang
}

func Set(lang string) {
	mux.Lock()
	defer mux.Unlock()

	for i := range langs {
		if langs[i].Value == lang {
			currentLang = langs[i]
			return
		}
	}
	currentLang = langs[0]
}

const (
	Tunnel             Key = "tunnel"
	Entrypoint         Key = "entrypoint"
	Type               Key = "type"
	Name               Key = "name"
	Address            Key = "address"
	Endpoint           Key = "endpoint"
	BasicAuth          Key = "basicAuth"
	Username           Key = "username"
	Password           Key = "password"
	DirectoryPath      Key = "dirPath"
	CustomHostname     Key = "customHostname"
	Hostname           Key = "hostname"
	EnableTLS          Key = "enableTLS"
	TunnelID           Key = "tunnelID"
	Keepalive          Key = "keepalive"
	FileTunnelDesc     Key = "fileTunnelDesc"
	HTTPTunnelDesc     Key = "httpTunnelDesc"
	TCPTunnelDesc      Key = "tcpTunnelDesc"
	UDPTunnelDesc      Key = "udpTunnelDesc"
	TCPEntrypointDesc  Key = "tcpEntrypointDesc"
	UDPEntrypointDesc  Key = "udpEntrypointDesc"
	Settings           Key = "settings"
	Language           Key = "language"
	Theme              Key = "theme"
	Light              Key = "light"
	Dark               Key = "dark"
	OK                 Key = "ok"
	Cancel             Key = "cancel"
	DeleteTunnel       Key = "deleteTunnel"
	DeleteEntrypoint   Key = "deleteEntrypoint"
	ErrInvalidTunnelID Key = "errInvalidTunnelID"
	ErrInvalidAddr     Key = "errInvalidAddr"
	ErrDigitOnly       Key = "errDigitOnly"
	ErrDirectory       Key = "errDir"

	English Key = "english"
	Chinese Key = "chinese"
)

type Key string

func (k Key) Value() string {
	return Get(k)
}

func Get(key Key) string {
	mux.RLock()
	defer mux.RUnlock()

	if v := currentLang.content[key]; v != "" {
		return v
	}

	return langs[0].content[key]
}
