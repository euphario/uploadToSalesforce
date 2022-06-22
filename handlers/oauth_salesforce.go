package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

func (app *App) oauthSalesforceLogin(w http.ResponseWriter, r *http.Request) {
	oauthState := generateStateOauthCookie(w)

	u := app.Oauth2.AuthCodeURL(oauthState)
	http.Redirect(w, r, u, http.StatusTemporaryRedirect)

}

func generateStateOauthCookie(w http.ResponseWriter) string {
	var expiration = time.Now().Add(20 * time.Minute)

	b := make([]byte, 16)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)
	cookie := http.Cookie{Name: "oauthstate", Value: state, Expires: expiration}
	http.SetCookie(w, &cookie)

	return state
}

func (app *App) oauthSalesforceCallback(w http.ResponseWriter, r *http.Request) {
	oauthState, err := r.Cookie("oauthstate")
	if err != nil {
		log.Println(err.Error())
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	if r.FormValue("state") != oauthState.Value {
		log.Println("invalid oauth salesforce state")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	salesforceOauth, err := app.getUserDataFromSalesforce(r.FormValue("code"))
	if err != nil {
		log.Println(err.Error())
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	if err != nil {
		log.Println(err.Error())
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// tx, err := app.DB.Begin()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// stmt, err := tx.Prepare("INSERT OR REPLACE INTO instance (url, token) VALUES (?, ?)")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer stmt.Close()

	url := salesforceOauth.Instance
	token, err := json.Marshal(salesforceOauth.Token)
	if err != nil {
		log.Fatal(err)
	}

	// fmt.Printf("token from salesforce: %s", salesforceOauth.Token.AccessToken)

	// _, err = stmt.Exec(url, string(token))
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// tx.Commit()

	session, _ := store.Get(r, "X-Salesforce-Session")
	session.Values["instance"] = url
	session.Values["token"] = token
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func (app *App) getUserDataFromSalesforce(code string) (*SalesforceOauth, error) {

	token, err := app.Oauth2.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("code exchange wrong: %s", err.Error())
	}

	instance_url := fmt.Sprintf("%v", token.Extra("instance_url"))

	instance := SalesforceOauth{
		Instance: instance_url,
		Token:    *token,
	}
	return &instance, nil
}
