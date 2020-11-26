package store

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"
)

// wait is used to give the store a moment to process filesystem events.
func wait() {
	time.Sleep(50 * time.Millisecond)
}

func fileExists(fp string) bool {
	if _, err := os.Stat(fp); os.IsNotExist(err) {
		return false
	}
	return true
}

func setup(t *testing.T) *ArticleStore {
	tmpdir, err := ioutil.TempDir("", "presence_test")
	if err != nil {
		t.Fatal(err)
	}
	as, err := NewArticleStore(tmpdir)
	if err != nil {
		t.Fatalf("couldn't create store: %v", err)
	}
	wait()
	return as
}

func teardown(t *testing.T, as *ArticleStore) {
	as.Close()
	if err := os.RemoveAll(as.Dir); err != nil {
		t.Error(err)
	}
}

func TestCreateRenameDelete(t *testing.T) {
	as := setup(t)
	defer teardown(t, as)

	slug := "hello-world"
	timestamp := time.Now().Unix()
	fname := fmt.Sprintf("%s.%d.md", slug, timestamp)
	fpath := filepath.Join(as.Dir, fname)
	title := "Hello world!"
	body := "This is a blog post."
	text := fmt.Sprintf("# %s\n\n%s", title, body)

	if err := ioutil.WriteFile(fpath, []byte(text), 0644); err != nil {
		t.Fatal(err)
	}
	wait()

	// The post should exist in the store.
	{
		article := as.Get(slug)
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

	// Renaming the file should update the store.
	{
		_slug := slug
		_fpath := fpath

		slug = "test"
		fname = fmt.Sprintf("%s.%d.md", slug, timestamp)
		fpath = filepath.Join(as.Dir, fname)

		if err := os.Rename(_fpath, fpath); err != nil {
			t.Fatal(err)
		}
		wait()

		if as.Get(slug) == nil {
			t.Error("article not accessible by new slug after rename")
		}
		if as.Get(_slug) != nil {
			t.Error("article still accessible by old slug after rename")
		}
	}

	// Post shouldn't exist in the store after deletion.
	{
		err := os.Remove(fpath)
		if err != nil {
			t.Fatal(err)
		}
		wait()
		if as.Get(slug) != nil {
			t.Error("article still accessible after deletion")
		}
	}
}

// CreateWithoutTimestamp checks if the store correctly appends a timestamp to
// a filename on creation if no timestamp is present.
func TestCreateWithoutTimestamp(t *testing.T) {
	as := setup(t)
	defer teardown(t, as)

	slug := "hello-world"
	fname := fmt.Sprintf("%s.md", slug)
	fpath := filepath.Join(as.Dir, fname)
	title := "Hello world!"
	body := "This is a blog post."
	text := fmt.Sprintf("# %s\n\n%s", title, body)

	if err := ioutil.WriteFile(fpath, []byte(text), 0644); err != nil {
		t.Fatal(err)
	}
	wait()

	files, err := ioutil.ReadDir(as.Dir)
	if err != nil {
		t.Fatal(err)
	}

	if len(files) != 1 {
		msg := "want 1 file in dir, got %d"
		if len(files) > 0 {
			msg += ":\n"
			for i, f := range files {
				msg += fmt.Sprintf("%d: %s", i, f.Name())
			}
		}
		t.Error(msg)
	}

	if fileExists(fpath) {
		t.Errorf("file should not exist: %s", fpath)
	}

	re := regexp.MustCompile(`^[a-zA-Z0-9\-_]+\.-?\d+\.md$`) // some-text.1234.md
	name := files[0].Name()
	if !re.MatchString(name) {
		t.Errorf("unexpected filename: %s", name)
	}

	article := as.Get(slug)
	if article == nil {
		t.Errorf(`article (slug: %s) is nil, want non-nil`, slug)
	} else if article.PubTime == nil {
		t.Errorf(`article (slug: %s) pubtime is nil, want non-nil`, slug)
	}
}
