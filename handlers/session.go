package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"golang.org/x/oauth2"
)

func getSession(w http.ResponseWriter, r *http.Request) (*SalesforceOauth, error) {
	session, err := store.Get(r, "X-Salesforce-Session")
	if err != nil {
		return nil, err
	}

	instance, ok := session.Values["instance"].(string)
	if instance == "" || !ok {
		return nil, errors.New("no instance found")
	}

	tokenString, ok := session.Values["token"].([]byte)
	if tokenString == nil || !ok {
		return nil, errors.New("no token found")
	}

	var token oauth2.Token
	json.Unmarshal(tokenString, &token)

	return &SalesforceOauth{Instance: instance, Token: token}, nil
}

func (app *App) storeSession(session *SalesforceOauth) error {

	// stmt, err := app.DB.Prepare("select token from instance")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer stmt.Close()
	// var name string
	// err = stmt.QueryRow("1").Scan(&name)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Println(name)
	return nil
}
