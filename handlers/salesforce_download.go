package handlers

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dpaks/goworkers"
	"github.com/gorilla/websocket"
)

func downloadRecords(c *http.Client, instance string, url string, conn *websocket.Conn, messageType int) (*Records, error) {
	log.Printf("Downloading: %s...", instance+url)
	// r := fmt.Sprintf("downloaded %s records\n", instance+url)

	// conn.WriteMessage(messageType, []byte(r))

	response, err := c.Get(instance + url)
	if err != nil {
		log.Printf("Error getting response: %s", err)
		return nil, err
	}

	if response.StatusCode != http.StatusOK {

		log.Printf("StatusCode is %d", response.StatusCode)
		log.Printf("%s", response.Status)
		log.Printf("%v", response.Header)

		time.Sleep(5 * time.Second)

		response, err = c.Get(instance + url)
		if err != nil {
			log.Printf("Error2 getting response: %s", err)
			return nil, err
		}
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	s := Records{}

	// decode JSON structure into Go structure
	err = json.Unmarshal([]byte(body), &s)
	if err != nil {
		log.Printf("body: %s", body)
		log.Printf("Error decoding %s", err)
		return nil, err
	}

	msg := wsMessageSend{TotalSize: s.TotalSize, Records: len(s.Records), Done: s.Done}

	msgJson, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error preparing message %s", err)
		return nil, err
	}

	conn.WriteMessage(messageType, msgJson)

	if s.Done {
		return &s, nil
	}

	nextBody, err := downloadRecords(c, instance, s.NextRecordsURL, conn, messageType)
	if err != nil {
		log.Printf("Error received %s", err)
		return nil, err
	}
	s.Records = append(s.Records, nextBody.Records...)

	return &s, nil
}

func downloadRecords2(c *http.Client, instance string, url string) (*Records, error) {
	log.Printf("Downloading: %s...", instance+url)
	// r := fmt.Sprintf("downloaded %s records\n", instance+url)

	// conn.WriteMessage(messageType, []byte(r))

	response, err := c.Get(instance + url)
	if err != nil {
		log.Printf("Error getting response: %s", err)
		return nil, err
	}

	if response.StatusCode != http.StatusOK {

		log.Printf("StatusCode is %d", response.StatusCode)
		log.Printf("%s", response.Status)
		log.Printf("%v", response.Header)

		time.Sleep(5 * time.Second)

		response, err = c.Get(instance + url)
		if err != nil {
			log.Printf("Error2 getting response: %s", err)
			return nil, err
		}
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	s := Records{}

	// decode JSON structure into Go structure
	err = json.Unmarshal([]byte(body), &s)
	if err != nil {
		log.Printf("body: %s", body)
		log.Printf("Error decoding %s", err)
		return nil, err
	}

	if s.Done {
		return &s, nil
	}

	nextBody, err := downloadRecords2(c, instance, s.NextRecordsURL)
	if err != nil {
		log.Printf("Error received %s", err)
		return nil, err
	}
	s.Records = append(s.Records, nextBody.Records...)

	return &s, nil
}

func hash_file_md5(filePath string) (string, error) {
	var returnMD5String string
	file, err := os.Open(filePath)
	if err != nil {
		return returnMD5String, err
	}
	defer file.Close()
	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return returnMD5String, err
	}
	hashInBytes := hash.Sum(nil)[:16]
	returnMD5String = hex.EncodeToString(hashInBytes)
	return returnMD5String, nil

}

func (app *App) Downloads(w http.ResponseWriter, r *http.Request) {
	type Download struct {
		Url               string `json:"Url"`
		Id                string `json:"Id"`
		Checksum          string `json:"Checksum"`
		ContentDocumentId string `json:"ContentDocumentId"`
	}
	files := []string{"./js/filesNotFound_20210910.json"}

	// wp := workpool.New(30) // Set the maximum number of threads

	fn := func(id string, contentDocumentId string, checksum string) error {
		hash, _ := hash_file_md5("files/" + contentDocumentId)
		if hash != "" && hash != checksum {
			log.Printf("hash is different from checksum, %s", id)
		}
		if hash == checksum {
			log.Printf("hash is ok, %s", id)
			return nil
		}

		log.Printf("Downloading: %v", id+"/VersionData")
		body, err := app.get(w, r, id+"/VersionData")
		if err != nil {
			return err
		}
		if err := os.WriteFile("files/"+contentDocumentId, *body, 0666); err != nil {
			return err
		}
		return nil
	}

	for _, file := range files {
		log.Printf("file %s", file)

		opts := goworkers.Options{Workers: 20}
		gw := goworkers.New(opts)

		// go func() {
		// 	// Error channel provides errors from job, if any
		// 	for err := range gw.ErrChan {
		// 		fmt.Println(err)
		// 	}
		// }()

		data, err := ioutil.ReadFile(file)
		if err != nil {
			fmt.Print(err)
		}

		var obj []Download

		err = json.Unmarshal(data, &obj)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		for _, v := range obj {
			url := v
			gw.Submit(func() {
				fn(url.Url, url.ContentDocumentId, url.Checksum)
			})
		}
		gw.Stop(false)
		log.Printf("Done")
	}
	log.Printf("Finished")
}

func (app *App) QueryDocumentLink(w http.ResponseWriter, r *http.Request) {
	type Download struct {
		Url               string `json:"Url"`
		Id                string `json:"Id"`
		Checksum          string `json:"Checksum"`
		ContentDocumentId string `json:"ContentDocumentId"`
	}
	files := []string{"./js/filesNotFound_20210910.json"}

	// wp := workpool.New(30) // Set the maximum number of threads
	session, err := getSession(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fn := func(contentDocumentId string) {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		c := app.Oauth2.Client(ctx, &session.Token)

		var salesforceQuery = "/services/data/v52.0/query?q=" + strings.Replace(fmt.Sprintf("SELECT ContentDocumentId, LinkedEntityId FROM ContentDocumentLink WHERE ContentDocumentId='%s'", contentDocumentId), " ", "+", -1)
		result, _ := downloadRecords2(c, session.Instance, salesforceQuery)

		writeAnyJSON(contentDocumentId, result)
	}

	for _, file := range files {
		log.Printf("file %s", file)

		opts := goworkers.Options{Workers: 20}
		gw := goworkers.New(opts)

		// go func() {
		// 	// Error channel provides errors from job, if any
		// 	for err := range gw.ErrChan {
		// 		fmt.Println(err)
		// 	}
		// }()

		data, err := ioutil.ReadFile(file)
		if err != nil {
			fmt.Print(err)
		}

		var obj []Download

		err = json.Unmarshal(data, &obj)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		for _, v := range obj {
			url := v
			gw.Submit(func() {
				fn(url.ContentDocumentId)
			})
		}
		gw.Stop(false)
		log.Printf("Done")
	}
	log.Printf("Finished")
}
