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
	"time"

	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/parser"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func htmlWrapper(hf http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "<html>")
		fmt.Fprintln(w, "<head>")
		fmt.Fprintln(w, "<meta charset=\"UTF-8\" />")
		fmt.Fprintln(w, "<meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\" />")
		fmt.Fprintln(w, `<style>
			@import url("/fonts/fonts.css");
			body {
				font-family: InterDisplay, sans-serif;
				font-weight: 400;
				max-width: 960px;
				padding: 25px;
				margin: 0 auto;
			}
		</style>`)
		fmt.Fprintln(w, "<title>Morgan Gallant</title>")
		fmt.Fprintln(w, "</head>")
		fmt.Fprintln(w, "<body>")
		fmt.Fprintln(w, "<hr />")
		start := time.Now()
		hf(w, r)
		fmt.Fprintln(w, "<hr />")
		fmt.Fprintf(w, "<p style=\"text-align: center;\">Made with love, and <a href=\"https://golang.org\">Go</a>. This website is <a href=\"https://github.com/morgangallant/homepage\">open source</a>. Rendered page in %d ms.</p>\n", time.Since(start).Milliseconds())
		fmt.Fprintln(w, "</body>")
		fmt.Fprintln(w, "</html>")
	}
}

//go:embed static
var embedded embed.FS

func run() error {
	// md := goldmark.New()
	bdir, err := fs.Sub(embedded, "static/blog")
	if err != nil {
		return err
	}
	bh, err := blogHandler(bdir)
	if err != nil {
		return err
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/blog/", htmlWrapper(bh))
	fdir, err := fs.Sub(embedded, "static/fonts")
	if err != nil {
		return err
	}
	mux.Handle("/fonts/", http.StripPrefix("/fonts/", http.FileServer(http.FS(fdir))))
	server := &http.Server{
		Addr:         ":8888",
		Handler:      mux,
		ReadTimeout:  time.Second,
		WriteTimeout: 10 * time.Second,
	}
	return server.ListenAndServe()
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
		p.content.WriteTo(w)
	}, nil
}
