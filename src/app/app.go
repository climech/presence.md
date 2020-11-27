package app

import (
	"fmt"
	"presence/config"
	"presence/model"
	"presence/store"
)

const AppName = "presence"

var Version = "" // injected on build

type App struct {
	Config *config.Config
	posts  *store.ArticleStore
	pages  *store.ArticleStore
}

func New(config *config.Config) (*App, error) {
	if config.PostsDir == "" || config.PagesDir == "" {
		return nil, fmt.Errorf("posts_dir and pages_dir must be set")
	}
	posts, err := store.NewArticleStore(config.PostsDir)
	if err != nil {
		return nil, fmt.Errorf("couldn't init posts: %s", err)
	}
	pages, err := store.NewArticleStore(config.PagesDir)
	if err != nil {
		return nil, fmt.Errorf("couldn't init pages: %s", err)
	}

	app := &App{
		Config: config,
		posts:  posts,
		pages:  pages,
	}

	return app, nil
}

func (a *App) GetPost(slug string) *model.Article {
	return a.posts.Get(slug)
}

func (a *App) GetAllPosts() []*model.Article {
	return a.posts.GetAll()
}

// GetRecentPosts returns a slice of the most recent articles sorted by
// publication time (newest first).
func (a *App) GetRecentPosts(offset, limit int) []*model.Article {
	return a.posts.GetRecent(offset, limit)
}

func (a *App) GetPage(slug string) *model.Article {
	return a.pages.Get(slug)
}

func (a *App) GetAllPages() []*model.Article {
	return a.pages.GetAll()
}

func (a *App) PostCount() int {
	return a.posts.Len()
}

func (a *App) Close() {
	a.posts.Close()
	a.pages.Close()
}
