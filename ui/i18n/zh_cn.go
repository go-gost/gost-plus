package i18n

var zh_CN = map[Key]string{
	Tunnel:             "隧道",
	Entrypoint:         "入口点",
	Type:               "类型",
	Name:               "名称",
	Address:            "地址",
	Endpoint:           "端点",
	BasicAuth:          "基本认证",
	Username:           "用户名",
	Password:           "密码",
	DirectoryPath:      "文件目录路径",
	CustomHostname:     "自定义主机名（重写HTTP Host头）",
	Hostname:           "主机名",
	EnableTLS:          "开启TLS",
	TunnelID:           "隧道ID",
	Keepalive:          "保持连接",
	FileTunnelDesc:     "将本地文件系统暴露到公网",
	HTTPTunnelDesc:     "将本地的一个HTTP服务暴露到公网",
	TCPTunnelDesc:      "将本地的一个TCP服务暴露到公网",
	UDPTunnelDesc:      "将本地的一个UDP服务暴露到公网",
	TCPEntrypointDesc:  "创建一个指定TCP隧道的入口点",
	UDPEntrypointDesc:  "创建一个指定UDP隧道的入口点",
	OK:                 "确认",
	Cancel:             "取消",
	DeleteTunnel:       "删除隧道？",
	DeleteEntrypoint:   "删除入口点？",
	ErrInvalidTunnelID: "无效的隧道ID， 仅支持合法的UUID格式，例如：6bcb409c-dd0f-4ce7-9869-651c52c09d1c",
	ErrInvalidAddr:     "无效的地址格式，仅支持[IP]:PORT或[HOST]:PORT",
	ErrDigitOnly:       "仅能输入数字",
	ErrDirectory:       "不是一个目录",

	Settings:    "设置",
	Language:    "语言",
	Theme:       "主题",
	English:     "英语",
	Chinese:     "中文",
	ThemeLight:  "浅色",
	ThemeDark:   "深色",
	ThemeSystem: "系统",

	Inspector: "流量观察",
}
