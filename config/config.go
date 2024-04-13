package config

type Config struct {
	GetScreenInfo    bool
	UnixDomainSocket string
	ServerMode       bool
	WindowTitle      string
	WindowWidth      int
	WindowHeight     int
	WindowX          int
	WindowY          int
	WindowBorder     bool
	WindowBgColor    string
	RunLuaScript     string
	MousePassthrough bool
}

var CFG = &Config{}
