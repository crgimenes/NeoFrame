package config

type Config struct {
	GetScreenInfo    bool
	UnixDomainSocket string
	ServerMode       bool
}

var CFG = &Config{}
