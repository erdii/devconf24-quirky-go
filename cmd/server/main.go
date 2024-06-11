package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/mux"
	"github.com/jaschaephraim/lrserver"
)

const (
	staticDir = "./static"
	slidesDir = "./slides"
)

var render = false

var tpl = template.Must(template.New("").Funcs(template.FuncMap{
	"safeHTML": func(html string) template.HTML {
		return template.HTML(html)
	},
	"safeJS": func(js string) template.JS {
		return template.JS(js)
	},
}).ParseFiles("template.html"))

type slide struct {
	Raw string
}

func listSlides(dir string) ([]slide, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("listing slides: %w", err)
	}

	slides := []slide{}
	for _, entry := range entries {
		path := filepath.Join(dir, entry.Name())
		b, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("reading file: %w", err)
		}
		slides = append(slides, slide{
			Raw: string(b),
		})
	}
	return slides, nil
}

func main() {
	if render {
		slides, err := listSlides(slidesDir)
		if err != nil {
			panic(fmt.Errorf("listing slides: %w", err))
		}

		f, err := os.Create("index.html")
		if err != nil {
			panic(fmt.Errorf("creating file: %w", err))
		}

		defer func() {
			if cerr := f.Close(); cerr != nil && err == nil {
				panic(fmt.Errorf("closing file: %w", cerr))
			} else if cerr != nil {
				panic(fmt.Errorf("closing file: %w while handling another error: %w", cerr, err))
			}
		}()

		if err := tpl.ExecuteTemplate(os.Stdout, "template.html", map[string]any{
			"slides": slides,
			"islive": false,
		}); err != nil {
			panic(err)
		}
		return
	}

	// Create file watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalln(err)
	}
	defer watcher.Close()

	err = watcher.Add(slidesDir)
	if err != nil {
		log.Fatalln(err)
	}

	entries, err := os.ReadDir(slidesDir)
	if err != nil {
		log.Fatal(err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if err := watcher.Add(filepath.Join(slidesDir, entry.Name())); err != nil {
			log.Fatal(err)
		}
	}

	// Create and start LiveReload server
	lr := lrserver.New(lrserver.DefaultName, lrserver.DefaultPort)
	go func() {
		if err := lr.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	// Start goroutine that requests reload upon watcher event
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				lr.Reload(event.Name)
				if event.Has(fsnotify.Rename) {
					continue
				}

				info, err := os.Stat(event.Name)
				if err != nil {
					log.Fatal(err)
				}
				if !info.IsDir() {
					continue
				}
				if event.Has(fsnotify.Create) {
					if err := watcher.Add(event.Name); err != nil {
						log.Fatal(err)
					}
					continue
				}
			case err := <-watcher.Errors:
				log.Println(err)
			}
		}
	}()

	router := mux.NewRouter()
	fserver := http.FileServer(http.Dir(staticDir))

	router.PathPrefix("/static/").Handler(http.StripPrefix("/static", fserver))

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "error parsing template: %s", err)
			return
		}

		slides, err := listSlides(slidesDir)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "error listing slides: %s", err)
		}

		if err := tpl.ExecuteTemplate(w, "template.html", map[string]any{
			"slides": slides,
			"islive": true,
		}); err != nil {
			panic(err)
		}
	})

	srv := &http.Server{
		Handler:      router,
		Addr:         "127.0.0.1:8000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	fmt.Println("will listen on http://" + srv.Addr)

	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
