package main

import (
	"database/sql"
	"flag"
	"log"
	"net/http"
	"os/exec"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/oauth2"

	"github.com/euphario/uploadToSalesforce/handlers"
)

var addr = flag.String("addr", "localhost:3000", "http service address")
var sqlFile = flag.String("f", "./salesforce.db", "SQLite3 database file")
var uploadFile = flag.String("upload", "", "file for upload task")
var outputFormat = flag.String("format", "csv", "result format (csv or json)")
var instanceType = flag.String("instance", "prod", "Salesforce instance type (prod or test)")
var resultFile = flag.String("output", "", "output file name")
var concurrency = flag.Int("c", 10, "number of concurrent workers")

func main() {
	finish := make(chan bool)

	flag.Parse()
	log.SetFlags(0)

	var database *sql.DB
	if sqlFile != nil {
		database, _ = sql.Open("sqlite3", *sqlFile)
	}
	// statement, _ := database.Prepare("CREATE TABLE IF NOT EXISTS instance (url TEXT PRIMARY KEY, name TEXT, token TEXT)")
	// statement.Exec()

	app := &handlers.App{
		DB:           database,
		FileUpload:   uploadFile,
		OutputFormat: outputFormat,
		Output:       resultFile,
		Concurrency:  concurrency,
		Oauth2: &oauth2.Config{
			RedirectURL:  "http://localhost:3000/oauth/_callback",
			ClientID:     "3MVG9rFJvQRVOvk5nd6A4swCyck.4BFLnjFuASqNZmmxzpQSFWSTe6lWQxtF3L5soyVLfjV3yBKkjcePAsPzi",
			ClientSecret: "9154137956044345875",
			Scopes:       []string{"api", "refresh_token", "offline_access"},
		},
	}

	if *instanceType == "prod" {
		app.Oauth2.Endpoint = oauth2.Endpoint{
			AuthURL:  "https://login.salesforce.com/services/oauth2/authorize",
			TokenURL: "https://login.salesforce.com/services/oauth2/token",
		}
	} else {
		app.Oauth2.Endpoint = oauth2.Endpoint{
			AuthURL:  "https://test.salesforce.com/services/oauth2/authorize",
			TokenURL: "https://test.salesforce.com/services/oauth2/token",
		}
	}

	server := &http.Server{
		Addr:    *addr,
		Handler: handlers.New(app),
	}

	go func() {
		log.Printf("Starting HTTP Server. Listening at %q", server.Addr)
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Printf("%v", err)
		} else {
			log.Println("Server closed!")
		}
	}()

	err := exec.Command("open", "http://localhost:3000/").Run()

	if err != nil {
		log.Printf("%v", err)
	}

	<-finish
}
