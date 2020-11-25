package store

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"path/filepath"
	"presence/model"
	"regexp"
	"strconv"
	"time"
	"unicode/utf8"
)

func (as *ArticleStore) loadArticle(filename string) (*model.Article, error) {
	slug, pubtime, err := parseFilename(filepath.Base(filename))
	if err != nil {
		return nil, err
	}

	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	if !utf8.Valid(contents) {
		return nil, errors.New("file contains invalid UTF-8")
	}

	title, body, err := extractTitle(contents)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if len(body) == 0 {
		buf.WriteString("<p>(empty)</p>")
	} else {
		err := as.markdown.Convert(body, &buf)
		if err != nil {
			return nil, err
		}
	}

	article := &model.Article{
		Slug:     slug,
		PubTime:  pubtime,
		Title:    title,
		BodyRaw:  contents,
		BodyHTML: template.HTML(buf.String()),
		Filename: filename,
	}

	return article, nil
}

func makeFilename(article *model.Article) string {
	if article.PubTime == nil {
		return article.Slug + ".md"
	}
	return fmt.Sprintf("%s.%d.md", article.Slug, article.PubTime.Unix())
}

// extractTitle extracts the first heading from the markdown-formatted
// text. It returns the title and the remainder of the document.
func extractTitle(text []byte) (string, []byte, error) {
	reader := bufio.NewReader(bytes.NewReader(text))
	line, err := reader.ReadString(byte('\n'))
	if err != nil && err != io.EOF {
		return "", text, err
	}
	r := regexp.MustCompile(`(?U)^\s*#\s+(.+)(?:\s+#*\s*)?$`)
	m := r.FindStringSubmatch(line)
	if m == nil {
		return "", text, nil
	}
	var remainder bytes.Buffer
	if _, err := remainder.ReadFrom(reader); err != nil {
		return "", text, err
	}
	return m[1], remainder.Bytes(), nil
}

var reFilename *regexp.Regexp = regexp.MustCompile(
	`^([a-zA-Z0-9\-_]+)(?:\.(-?\d+))?\.md$`,
)

func isValidFilename(filename string) bool {
	return reFilename.MatchString(filepath.Base(filename))
}

// parseFilename extracts and returns the slug and pubtime from the
// filename.
func parseFilename(filename string) (string, *time.Time, error) {
	m := reFilename.FindStringSubmatch(filename)
	if m == nil {
		msg := fmt.Sprintf("invalid article filename: '%s'", filename)
		return "", nil, errors.New(msg)
	}

	var t *time.Time
	if m[2] != "" {
		epoch, err := strconv.ParseInt(m[2], 10, 64)
		if err != nil {
			msg := fmt.Sprintf("invalid article filename: '%s' "+
				"(timestamp out of range)", filename)
			return "", nil, errors.New(msg)
		}
		tval := time.Unix(epoch, 0)
		t = &tval
	}

	return m[1], t, nil
}
