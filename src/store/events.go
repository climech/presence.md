package store

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

func (as *ArticleStore) onCreate(event fsnotify.Event) {
	article, err := as.loadArticle(event.Name)
	if err != nil {
		log.Printf("couldn't load article '%s': %v\n", event.Name, err)
		return
	}

	// Rename if no timestamp; this will trigger a Rename event followed by
	// Create.
	if article.PubTime == nil {
		now := time.Now()
		article.PubTime = &now
		dir, _ := filepath.Split(event.Name)
		newName := filepath.Join(dir, makeFilename(article))
		if err := os.Rename(event.Name, newName); err != nil {
			log.Printf("couldn't rename file '%s': %s\n", event.Name, err)
			return
		}
		log.Printf("renamed file: '%s' -> '%s'\n", event.Name, newName)
	} else {
		as.insert(article)
		log.Printf("loaded entry: '%s'\n", article.Slug)
	}
}

func (as *ArticleStore) onRemove(event fsnotify.Event) {
	article := as.GetByFilename(event.Name)
	if article != nil {
		as.remove(article.Slug)
		log.Printf("removed entry: '%s'\n", article.Slug)
	}
}
