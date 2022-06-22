package handlers

import (
	"fmt"
	"net/http"
)

// const query = "SELECT Id, ContentDocumentId, Title, Checksum, FileExtension, PathOnClient FROM ContentVersion WHERE IsDeleted=false AND IsLatest=true"

// var oauthSalesforceUrlAPI = "/services/data/v52.0/query?q=" + strings.Replace(query, " ", "+", -1)

func home(w http.ResponseWriter, r *http.Request) {
	err := templates.ExecuteTemplate(w, "home.html", "ws://"+r.Host+"/echo")
	if err != nil {
		fmt.Fprintln(w, err.Error())
	}
}
