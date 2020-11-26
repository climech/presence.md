package config

import (
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
	Port       uint
	PortTLS    uint
	ForceTLS   bool
	TLSKey     string
	TLSCert    string
	StaticDir  string
	PostsDir   string
	PagesDir   string
	ErrorLog   string
	AccessLog  string
	ProxyCount uint
}

type Config struct {
	*SiteConfig
	*ServerConfig
}

func LoadConfig(dir string) (*Config, error) {
	viper.SetDefault("server.port", 9001)
	viper.SetDefault("server.port_tls", 0)
	viper.SetDefault("server.force_tls", false)
	viper.SetDefault("server.tls_key", "")
	viper.SetDefault("server.tls_cert", "")
	viper.SetDefault("server.static_dir", "")
	viper.SetDefault("server.posts_dir", "")
	viper.SetDefault("server.pages_dir", "")
	viper.SetDefault("server.error_log", "")
	viper.SetDefault("server.access_log", "")
	viper.SetDefault("site.title", "My Blog")
	viper.SetDefault("site.author", "John Doe")
	viper.SetDefault("site.description", "John Doe's personal blog")
	viper.SetDefault("site.max_entries_per_page", 10)

	viper.AddConfigPath(dir)
	viper.AddConfigPath("./config")
	viper.SetConfigName("config")
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	config := &Config{
		&SiteConfig{
			Title:             viper.GetString("site.title"),
			Author:            viper.GetString("site.author"),
			Description:       viper.GetString("site.description"),
			MaxEntriesPerPage: viper.GetUint("site.max_entries_per_page"),
		},
		&ServerConfig{
			Port:       viper.GetUint("server.port"),
			PortTLS:    viper.GetUint("server.port_tls"),
			ForceTLS:   viper.GetBool("server.force_tls"),
			TLSKey:     viper.GetString("server.tls_key"),
			TLSCert:    viper.GetString("server.tls_cert"),
			StaticDir:  viper.GetString("server.static_dir"),
			PostsDir:   viper.GetString("server.posts_dir"),
			PagesDir:   viper.GetString("server.pages_dir"),
			AccessLog:  viper.GetString("server.access_log"),
			ErrorLog:   viper.GetString("server.error_log"),
			ProxyCount: viper.GetUint("server.proxy_count"),
		},
	}

	return config, nil
}
