package server

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path"
	"presence/model"
	"strconv"

	"github.com/gorilla/feeds"
	"github.com/gorilla/mux"
	"github.com/lestrrat-go/strftime"
)

type articleData struct {
	Slug  string
	Title string
	Date  string
	URL   string
	Body  template.HTML
}

func (s *Server) newArticleData(a *model.Article) *articleData {
	var date string
	if a.PubTime != nil {
		if d, err := strftime.Format(s.app.Config.DateFormat, *a.PubTime); err != nil {
			panic(err)
		} else {
			date = d
		}
	}
	return &articleData{
		Slug:  a.Slug,
		Title: a.Title,
		Date:  date,
		URL:   path.Join(s.BaseURL(), a.Slug),
		Body:  template.HTML(a.BodyHTML),
	}
}

func (s *Server) newArticleDataSlice(articles []*model.Article) []*articleData {
	result := make([]*articleData, 0, len(articles))
	for _, a := range articles {
		result = append(result, s.newArticleData(a))
	}
	return result
}

type commonData struct {
	Path        string
	Title       string
	Author      string
	Description string
	Pages       []*articleData
}

func (s *Server) newCommonData(r *http.Request) *commonData {
	return &commonData{
		Path:        r.URL.Path,
		Title:       s.app.Config.Title,
		Author:      s.app.Config.Author,
		Description: s.app.Config.Description,
		Pages:       s.newArticleDataSlice(s.app.GetAllPages()),
	}
}

func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	page := 1
	if s, ok := mux.Vars(r)["page"]; ok {
		page, _ = strconv.Atoi(s)
		if page == 0 {
			http.Error(w, "page not found", 404)
			return
		}
	}

	limit := int(s.app.Config.MaxEntriesPerPage)
	posts := s.app.GetRecentPosts(limit*(page-1), limit)
	if page != 1 && len(posts) == 0 {
		http.Error(w, "page not found", 404)
		return
	}

	next := 0
	if limit*page < s.app.PostCount() {
		next = page + 1
	}

	data := struct {
		*commonData
		Posts    []*articleData
		Previous int
		Next     int
	}{
		s.newCommonData(r),
		s.newArticleDataSlice(posts),
		page - 1,
		next,
	}

	tname := "home.html"
	t, ok := s.templates[tname]
	if !ok {
		log.Println("couldn't load template: " + tname)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if err := t.Execute(w, data); err != nil {
		log.Println(err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}

func (s *Server) handleArticle(w http.ResponseWriter, r *http.Request) {
	slug := mux.Vars(r)["slug"]

	article := s.app.GetPage(slug)
	if article == nil {
		article = s.app.GetPost(slug)
		if article == nil {
			http.Error(w, "not found", 404)
			return
		}
	} else {
		// For pages, PubTime is only used for sorting and shouldn't be displayed
		// to visitors.
		article.PubTime = nil
	}

	data := struct {
		*commonData
		Article *articleData
	}{
		s.newCommonData(r),
		s.newArticleData(article),
	}

	tname := "article.html"
	t, ok := s.templates[tname]
	if !ok {
		log.Println("couldn't load template: " + tname)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if err := t.Execute(w, data); err != nil {
		log.Println(err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}

type yearData struct {
	Year  int
	Posts []*articleData
}

func (s *Server) handleArchive(w http.ResponseWriter, r *http.Request) {
	posts := s.app.GetAllPosts()
	var years []*yearData

	var current int
	for _, p := range posts {
		y := p.PubTime.Year()
		if y != current {
			years = append(years, &yearData{y, []*articleData{}})
			current = y
		}
		yearPosts := &years[len(years)-1].Posts
		*yearPosts = append(*yearPosts, s.newArticleData(p))
	}

	data := struct {
		*commonData
		Years []*yearData
	}{
		s.newCommonData(r),
		years,
	}

	tname := "archive.html"
	t, ok := s.templates[tname]
	if !ok {
		log.Println("couldn't load template: " + tname)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if err := t.Execute(w, data); err != nil {
		log.Println(err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}

func (s *Server) handleRSS(w http.ResponseWriter, r *http.Request) {
	feed := &feeds.Feed{
		Title:       s.app.Config.Title,
		Link:        &feeds.Link{Href: s.BaseURL()},
		Description: s.app.Config.Description,
		Author:      &feeds.Author{Name: s.app.Config.Author},
	}

	maxItems := 25
	items := make([]*feeds.Item, 0, maxItems)
	posts := s.app.GetRecentPosts(0, maxItems)

	for _, p := range posts {
		items = append(items, &feeds.Item{
			Title:       p.Title,
			Link:        &feeds.Link{Href: path.Join(s.BaseURL(), p.Slug)},
			Description: p.BodyHTML,
			Created:     *p.PubTime,
		})
	}

	feed.Items = items
	rss, err := feed.ToRss()
	if err != nil {
		log.Println(err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}

	fmt.Fprintf(w, rss)
}
