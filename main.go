package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/ethanefung/mind/controllers"
	"github.com/ethanefung/mind/models"
	"github.com/ethanefung/mind/templates"
	"github.com/ethanefung/mind/views"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func executeTemplate(w http.ResponseWriter, path string, data interface{}) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tpl, err := template.ParseFiles(path)
	if err != nil {
		log.Printf("Error Parsing the file with path \"%s\"", path)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	err = tpl.Execute(w, data)
	if err != nil {
		log.Printf("Error executing template with path \"%s\": %v", path, err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	if username == "" {
		http.Error(w, "username required", http.StatusForbidden)
	}

	/*
	  here we will want to create some credentials for
	  the user so that when the user hits the lobby or
	  any room routes, that a middleware will be able to
	  authenticate the user with.
	*/
	http.SetCookie(w, &http.Cookie{
		Name:   "auth",
		Value:  "cookie",
		MaxAge: 300,
	})
	http.Redirect(w, r, "/lobby", http.StatusSeeOther)
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(w, "Resource Not Found")
}

func main() {
	r := chi.NewRouter()
	hub := models.NewHub()

	go hub.Run()

	r.Use(middleware.Logger)

	tpl := views.Must(views.ParseFS(templates.FS, "home.gohtml"))
	r.Route("/", func(r chi.Router) {
		r.Get("/", controllers.StaticHandler(tpl))
		r.Post("/", loginHandler)
	})

	tpl = views.Must(views.ParseFS(templates.FS, "lobby.gohtml"))
	r.Route("/lobby", func(r chi.Router) {
		r.Get("/", controllers.StaticHandler(tpl))
		r.Get("/ws", controllers.WSHandler(hub))
	})

	tpl = views.Must(views.ParseFS(templates.FS, "room.gohtml"))
	r.Route("/room/{roomID}", func(r chi.Router) {
		r.Get("/", controllers.StaticHandler(tpl))
	})

	r.NotFound(notFoundHandler)
	http.ListenAndServe(":3000", r)
}
