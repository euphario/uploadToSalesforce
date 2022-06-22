package handlers

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

var (
	// key must be 16, 24 or 32 bytes long (AES-128, AES-192 or AES-256)
	key       = []byte("o73PiUcKGtcuEnQo44AsMJ84t4EddLTv")
	store     = sessions.NewCookieStore(key)
	templates = ParseTemplates()
)

func New(app *App) http.Handler {
	staticDir := "/public/"

	mux := mux.NewRouter().StrictSlash(true)

	// Root

	mux.PathPrefix(staticDir).Handler(http.StripPrefix(staticDir, http.FileServer(http.Dir("./frontend"+staticDir))))

	mux.HandleFunc("/echo", app.socketHandler)

	mux.HandleFunc("/oauth/login", app.oauthSalesforceLogin)
	mux.HandleFunc("/oauth/_callback", app.oauthSalesforceCallback)

	mux.HandleFunc("/", home)

	// mux.HandleFunc("/describe/{sobject}", app.Describe).Methods(http.MethodGet)
	// mux.HandleFunc("/salesforce/old", oldHandler)
	// mux.HandleFunc("/salesforce/oldLinks", oldHandlerLinks)
	// mux.HandleFunc("/salesforce/oldFunding", oldHandlerFunding)
	mux.HandleFunc("/salesforce/downloads", app.Downloads).Methods(http.MethodGet)
	mux.HandleFunc("/salesforce/uploads", app.Uploads).Methods(http.MethodGet)

	mux.HandleFunc("/salesforce/links", app.QueryDocumentLink).Methods(http.MethodGet)

	return mux
}
