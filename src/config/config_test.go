package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var tmpdir string

func setup(t *testing.T) {
	var err error
	if tmpdir, err = ioutil.TempDir("", "presence_test_config"); err != nil {
		t.Fatal(err)
	}
}

func teardown(t *testing.T) {
	if err := os.RemoveAll(tmpdir); err != nil {
		t.Fatal(err)
	}
}

const yamlFmtString = `
site:
    title:                %s
    description:          %s
    author:               %s
    max_entries_per_page: %d
    date_format:          %s
server:
    port:        %d
    port_tls:    %d
    force_tls:   %v
    tls_key:     %s
    tls_cert:    %s
    static_dir:  %s
    posts_dir:   %s
    pages_dir:   %s
    access_log:  %s
    error_log:   %s
    proxy_count: %d
`

func yamlFromConfig(c *Config) string {
	return fmt.Sprintf(
		strings.TrimSpace(yamlFmtString),
		c.SiteConfig.Title,
		c.SiteConfig.Description,
		c.SiteConfig.Author,
		c.SiteConfig.MaxEntriesPerPage,
		c.SiteConfig.DateFormat,
		c.ServerConfig.Port,
		c.ServerConfig.PortTLS,
		c.ServerConfig.ForceTLS,
		c.ServerConfig.TLSKey,
		c.ServerConfig.TLSCert,
		c.ServerConfig.StaticDir,
		c.ServerConfig.PostsDir,
		c.ServerConfig.PagesDir,
		c.ServerConfig.AccessLog,
		c.ServerConfig.ErrorLog,
		c.ServerConfig.ProxyCount,
	)
}

func TestLoadConfig(t *testing.T) {
	setup(t)
	defer teardown(t)

	want := &Config{
		&SiteConfig{
			Title:             "My title",
			Description:       "This is my blog",
			Author:            "Johnny",
			MaxEntriesPerPage: 5,
		},
		&ServerConfig{
			Port:       80,
			PortTLS:    443,
			ForceTLS:   true,
			TLSKey:     filepath.Join("path", "to", "key.pem"),
			TLSCert:    filepath.Join("path", "to", "cert.pem"),
			StaticDir:  filepath.Join("path", "to", "static"),
			PostsDir:   filepath.Join("path", "to", "posts"),
			PagesDir:   filepath.Join("path", "to", "pages"),
			AccessLog:  filepath.Join("path", "to", "access.log"),
			ErrorLog:   filepath.Join("path", "to", "error.log"),
			ProxyCount: 1,
		},
	}

	yaml := yamlFromConfig(want)

	fp := filepath.Join(tmpdir, "config.yaml")
	if err := ioutil.WriteFile(fp, []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := loadConfig(tmpdir)
	if err != nil {
		t.Fatalf("couldn't load test config: %s", err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("config mismatch (-want +got):\n\n%s\n", diff)
	}
}
