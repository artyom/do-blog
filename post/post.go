package post

import (
	"crypto/md5"
	"fmt"
	"html"
	"html/template"
	"io"
	"io/ioutil"
	"strings"
	"time"

	html5 "code.google.com/p/go.net/html"
	"github.com/artyom/do-blog/helpers"
	"github.com/artyom/plist"
	"github.com/russross/blackfriday"
)

type ShallowPost struct {
	CreationDate time.Time
	Title        string
	Teaser       string
	Filename     string
}

type BlogPost struct {
	UUID          string
	CreationDate  time.Time `plist:"Creation Date"`
	TimeZone      string    `plist:"Time Zone"`
	EntryText     string    `plist:"Entry Text"`
	Starred       bool
	Tags          []string
	Stripped      bool
	Title         string
	Teaser        string
	EntryTextHTML template.HTML
}

func (post *BlogPost) Shallow() *ShallowPost {
	return &ShallowPost{
		CreationDate: post.CreationDate,
		Title:        post.Title,
		Teaser:       post.Teaser,
		Filename:     post.Filename(),
	}
}

// Populate extra fields: html, title, teaser
func (post *BlogPost) bake() error {
	loc, err := time.LoadLocation(post.TimeZone)
	if err != nil {
		return err
	}
	post.CreationDate = post.CreationDate.In(loc)

	post.EntryText = html.UnescapeString(post.EntryText)
	post.EntryTextHTML = template.HTML(Markdown([]byte(post.EntryText)))
	doc, err := html5.Parse(strings.NewReader(string(post.EntryTextHTML)))
	if err != nil {
		return err
	}
	_, post.Teaser = helpers.GetFirstElement(doc, "p")
	_, post.Title = helpers.GetFirstElement(doc, "h1")
	return nil
}

func (post *BlogPost) HasTag(tag string) (found bool) {
	if len(tag) == 0 {
		return
	}
	for _, v := range post.Tags {
		if v == tag {
			return true
		}
	}
	return
}

func (post *BlogPost) Filename() string {
	h := md5.New()
	io.WriteString(h, post.UUID)
	return fmt.Sprintf("%x.html", h.Sum(nil))
}

func NewPostFromFile(filename string) (post *BlogPost, err error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	post = new(BlogPost)
	if err = plist.Unmarshal(data, post); err != nil {
		return nil, err
	}
	if err := post.bake(); err != nil {
		return nil, err
	}
	return post, nil
}

func Markdown(input []byte) []byte {
	// set up the HTML renderer
	htmlFlags := 0
	htmlFlags |= blackfriday.HTML_USE_SMARTYPANTS
	htmlFlags |= blackfriday.HTML_SMARTYPANTS_FRACTIONS
	htmlFlags |= blackfriday.HTML_SMARTYPANTS_LATEX_DASHES
	htmlFlags |= blackfriday.HTML_SKIP_SCRIPT
	renderer := blackfriday.HtmlRenderer(htmlFlags, "", "")

	// set up the parser
	extensions := 0
	extensions |= blackfriday.EXTENSION_FOOTNOTES
	extensions |= blackfriday.EXTENSION_NO_INTRA_EMPHASIS
	extensions |= blackfriday.EXTENSION_TABLES
	extensions |= blackfriday.EXTENSION_AUTOLINK
	extensions |= blackfriday.EXTENSION_STRIKETHROUGH
	extensions |= blackfriday.EXTENSION_SPACE_HEADERS

	return blackfriday.Markdown(input, renderer, extensions)
}
