package main

import (
	"bytes"
	"flag"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/artyom/do-blog/post"
)

type IndexByTime []*post.ShallowPost

var (
	indexTemplate, postTemplate *template.Template
	IndexList                   IndexByTime

	tplPath     string
	journalPath string
	outPath     string
	tag         string
)

func init() {
	flag.StringVar(&journalPath, "journal", "Journal.dayone", "path to Day One journal")
	flag.StringVar(&outPath, "savedir", "/tmp/out", "output directory")
	flag.StringVar(&tag, "tag", "blog", "tag to filter journal items")
	flag.StringVar(&tplPath, "templates", "templates", "templates directory")
}

func main() {
	flag.Parse()

	funcMap := template.FuncMap{
		"short": short,
	}
	var err error
	postTemplate, err = template.ParseFiles(
		filepath.Join(tplPath, "post.template"),
		filepath.Join(tplPath, "extra.template"),
	)
	if err != nil {
		log.Fatal(err)
	}
	indexTemplate, err = template.New("index.template").Funcs(funcMap).ParseFiles(
		filepath.Join(tplPath, "index.template"),
		filepath.Join(tplPath, "extra.template"),
	)
	if err != nil {
		log.Fatal(err)
	}

	if err := os.MkdirAll(outPath, 0755); err != nil {
		log.Fatal(err)
	}
	if err := filepath.Walk(journalPath, processItem); err != nil {
		log.Fatal(err)
	}

	// FIXME
	if len(IndexList) == 0 {
		return
	}

	sort.Sort(sort.Reverse(IndexList))

	tf, err := ioutil.TempFile("", "index-temp-")
	if err != nil {
		log.Fatal(err)
	}
	defer tf.Close()
	defer os.Remove(tf.Name())
	if err := indexTemplate.Execute(tf,
		struct {
			Items []*post.ShallowPost
		}{Items: IndexList}); err != nil {
		log.Fatal(err)
	}
	tf.Chmod(0644)
	indexPath := filepath.Join(outPath, "index.html")
	log.Printf("updating file %s", indexPath)
	if err := os.Rename(tf.Name(), indexPath); err != nil {
		log.Fatal(err)
	}
}

func processItem(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if info.IsDir() || filepath.Ext(path) != ".doentry" {
		return nil
	}
	article, err := post.NewPostFromFile(path)
	if err != nil {
		return err
	}
	if !article.HasTag(tag) {
		return nil
	}
	IndexList = append(IndexList, article.Shallow())

	buf := new(bytes.Buffer)
	if err := postTemplate.Execute(buf, article); err != nil {
		return err
	}

	resultPath := filepath.Join(outPath, article.Filename())
	resultFi, err := os.Stat(resultPath)
	if err != nil {
		goto RENAME
	}
	if resultFi.Size() == int64(buf.Len()) && resultFi.ModTime().Equal(article.CreationDate) {
		log.Printf("file %s up to date, skipping", resultPath)
		return nil
	}
RENAME:
	tf, err := ioutil.TempFile("", "article-temp-")
	if err != nil {
		return err
	}
	defer tf.Close()
	defer os.Remove(tf.Name())
	tf.Chmod(0644)
	if _, err := buf.WriteTo(tf); err != nil {
		return err
	}
	if err := os.Chtimes(tf.Name(), article.CreationDate, article.CreationDate); err != nil {
		log.Print("failed to change times: ", err)
	}
	log.Printf("updating file %s", resultPath)
	return os.Rename(tf.Name(), resultPath)
}

// comply to sort.Interface
func (a IndexByTime) Len() int           { return len(a) }
func (a IndexByTime) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a IndexByTime) Less(i, j int) bool { return a[i].CreationDate.Before(a[j].CreationDate) }

// short limits length of teaser to specified number of words
func short(input string, limit int) (output string) {
	words := strings.Fields(input)
	if len(words) < limit {
		limit = len(words)
	}
	return strings.Join(words[:limit], " ")
}
