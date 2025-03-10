package i18n

var en_US = map[Key]string{
	Tunnel:             "Tunnel",
	Entrypoint:         "Entrypoint",
	Type:               "Type",
	Name:               "Name",
	Address:            "Address",
	Endpoint:           "Endpoint",
	BasicAuth:          "Basic auth",
	Username:           "Username",
	Password:           "Password",
	DirectoryPath:      "Directory path",
	CustomHostname:     "Custom hostname (rewrite HTTP Host header)",
	Hostname:           "Hostname",
	EnableTLS:          "Enalbe TLS",
	TunnelID:           "Tunnel ID",
	Keepalive:          "Keepalive",
	FileTunnelDesc:     "Expose local files to public network",
	HTTPTunnelDesc:     "Expose local HTTP service to public network",
	TCPTunnelDesc:      "Expose local TCP service to public network",
	UDPTunnelDesc:      "Expose local UDP service to public network",
	TCPEntrypointDesc:  "Create an entrypoint to the specified TCP tunnel",
	UDPEntrypointDesc:  "Create an entrypoint to the specified UDP tunnel",
	OK:                 "OK",
	Cancel:             "Cancel",
	DeleteTunnel:       "Delete tunnel?",
	DeleteEntrypoint:   "Delete entrypoint?",
	ErrInvalidTunnelID: "invalid tunnel ID, should be a valid UUID",
	ErrInvalidAddr:     "invalid address format, should be [IP]:PORT or [HOST]:PORT",
	ErrDigitOnly:       "Must contain only digits",
	ErrDirectory:       "is not a directory",

	English:     "English",
	Chinese:     "Chinese",
	Settings:    "Settings",
	Language:    "Language",
	Theme:       "Theme",
	ThemeLight:  "Light",
	ThemeDark:   "Dark",
	ThemeSystem: "System",

	Inspector: "Traffic Inspector",
}
