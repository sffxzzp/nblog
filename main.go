package main

import (
	"embed"
	"fmt"
	"html/template"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type (
	tplData map[string]interface{}
)

//go:embed templates css favicon.ico
var embedFS embed.FS

func formatTime(inTime time.Time, ifTime bool) string {
	var formatStr string
	if ifTime {
		formatStr = "2006-01-02 15:04:05"
	} else {
		formatStr = "2006-01-02"
	}
	return inTime.Format(formatStr)
}

func summary(instr string, max int) string {
	runestr := []rune(instr)
	strlen := len(runestr)
	if strlen < max {
		max = strlen
	}
	return string(runestr[:max]) + " ..."
}

func htmlSafe(html string) template.HTML {
	return template.HTML(html)
}

func initRoutes(router *mux.Router, config config, notion *NotionAPI, posts []PostData, pageSize int) {
	postsCount := len(posts)
	common := tplData{
		"Title":   config.SiteName,
		"FavIcon": config.FavIcon,
		"Start":   config.Start,
		"More":    config.More,
		"Now":     time.Now().Year(),
	}
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.FS(embedFS))))
	tpl := template.Must(template.New("").Funcs(template.FuncMap{
		"formatTime": formatTime,
		"summary":    summary,
		"htmlSafe":   htmlSafe,
	}).ParseFS(embedFS, "templates/*"))
	router.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		http.Redirect(rw, r, "/page/1", http.StatusTemporaryRedirect)
	})
	router.HandleFunc("/page/{pageNum}", func(rw http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		pageNum, err := strconv.Atoi(vars["pageNum"])
		maxPage := int(math.Ceil(float64(postsCount) / float64(pageSize)))
		if err != nil || pageNum < 1 {
			http.Redirect(rw, r, "/page/1", http.StatusTemporaryRedirect)
		}
		if pageNum > maxPage {
			http.Redirect(rw, r, fmt.Sprintf("/page/%d", maxPage), http.StatusTemporaryRedirect)
		}
		min := (pageNum - 1) * pageSize
		if min < 0 {
			min = 0
		}
		max := pageNum * pageSize
		if max > postsCount {
			max = postsCount
		}
		body := common
		body["Posts"] = posts[min:max]
		body["More"] = config.More
		body["Prev"] = pageNum - 1
		body["Next"] = pageNum + 1
		if pageNum == maxPage {
			body["Next"] = 0
		}
		tpl.ExecuteTemplate(rw, "index.html", body)
	})
	router.HandleFunc("/post/{id}", func(rw http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		pid := vars["id"]
		matched := false
		for _, post := range posts {
			if post.ID == pid {
				matched = true
				body := common
				body["Post"] = post
				tpl.ExecuteTemplate(rw, "page.html", body)
			}
		}
		if !matched {
			fmt.Fprintf(rw, "404 Not Found!")
		}
	})
	router.HandleFunc("/update", func(rw http.ResponseWriter, r *http.Request) {
		posts = notion.getAllData()
		postsCount = len(posts)
		fmt.Fprintf(rw, "update success!")
	})
}

func main() {
	config := initConfigs()

	notion := newNotionAPI(config.APIKey)
	notion.database(config.Database)
	posts := notion.getAllData()

	if config.Debug {
		config.IP = "127.0.0.1"
	}
	router := mux.NewRouter()
	initRoutes(router, config, notion, posts, 10)
	listen := fmt.Sprintf("%s:%d", config.IP, config.Port)
	fmt.Println("NBlog listening at: " + listen)
	srv := &http.Server{
		Handler:      router,
		Addr:         listen,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	srv.ListenAndServe()
}
