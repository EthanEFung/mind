package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/ethanefung/mind/controllers"
	"github.com/ethanefung/mind/models"
	"github.com/ethanefung/mind/templates"
	"github.com/ethanefung/mind/views"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
	"abcdefghijklmnopqrstuvwxyz" +
	"0123456789" + 
	"!@#$%^&*()_+-=;"

var password = createPassword(20)

func createPassword(length int) string {
	var seededRand *rand.Rand = rand.New(
		rand.NewSource(time.Now().UnixNano()),
	)
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	if username == "" {
		http.Error(w, "username required", http.StatusForbidden)
	}

	r.SetBasicAuth(username, password)
	token := r.Header["Authorization"][0]
  w.Header().Add("Authorization", token)
	// fmt.Printf("request headers: %v\n", r.Header)

	fmt.Printf("headers: %v\n", w.Header())

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
		r.Use(controllers.LobbyContext)
		r.Get("/", controllers.StaticHandler(tpl))
		r.Get("/ws", controllers.WSHandler(hub))
	})

	tpl = views.Must(views.ParseFS(templates.FS, "room.gohtml"))

	r.Route("/room/{roomID}", func(r chi.Router) {
		r.Use(controllers.RoomCtx)
		r.Get("/", controllers.StaticHandler(tpl))
		r.Get("/ws", controllers.WSHandler(hub))
	})

	r.NotFound(notFoundHandler)
	http.ListenAndServe(":3000", r)
}
