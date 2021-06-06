// Homepage is the binary powering https://morgangallant.com.
package main

import (
	"bytes"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"strings"

	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/parser"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

//go:embed blog
var embedded embed.FS

func run() error {
	// md := goldmark.New()
	bdir, err := fs.Sub(embedded, "blog")
	if err != nil {
		return err
	}
	bh, err := blogHandler(bdir)
	if err != nil {
		return err
	}
	http.HandleFunc("/blog/", bh)
	return http.ListenAndServe(":8888", nil)
}

func extractSlug(uri string) string {
	uri = strings.TrimPrefix(uri, "/blog/")
	uri = strings.ToLower(uri)
	if idx := strings.Index(uri, "#"); idx != -1 {
		uri = uri[:idx]
	}
	uri = strings.TrimSuffix(uri, "/")
	return uri
}

func writeBlogIndex(w http.ResponseWriter, posts map[string]post) {
	fmt.Fprintln(w, "<p>Blog Posts:</p>")
	fmt.Fprintln(w, "<ul>")
	for _, p := range posts {
		fmt.Fprintf(w, "<li><a href=\"/blog/%s\">%s</a></li>\n", p.slug, p.name)
	}
	fmt.Fprintf(w, "</ul>")
	w.Header().Set("Content-Type", "text/html")
}

type post struct {
	name    string
	slug    string
	content bytes.Buffer
}

func processPost(bfs fs.FS, fname string) (post, error) {
	f, err := bfs.Open(fname)
	if err != nil {
		return post{}, err
	}
	defer f.Close()
	contents, err := io.ReadAll(f)
	if err != nil {
		return post{}, err
	}
	p := post{}
	md := goldmark.New(
		goldmark.WithExtensions(
			meta.Meta,
		),
	)
	ctx := parser.NewContext()
	if err := md.Convert(contents, &p.content, parser.WithContext(ctx)); err != nil {
		return post{}, err
	}
	extra := meta.Get(ctx)
	p.name = extra["title"].(string)
	p.slug = extra["slug"].(string)
	return p, nil
}

func blogHandler(bfs fs.FS) (http.HandlerFunc, error) {
	entries, err := fs.ReadDir(bfs, ".")
	if err != nil {
		return nil, err
	}
	posts := make(map[string]post)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		post, err := processPost(bfs, entry.Name())
		if err != nil {
			return nil, err
		}
		if _, ok := posts[post.slug]; ok {
			return nil, fmt.Errorf("post with slug %s already exists", post.slug)
		}
		posts[post.slug] = post
	}
	return func(w http.ResponseWriter, r *http.Request) {
		slug := extractSlug(r.RequestURI)
		if slug == "" {
			writeBlogIndex(w, posts)
			return
		}
		p, ok := posts[slug]
		if !ok {
			fmt.Fprintf(w, "Invalid slug %s.", slug)
			return
		}
		if _, err := p.content.WriteTo(w); err != nil {
			log.Printf("failed to write content: %v", err)
			http.Error(w, "failed to write content", http.StatusInternalServerError)
			return
		}
	}, nil
}
