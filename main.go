package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/NYTimes/gziphandler"
)

type markdown string

type Post struct {
	Title []string
	Date  time.Time
	Image []string
	Draft bool
	Body  markdown
}

func main() {
	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("static"))
	gfs := gziphandler.GzipHandler(Headers(fs))

    im := http.FileServer(http.Dir("posts/images"))
    gim := gziphandler.GzipHandler(Headers(im))

    da := http.FileServer(http.Dir("posts/data"))
    gda := gziphandler.GzipHandler(Headers(da))

	head := func(f func(w http.ResponseWriter, r *http.Request)) http.Handler {
		return gziphandler.GzipHandler(Headers(http.HandlerFunc(f)))
		// return Headers(http.HandlerFunc(f))
	}

	// fmt.Printf("Enabling HTTP Redirect...\n")
	// go httpRedirect()

	// mux.Handle("/", hfs)
	mux.Handle("/", gfs)
    mux.Handle("/images/", http.StripPrefix("/images/", gim))
    mux.Handle("/data/", http.StripPrefix("/data/", gda))
	mux.Handle("/api/latest", head(LatestPost))
	mux.Handle("/api/all", head(AllPosts))
	mux.Handle("/api/draft", head(AllDrafts))
	mux.Handle("/api/get", head(GetPost))
	mux.Handle("/api/search", head(SearchPosts))
	mux.Handle("/api/markdown", head(GetMarkdown))
	mux.Handle("/api/edit", head(EditPost))
	mux.Handle("/api/delete", head(DeletePost))

	fmt.Printf("Configuring HTTPS Security...\n")
	cfg := configureTLS()
	srv := configureServer(mux, cfg)

	fmt.Print("Server Starting...\n")
	log.Fatal(srv.ListenAndServe())
}

func Headers(fs http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
		w.Header().Set("Content-Security-Policy", "script-src 'self' unpkg.com; connect-src 'self'; style-src 'self' 'unsafe-inline' unpkg.com maxcdn.bootstrapcdn.com; font-src 'self' unpkg.com maxcdn.bootstrapcdn.com; img-src https: data:; media-src 'self'; object-src 'self'; frame-src 'self' replit.com www.youtube-nocookie.com; frame-ancestors 'self' replit.com; form-action 'self'; base-uri 'self'; default-src 'none'; upgrade-insecure-requests")
		w.Header().Set("X-XSS-Protection", "1")
		w.Header().Set("X-Frame-Options", "samedomain, replit.com")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Expect-CT", "max-age=7776000, enforce")
		w.Header().Set("Permissions-Policy", "encrypted-media=(self), fullscreen=(self), sync-xhr=(self)")
		w.Header().Set("Cache-Control", "max-age=86400")
		w.Header().Set("Set-Cookie", "__Secure-munch=true; Secure; HttpOnly; SameSite=Strict")
		// w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Date", "Mon, 21 May 2021 19:19:19 GMT")
		w.Header().Set("Referrer-Policy", "no-referrer")

		fs.ServeHTTP(w, r)
	}
}

func httpRedirect() {
	redirectToHTTPS := func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Connection", "close")
		url := "https://" + req.Host + req.URL.String()
		http.Redirect(w, req, url, http.StatusMovedPermanently)
	}
	srv := &http.Server{
		Addr:         ":80",
		Handler:      http.HandlerFunc(redirectToHTTPS),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}

func configureServer(mux *http.ServeMux, cfg *tls.Config) *http.Server {
	return &http.Server{
		Addr:         ":443",
		Handler:      mux,
		TLSConfig:    cfg,
		ReadTimeout:  2 * time.Minute,
		WriteTimeout: 2 * time.Minute,
		IdleTimeout:  2 * time.Minute,
	}
}

func configureTLS() *tls.Config {
	return &tls.Config{
		MinVersion:               tls.VersionTLS12,
		PreferServerCipherSuites: true,
		CurvePreferences: []tls.CurveID{
			tls.CurveP256,
			tls.X25519,
		},
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},
	}
}

