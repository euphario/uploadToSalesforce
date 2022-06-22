package handlers

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const API_URL = "/services/data/v52.0"

func (app *App) get(w http.ResponseWriter, r *http.Request, url string) (*[]byte, error) {
	session, err := getSession(w, r)
	if err != nil {
		fmt.Fprintf(w, "Error: %s", err.Error())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 900*time.Second)
	defer cancel()

	c := app.Oauth2.Client(ctx, &session.Token)

	// const query = "SELECT Id, ContentDocumentId, Title, Checksum, FileExtension, PathOnClient FROM ContentVersion WHERE IsDeleted=false AND IsLatest=true"

	response, err := c.Get(session.Instance + url)
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
