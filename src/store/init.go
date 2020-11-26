package store

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/alecthomas/chroma/formatters/html"
	"github.com/fsnotify/fsnotify"
	"github.com/yuin/goldmark"
	mdhl "github.com/yuin/goldmark-highlighting"
	mdext "github.com/yuin/goldmark/extension"
	mdparser "github.com/yuin/goldmark/parser"
	mdhtml "github.com/yuin/goldmark/renderer/html"
)

// watch initializes and starts the file watcher.
func (as *ArticleStore) initWatcher() error {
	var err error
	as.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	if err := as.watcher.Add(as.Dir); err != nil {
		return err
	}

	go func() {
		for {
			select {
			case event, ok := <-as.watcher.Events:
				if !ok {
					return
				}
				if !isValidFilename(event.Name) {
					break
				}
				log.Println("event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write ||
					event.Op&fsnotify.Create == fsnotify.Create ||
					event.Op&fsnotify.Chmod == fsnotify.Chmod {
					as.onCreate(event)
				} else if event.Op&fsnotify.Remove == fsnotify.Remove ||
					event.Op&fsnotify.Rename == fsnotify.Rename {
					as.onRemove(event)
				}
			case err, ok := <-as.watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	return nil
}

func (as *ArticleStore) initArticles() {
	// Create dir and its parents if needed.
	if err := os.MkdirAll(as.Dir, 0755); err != nil {
		log.Fatalln(err)
	}

	files, err := ioutil.ReadDir(as.Dir)
	if err != nil {
		log.Printf("couldn't list directory: %s\n", err)
	}

	for _, f := range files {
		filename := filepath.Join(as.Dir, f.Name())
		if !isValidFilename(filename) {
			continue
		}
		// Mock fsnotify.Create events to load the articles.
		as.onCreate(fsnotify.Event{Name: filename, Op: fsnotify.Create})
	}
}

func (as *ArticleStore) initMarkdown() {
	as.markdown = goldmark.New(
		goldmark.WithExtensions(
			mdext.Linkify,
			mdext.Strikethrough,
			mdext.Table,
			mdext.Footnote,
			mdext.Typographer,
			mdhl.NewHighlighting(
				mdhl.WithFormatOptions(
					html.WithClasses(true),
				),
			),
		),
		goldmark.WithParserOptions(
			mdparser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			mdhtml.WithUnsafe(),
			//mdhtml.WithHardWraps(),
		),
	)
}
