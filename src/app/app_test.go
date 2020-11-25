package app

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"presence/config"
)

var tmpdir string

func setup(t *testing.T) *App {
	var err error
	if tmpdir, err = ioutil.TempDir("", "presence_test"); err != nil {
		t.Fatal(err)
	}

	conf := &config.Config{
		&config.SiteConfig{
			Title:             "Test Blog",
			Author:            "John Doe",
			MaxEntriesPerPage: 5,
		},
		&config.ServerConfig{
			Port:       9001,
			StaticDir:  filepath.Join(tmpdir, "static"),
			PostsDir:   filepath.Join(tmpdir, "posts"),
			PagesDir:   filepath.Join(tmpdir, "pages"),
			AccessLog:  filepath.Join(tmpdir, "logs", "access.log"),
			ErrorLog:   filepath.Join(tmpdir, "logs", "error.log"),
			ProxyCount: 1,
		},
	}

	a, err := NewApp(conf)
	if err != nil {
		t.Fatalf("couldn't create App: %v", err)
	}

	return a
}

func teardown(t *testing.T, a *App) {
	if err := a.Close(); err != nil {
		t.Error(err)
	}
	if err := os.RemoveAll(tmpdir); err != nil {
		t.Error(err)
	}
}

func createPost(a *App, slug, title, body string) error {
	md := fmt.Sprintf("# %s\n\n%s", title, body)
	fp := filepath.Join(a.Config.PostsDir, fmt.Sprintf("%s.md", slug))
	if err := ioutil.WriteFile(fp, []byte(text), 0644); err != nil {
		return err
	}
	// Give app a moment to process the event.
	time.Sleep(0.1)
	return nil
}

func TestArticles(t *testing.T) {
	a := setup(t)
	defer teardown(t, a)

	slug := "hello-world"
	title := "Hello world!"
	body := "This is a blog post."
	createPost(a, slug, title, body)

	// The post should exist in the cache.
	{
		article := a.GetPost(slug)
		if article == nil {
			t.Fatal("article is nil, want non-nil")
		}
		if article.Title != title {
			t.Errorf(
				`article title mismatch; want "%s", got "%s"`,
				title,
				article.Title,
			)
		}
	}

	// Renaming the file should update the cache.
	oldSlug := slug
	slug = "test"
	{
		from := filepath.Join(a.Config.PostsDir, oldSlug+".md")
		to := filepath.Join(a.Config.PostsDir, slug+".md")
		os.Rename(from, to)
		time.Sleep(0.1)

		if a.GetPost(slug) == nil {
			t.Error("article not accessible by new slug after rename")
		}
		if a.GetPost(oldSlug) == nil {
			t.Error("article still accessible by old slug after rename")
		}
	}

	// Post shouldn't exist in cache after deleting the file.
	{
		err := os.Remove(filepath.Join(a.Config.PostsDir, slug+".md"))
		time.Sleep(0.1)
		if err != nil {
			t.Fatal(err)
		}
		if a.GetPost(slug) == nil {
			t.Error("article still accessible after deletion")
		}
	}
}