func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

func getMeta(pair string) []string {
	pair = strings.Replace(pair, " ", "", 1)
	pair = strings.ReplaceAll(pair, ", ", ",")
	pair = strings.Join(strings.Split(pair, ":")[1:], ":")
	return strings.Split(pair, ",")
}

func getPassword() string {
	return os.Getenv("password")
}

func setPost(blog []string) Post {
	var post Post
	post.Title = getMeta(blog[1])
	post.Date = getDate(getMeta(blog[2])[0])
	post.Image = getMeta(blog[3])
	post.Draft, _ = strconv.ParseBool(getMeta(blog[4])[0])
	post.Body = markdown(strings.Join(blog[6:], "\n"))
	return post
}

func getDate(date string) time.Time {
	// month/day/year format
	dateSli := strings.Split(date, "/")
	var day, month, year uint64 = 1, 1, 1
	if len(dateSli) == 3 {
		day, _ = strconv.ParseUint(dateSli[1], 10, 16)
		month, _ = strconv.ParseUint(dateSli[0], 10, 16)
		year, _ = strconv.ParseUint(dateSli[2], 10, 16)
	}
	return time.Date(int(year), time.Month(month), int(day), 0, 0, 0, 0, time.UTC)
}

func getFiles(path string) ([]os.FileInfo, string) {
	files, _ := ioutil.ReadDir(path)
	return files, path
}

func appendPosts(files []os.FileInfo, path string, draft bool) []Post {
	var posts []Post
	for _, file := range files {
		content, _ := ioutil.ReadFile(path + file.Name())
		blog := strings.Split(string(content), "\n")
        if len(blog) > 6 {
            post := setPost(blog)
            if post.Draft == draft && unicode.IsLower(rune(file.Name()[0])) {
                posts = append(posts, post)
            }
        }
	}
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Date.After(posts[j].Date)
	})
	return posts
}

func searchTitles(query string, files []os.FileInfo, path string) Post {
	for _, file := range files {
		content, _ := ioutil.ReadFile(path + file.Name())
		blog := strings.Split(string(content), "\n")
        if len(blog) > 6 {
            post := setPost(blog)

            if query == post.Title[0] {
                return post
            }
        }
	}
	return searchTitles("Error Page Not Found", files, path)
}

func searchBodies(query string, files []os.FileInfo, path string) []Post {
	var posts []Post
	for _, file := range files {
		content, _ := ioutil.ReadFile(path + file.Name())
		blog := strings.Split(string(content), "\n")
        if len(blog) > 6 {
            post := setPost(blog)
            title := strings.Join(post.Title, " ")

            if strings.Contains(strings.ToLower(string(post.Body))+" "+strings.ToLower(title), strings.ToLower(query)) {
                if !post.Draft {
                    preview := strings.Split(string(post.Body), " ")
                    post.Body = markdown(strings.Join(preview[:min(len(preview), 9)], " ") + " …")
                    posts = append(posts, post)
                }
            }
        }
	}
    sort.Slice(posts, func(i, j int) bool {
		return posts[i].Date.After(posts[j].Date)
	})
	return posts
}

func searchMarkdown(query string, files []os.FileInfo, path string) markdown {
	for _, file := range files {
		content, _ := ioutil.ReadFile(path + file.Name())
		blog := strings.Split(string(content), "\n")
        if len(blog) > 6 {
            post := setPost(blog)

            if query == post.Title[0] {
                return markdown(content)
            }
        }
	}
	return markdown("")
}

func searchFiles(title string, files []os.FileInfo, path string) string {
	var fileName string
	for i, word := range strings.Split(title, " ") {
		if i == 0 {
			fileName += strings.ToLower(word)
		} else {
			fileName += strings.Title(strings.ToLower(word))
		}
	}

	for _, file := range files {
		if title+".md" == file.Name() {
			return path + file.Name()
		} else {
			content, _ := ioutil.ReadFile(path + file.Name())
			blog := strings.Split(string(content), "\n")
            if len(blog) > 6 {
                titles := getMeta(blog[1])

                if title == titles[0] {
                    return path + file.Name()
                }
            }
		}
	}

	if fileName != "" {
		return path + fileName + ".md"
	}
	return ""
}

