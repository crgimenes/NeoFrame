package config

type Config struct {
	GetScreenInfo    bool
	UnixDomainSocket string
	ServerMode       bool
	Silent           bool
}

var CFG = &Config{}
