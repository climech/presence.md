package store

import (
	"presence/model"
	"sort"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/yuin/goldmark"
)

// ArticleStore contains a collection of articles generated from Markdown files
// present in a directory. Changes to the files are immediately reflected in
// the store.
type ArticleStore struct {
	Dir      string
	items    map[string]*model.Article
	watcher  *fsnotify.Watcher
	markdown goldmark.Markdown
	mux      sync.Mutex
}

func NewArticleStore(dirpath string) (*ArticleStore, error) {
	as := &ArticleStore{
		Dir:   dirpath,
		items: make(map[string]*model.Article),
	}
	if err := as.initWatcher(); err != nil {
		return nil, err
	}
	as.initMarkdown()
	return as, nil
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

func (as *ArticleStore) Close() {
	as.watcher.Close()
}
