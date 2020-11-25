package store

import (
	"log"
	"os"
	"path/filepath"
	"presence/model"
	"sort"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/yuin/goldmark"
)

// ArticleStore contains a collection of articles generated from Markdown files
// present in a directory. Changes to the files are immediately reflected in
// the store.
type ArticleStore struct {
	items    map[string]*model.Article
	mux      sync.Mutex
	watcher  *fsnotify.Watcher
	markdown goldmark.Markdown
}

func NewArticleStore(dir string) (*ArticleStore, error) {
	as := &ArticleStore{
		items: make(map[string]*model.Article),
	}
	if err := as.initWatcher(dir); err != nil {
		return nil, err
	}
	return as, nil
}

func (as *ArticleStore) Close() {
	as.watcher.Close()
}

// watch initializes and starts the file watcher.
func (as *ArticleStore) initWatcher(dir string) error {
	var err error
	as.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	if err := as.watcher.Add(dir); err != nil {
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

func (as *ArticleStore) insert(article *model.Article) {
	as.mux.Lock()
	as.items[article.Slug] = article
	as.mux.Unlock()
}

func (as *ArticleStore) remove(slug string) {
	as.mux.Lock()
	delete(as.items, slug)
	as.mux.Unlock()
}
func (as *ArticleStore) Len() int {
	as.mux.Lock()
	defer as.mux.Unlock()
	return len(as.items)
}

// Get returns the *model.Article from the store given its slug, or nil, if it
// doesn't exist.
func (as *ArticleStore) Get(slug string) *model.Article {
	as.mux.Lock()
	defer as.mux.Unlock()
	if article, ok := as.items[slug]; ok {
		return article
	}
	return nil
}

func (as *ArticleStore) GetByFilename(filename string) *model.Article {
	as.mux.Lock()
	defer as.mux.Unlock()
	for _, v := range as.items {
		if v.Filename == filename {
			return v
		}
	}
	return nil
}

// GetRecent gets the most recent articles in the store. Returns empty slice if
// offset and limit go out of range.
func (as *ArticleStore) GetRecent(offset, limit int) []*model.Article {
	all := as.GetAll()
	length := len(all)
	end := offset + limit
	if offset > length {
		offset = length
	}
	if end > length {
		end = length
	}
	return all[offset:end]
}

// All returns all the articles in the store, sorted by pubtime (most recent
// first).
func (as *ArticleStore) GetAll() []*model.Article {
	as.mux.Lock()
	defer as.mux.Unlock()

	values := make([]*model.Article, 0, len(as.items))
	for _, v := range as.items {
		values = append(values, v)
	}
	sort.SliceStable(values, func(i, j int) bool {
		return values[i].Title < values[j].Title
	})
	sort.SliceStable(values, func(i, j int) bool {
		if values[i].PubTime == nil || values[j].PubTime == nil {
			return false
		}
		return values[i].PubTime.Unix() > values[j].PubTime.Unix()
	})

	return values
}
