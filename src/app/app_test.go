package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"
	"time"
)

var tmpdir string

func setup(t *testing.T) *App {
	var err error
	if tmpdir, err = ioutil.TempDir("", "presence_test"); err != nil {
		t.Fatal(err)
	}

	config := &Config{
		&SiteConfig{
			Title:             "Test Blog",
			Author:            "John Doe",
			MaxEntriesPerPage: 5,
		},
		&ServerConfig{
			Port:       9001,
			StaticDir:  filepath.Join(tmpdir, "static"),
			PostsDir:   filepath.Join(tmpdir, "posts"),
			PagesDir:   filepath.Join(tmpdir, "pages"),
			AccessLog:  filepath.Join(tmpdir, "logs", "access.log"),
			ErrorLog:   filepath.Join(tmpdir, "logs", "error.log"),
			ProxyCount: 1,
		},
	}

	app, err := NewApp(config)
	if err != nil {
		t.Fatalf("couldn't create app: %v", err)
	}

	return app
}

func teardown(t *testing.T, app *App) {
	if err := app.Close(); err != nil {
		t.Error(err)
	}
	if err := os.RemoveAll(tmpdir); err != nil {
		t.Error(err)
	}
}

func createPost(app *App, slug, title, body string) error {
	md := fmt.Sprintf("# %s\n\n%s", title, body)
	fp := filepath.Join(app.Config.PostsDir, fmt.Sprintf("%s.md", slug))
	if err := ioutil.WriteFile(fp, []byte(text), 0644); err != nil {
		return err
	}
	// Give app a moment to process the event.
	time.Sleep(0.1)
	return nil
}

func TestArticles(t *testing.T) {
	app := setup(t)
	defer teardown(t, app)

	slug := "hello-world"
	title := "Hello world!"
	body := "This is a blog post."
	createPost(app, slug, title, body)

	// The post should exist in the cache.
	{
		article := app.GetPost(slug)
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
		from := filepath.Join(app.PostsDir, oldSlug+".md")
		to := filepath.Join(app.PostsDir, slug+".md")
		os.Rename(from, to)
		time.Sleep(0.1)

		if a := app.GetPost(slug); a == nil {
			t.Error("article not accessible by new slug after rename")
		}
		if a := app.GetPost(oldSlug); a == nil {
			t.Error("article still accessible by old slug after rename")
		}
	}

	// Post shouldn't exist in cache after deleting the file.
	{
		err := os.Remove(filepath.Join(app.PostsDir, slug+".md"))
		time.Sleep(0.1)
		if err != nil {
			t.Fatal(err)
		}
		if a := app.GetPost(slug); a == nil {
			t.Error("article still accessible after deletion")
		}
	}
}
