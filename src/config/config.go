package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

type SiteConfig struct {
	Title             string
	Author            string
	Description       string
	MaxEntriesPerPage uint
	DateFormat        string
}

type ServerConfig struct {
	Host         string
	Port         uint
	PortTLS      uint
	ForceTLS     bool
	TLSKey       string
	TLSCert      string
	StaticDir    string
	PostsDir     string
	PagesDir     string
	TemplatesDir string
	ErrorLog     string
	AccessLog    string
	ProxyCount   uint
}

type Config struct {
	*SiteConfig
	*ServerConfig
}

func expandPath(path, home, cwd string) string {
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(home, path[2:])
	} else if strings.HasPrefix(path, "./") {
		return filepath.Join(cwd, path[2:])
	}
	return path
}

func LoadConfig(dir string) (*Config, error) {
	viper.SetDefault("server.host", "127.0.0.1")
	viper.SetDefault("server.port", 9001)
	viper.SetDefault("server.port_tls", 0)
	viper.SetDefault("server.force_tls", false)
	viper.SetDefault("server.tls_key", "")
	viper.SetDefault("server.tls_cert", "")
	viper.SetDefault("server.static_dir", "")
	viper.SetDefault("server.posts_dir", "")
	viper.SetDefault("server.pages_dir", "")
	viper.SetDefault("server.templates_dir", "")
	viper.SetDefault("server.error_log", "")
	viper.SetDefault("server.access_log", "")
	viper.SetDefault("site.title", "My Blog")
	viper.SetDefault("site.author", "John Doe")
	viper.SetDefault("site.description", "John Doe's personal blog")
	viper.SetDefault("site.max_entries_per_page", 10)
	viper.SetDefault("site.date_format", "%F")

	viper.AddConfigPath(dir)
	viper.SetConfigName("config")
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	cwd := ""
	if fp := viper.ConfigFileUsed(); fp != "" {
		cwd = filepath.Dir(fp)
	}

	config := &Config{
		&SiteConfig{
			Title:             viper.GetString("site.title"),
			Author:            viper.GetString("site.author"),
			Description:       viper.GetString("site.description"),
			MaxEntriesPerPage: viper.GetUint("site.max_entries_per_page"),
			DateFormat:        viper.GetString("site.date_format"),
		},
		&ServerConfig{
			Host:         viper.GetString("server.host"),
			Port:         viper.GetUint("server.port"),
			PortTLS:      viper.GetUint("server.port_tls"),
			ForceTLS:     viper.GetBool("server.force_tls"),
			TLSKey:       expandPath(viper.GetString("server.tls_key"), home, cwd),
			TLSCert:      expandPath(viper.GetString("server.tls_cert"), home, cwd),
			StaticDir:    expandPath(viper.GetString("server.static_dir"), home, cwd),
			PostsDir:     expandPath(viper.GetString("server.posts_dir"), home, cwd),
			PagesDir:     expandPath(viper.GetString("server.pages_dir"), home, cwd),
			TemplatesDir: expandPath(viper.GetString("server.templates_dir"), home, cwd),
			AccessLog:    expandPath(viper.GetString("server.access_log"), home, cwd),
			ErrorLog:     expandPath(viper.GetString("server.error_log"), home, cwd),
			ProxyCount:   viper.GetUint("server.proxy_count"),
		},
	}

	return config, nil
}
