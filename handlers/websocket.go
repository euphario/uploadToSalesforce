package handlers

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{} // use default options

func (app *App) socketHandler(w http.ResponseWriter, r *http.Request) {

	session, err := getSession(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("Error during connection upgrade:", err)
		return
	}
	defer conn.Close()

	// The event loop
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error during message reading:", err)
			break
		}
		log.Printf("Received: %s", message)
		var msg wsMessageReceived
		err = json.Unmarshal(message, &msg)
		if err != nil {
			log.Println("Error during message decoding:", err)
			break
		}
		if string(msg.Cmd) == "query" {
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			c := app.Oauth2.Client(ctx, &session.Token)

			var salesforceQuery = "/services/data/v52.0/query?q=" + strings.Replace(msg.Query, " ", "+", -1)
			result, _ := downloadRecords(c, session.Instance, salesforceQuery, conn, messageType)

			writeAnyJSON(msg.Filename, result)
		}
		// if string(msg.Cmd) == "upload" {
		// 	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		// 	defer cancel()

		// 	c := app.Oauth2.Client(ctx, &session.Token)

		// 	var salesforceQuery = "/services/data/v52.0/query?q=" + strings.Replace(msg.Query, " ", "+", -1)
		// 	result, _ := downloadRecords(c, session.Instance, salesforceQuery, conn, messageType)

		// 	writeAnyJSON(msg.Filename, result)
		// }
		if string(msg.Cmd) == "update" {
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			c := app.Oauth2.Client(ctx, &session.Token)

			var data []string

			file, err := ioutil.ReadFile("./js/ContentDocumentId.json")
			if err != nil {
				log.Printf("Unable to read file: %s", err)
				os.Exit(1)
			}
			err = json.Unmarshal(file, &data)
			if err != nil {
				log.Printf("Unable to process file: %s", err)
				os.Exit(1)
			}

			var Ids []string
			for _, element := range data {
				Ids = append(Ids, "ContentDocumentId='"+element+"'")
			}

			limit := 50

			var records []map[string]interface{}

			for i := 0; i < len(Ids); i += limit {
				batch := Ids[i:min(i+limit, len(Ids))]

				newQuery := "SELECT Id, ContentDocumentId, LinkedEntityId FROM ContentDocumentLink WHERE " + strings.Join(batch, " OR ")
				newResult, err := downloadRecords(c, session.Instance, "/services/data/v52.0/query?q="+strings.Replace(newQuery, " ", "+", -1), conn, messageType)

				records = append(records, newResult.Records...)

				if err != nil {
					panic(err)
				}
			}

			writeAnyJSON(msg.Filename, records)
		}

		err = conn.WriteMessage(messageType, message)
		if err != nil {
			log.Println("Error during message writing:", err)
			break
		}
	}
}

func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}
