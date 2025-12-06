package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

var tpl *template.Template

func init() {
	files, err := filepath.Glob("pages/*.html")
	if err != nil {
		log.Fatal(err)
	}

	if len(files) == 0 {
		log.Fatal("No HTML files found in pages directory")
	}

	tpl = template.Must(template.ParseFiles(files...))

}

func indexHandler(w http.ResponseWriter, r *http.Request) {

	data := map[string]interface{}{
		"isLoggedIn": false,
		"username":   "Tytyber",
		"rules":      "3",
		//"date":       profile.DateRegistry,
		"money": 1000 / 2,
	}
	if data["isLoggedIn"] == true {
		err := tpl.ExecuteTemplate(w, "main.html", data)
		if err != nil {
			http.Error(w, "Ошибка при рендере страницы", http.StatusInternalServerError)
		}
	} else {
		err := tpl.ExecuteTemplate(w, "index.html", data)
		if err != nil {
			http.Error(w, "Ошибка при рендере страницы", http.StatusInternalServerError)
		}
	}
}

func notFoundPageHundler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)

	// опционально: передаём данные в шаблон
	data := struct {
		URL string
	}{
		URL: r.URL.Path,
	}

	// рендерим именно 404.html
	err := tpl.ExecuteTemplate(w, "404.html", data)
	if err != nil {
		// если шаблон упал — возвращаем простой текст
		http.Error(w, "Ошибка при рендере страницы 404", http.StatusInternalServerError)
	}
}

func ProfileHundler(w http.ResponseWriter, r *http.Request) {

	data := map[string]interface{}{
		"isLoggedIn": true,
		"username":   "Tytyber",
		"rules":      "3",
		//"date":       profile.DateRegistry,
		"money": 1000 / 2,
	}

	err := tpl.ExecuteTemplate(w, "profile.html", data)
	if err != nil {
		http.Error(w, "Ошибка рендеринга страницы profile", http.StatusInternalServerError)
	}

}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("assets"))
	mux.Handle("/assets/", http.StripPrefix("/assets/", fs))

	fmt.Println("Server starting on", port)

	mux.HandleFunc("/", indexHandler)
	mux.HandleFunc("/notFound", notFoundPageHundler)
	mux.HandleFunc("/profile", ProfileHundler)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h, pattern := mux.Handler(r)

		switch {
		case pattern == "":
			// Ни один маршрут не подошел
			notFoundPageHundler(w, r)
			return
		case pattern == "/" && r.URL.Path != "/":
			// Если бы мы ловили "/" — а это не /
			notFoundPageHundler(w, r)
			return
		default:
			// Всё ок — передаём исполнению оригинальный хендлер
			h.ServeHTTP(w, r)
		}
	})
	log.Println("|-| Сервер запущен на порту", port, " |-|")
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