func EditPost(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	data := strings.Split(string(body), ",")
	key := data[0]
	blog := strings.Split(strings.Join(data[1:], ","), "\n")

	password := getPassword()
	if key == string(password) {
        if len(blog) > 6 {
            post := setPost(blog)
            if len(post.Title) != 0 {
                files, path := getFiles("posts/")
                fileName := searchFiles(post.Title[0], files, path)
                content := []byte(strings.Join(blog, "\n"))
                ioutil.WriteFile(fileName, content, 0644)
            }
        }
	} else {
		fmt.Fprint(w, "Denied")
	}
}

func DeletePost(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	data := strings.Split(string(body), ",")
	key := data[0]
	title := data[1]

	password := getPassword()
	if key == string(password) {
		if len(title) != 0 {
			files, path := getFiles("posts/")
			fileName := searchFiles(title, files, path)
			if fileName != "" {
				os.Remove(fileName)
			}
		}
	} else {
		fmt.Fprint(w, "Denied")
	}
}

func GetMarkdown(w http.ResponseWriter, r *http.Request) {
	query := strings.Split(r.RequestURI, "?query=")[1]
	query, _ = url.QueryUnescape(query)
	files, path := getFiles("posts/")

	if query == "Blog" {
		query = "Template Post"
	}
	post := searchMarkdown(query, files, path)

	fmt.Fprintln(w, post)
}

func GetPost(w http.ResponseWriter, r *http.Request) {
	query := strings.Split(r.RequestURI, "?query=")[1]
	query, _ = url.QueryUnescape(query)
	files, path := getFiles("posts/")
	post := searchTitles(query, files, path)
	rsp, _ := json.Marshal(post)

	fmt.Fprintln(w, string(rsp))
}

func AllPosts(w http.ResponseWriter, r *http.Request) {
	files, path := getFiles("posts/")
	posts := appendPosts(files, path, false)
	for i := range posts {
		preview := strings.Split(string(posts[i].Body), " ")
		posts[i].Body = markdown(strings.Join(preview[:min(len(preview), 9)], " ") + " …")
	}
	rsp, _ := json.Marshal(posts)

	fmt.Fprintln(w, string(rsp))
}

func AllDrafts(w http.ResponseWriter, r *http.Request) {
	files, path := getFiles("posts/")
	posts := appendPosts(files, path, true)
	for i := range posts {
		preview := strings.Split(string(posts[i].Body), " ")
		posts[i].Body = markdown(strings.Join(preview[:min(len(preview), 9)], " ") + " …")
	}
	rsp, _ := json.Marshal(posts)

	fmt.Fprintln(w, string(rsp))
}

func LatestPost(w http.ResponseWriter, r *http.Request) {
	files, path := getFiles("posts/")
	post := appendPosts(files, path, false)[:1]
	preview := strings.Split(string(post[0].Body), " ")
	post[0].Body = markdown(strings.Join(preview[:min(len(preview), 9)], " ") + " …")
	rsp, _ := json.Marshal(post)

	fmt.Fprintln(w, string(rsp))
}

func SearchPosts(w http.ResponseWriter, r *http.Request) {
	query := strings.Split(r.RequestURI, "?query=")[1]
	query, _ = url.QueryUnescape(query)
	files, path := getFiles("posts/")
	posts := searchBodies(query, files, path)
	rsp, _ := json.Marshal(posts)

	fmt.Fprintln(w, string(rsp))
}

/*
	generated 2021-08-10, Mozilla Guideline v5.6, Go 1.16.7, intermediate configuration
	https://ssl-config.mozilla.org/#server=go&version=1.16.7&config=intermediate&guideline=5.6
	https://blog.cloudflare.com/exposing-go-on-the-internet/
*/