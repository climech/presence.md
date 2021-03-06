package model

import (
	"time"
)

// Article represents a blog post or a page.
type Article struct {
	Slug     string
	Title    string
	PubTime  *time.Time
	Filename string
	BodyRaw  []byte
	BodyHTML string
}
