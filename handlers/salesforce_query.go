package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

func (app *App) put(w http.ResponseWriter, r *http.Request, url string, post string) (*[]byte, error) {
	session, err := getSession(w, r)
	if err != nil {
		fmt.Fprintf(w, "Error: %s", err.Error())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 900*time.Second)
	defer cancel()

	c := app.Oauth2.Client(ctx, &session.Token)
	// const query = "SELECT Id, ContentDocumentId, Title, Checksum, FileExtension, PathOnClient FROM ContentVersion WHERE IsDeleted=false AND IsLatest=true"

	response, err := c.Post(session.Instance+url, "application/json", strings.NewReader(post))
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return &body, nil
}

func prepareSqlTable(describe Describe) ([]string, error) {
	var columns []string
	var indexes []string
	for _, v := range describe.Fields {
		if v.Name == "Id" {
			columns = append(columns, fmt.Sprintf("%s TEXT PRIMARY KEY", v.Name))
		} else {
			columns = append(columns, fmt.Sprintf("%s TEXT", v.Name))

		}
		if v.Type == "reference" {
			indexes = append(indexes, fmt.Sprintf("CREATE INDEX %s ON %s (%s)", v.Name, describe.Name, v.Name))
		}
	}

	var statements []string
	statements = append(statements, fmt.Sprintf("DROP TABLE IF EXISTS %s", describe.Name))
	statements = append(statements, fmt.Sprintf("CREATE TABLE %s (%s)", describe.Name, strings.Join(columns, ", ")))
	statements = append(statements, indexes...)
	return statements, nil
}

func (app *App) prepareSqlQuery(describe Describe) string {
	var columns []string
	for _, v := range describe.Fields {
		columns = append(columns, v.Name)
	}
	return fmt.Sprintf("SELECT %s FROM %s", strings.Join(columns, ", "), describe.Name)
}

func (app *App) Describe(w http.ResponseWriter, r *http.Request, sobject string) {
	var query = strings.Join([]string{API_URL, "sobjects", sobject, "describe"}, "/")
	body, err := app.get(w, r, query)
	if err != nil {
		fmt.Fprintf(w, "Error: %s\n", err.Error())
		return
	}
	var describe Describe
	err = json.Unmarshal(*body, &describe)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	// statements, err := prepareSqlTable(*body)
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// }

	// err = app.execSqlStatements(statements)
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// }

	// fmt.Fprintf(w, "%s", body)
}
