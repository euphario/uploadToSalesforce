package handlers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/dpaks/goworkers"
)

func (app *App) LogResult(file *os.File, encoder *json.Encoder, r Result) {
	if *app.OutputFormat == "json" {
		err := encoder.Encode(r)
		if err != nil {
			log.Fatalf("Unable to encode %s", err)
		}
		fmt.Fprintf(file, ",\n")
		return
	}
	fmt.Fprintf(file, "\"%s\",\"%s\",\"%s\",\"%s\",\"%s\"\n", r.OldContentDocumentId, r.ContentDocumentId, r.LinkedEntityId, r.FilePath, r.Error)
}

func (app *App) Uploads(w http.ResponseWriter, r *http.Request) {
	wg := new(sync.WaitGroup)

	if _, err := os.Stat(*app.Output); err == nil {
		os.Remove(*app.Output)
	}
	output, err := os.OpenFile(*app.Output, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		log.Fatalf("Unable to open output file %s", err)
	}
	var encoder *json.Encoder

	if *app.OutputFormat == "json" {
		fmt.Fprintf(output, "[\n")
		defer func() {
			fmt.Fprintf(output, "]")
			output.Close()
		}()
		encoder = json.NewEncoder(output)
	} else {
		fmt.Fprintf(output, "\"OldContentDocumentId\",\"ContentDocumentId\",\"LinkedEntityId\",\"File\",\"Error\"\n")
		defer output.Close()
	}

	opts := goworkers.Options{Workers: uint32(*app.Concurrency)}
	gw := goworkers.New(opts)

	go func() {
		for {
			select {
			// Error channel provides errors from job, if any
			case err, ok := <-gw.ErrChan:
				// The error channel is closed when the workers are done with their tasks.
				// When the channel is closed, ok is set to false
				if !ok {
					return
				}
				fmt.Printf("Error: %s\n", err.Error())
			// Result channel provides output from job, if any
			// It will be of type interface{}
			case res, ok := <-gw.ResultChan:
				// The result channel is closed when the workers are done with their tasks.
				// When the channel is closed, ok is set to false
				if !ok {
					return
				}

				app.LogResult(output, encoder, res.(Result))
				log.Printf("Done: %s\n", res.(Result).ContentDocumentId)
				wg.Done()
			}
		}
	}()

	fn := func(filePath string, title string, pathOnClient string, contentDocumentId string, description string, linkedEntityIds string) (Result, error) {
		wg.Add(1)
		result := Result{OldContentDocumentId: contentDocumentId, FilePath: pathOnClient}

		log.Printf("FilePath: %s\n", filePath)

		file, err := ioutil.ReadFile(filePath)
		if err != nil {
			result.Error = err.Error()
			return result, nil
		}

		post := Upload{Title: title, PathOnClient: pathOnClient, Description: description, VersionData: base64.StdEncoding.EncodeToString(file)}

		postString, err := json.Marshal(post)
		if err != nil {
			result.Error = err.Error()
			return result, nil
		}

		body, err := app.put(w, r, "/services/data/v52.0/sobjects/ContentVersion", string(postString))
		if err != nil {
			result.Error = fmt.Sprintf("%s %s", err.Error(), body)
			return result, nil
		}

		var bodyResponse UploadResponse
		err = json.Unmarshal(*body, &bodyResponse)
		if err != nil {
			result.Error = fmt.Sprintf("%s %s", err.Error(), body)
			return result, nil
		}

		body, err = app.get(w, r, "/services/data/v52.0/sobjects/ContentVersion/"+bodyResponse.Id)
		if err != nil {
			result.Error = fmt.Sprintf("%s %s", err.Error(), body)
			return result, nil
		}

		var bodyResponse2 ContentVersionResponse
		err = json.Unmarshal(*body, &bodyResponse2)
		if err != nil {
			result.Error = err.Error()
			return result, nil
		}

		result.ContentDocumentId = bodyResponse2.ContentDocumentId

		var doneLinks []string

		for _, linkedEntityId := range strings.Split(linkedEntityIds, ",") {
			var post2 = ContentDocumentLink{ContentDocumentId: bodyResponse2.ContentDocumentId, LinkedEntityId: linkedEntityId}

			postString, err = json.Marshal(post2)
			if err != nil {
				result.Error = err.Error()
				return result, nil
			}

			body, err = app.put(w, r, "/services/data/v52.0/sobjects/ContentDocumentLink", string(postString))
			if err != nil {
				result.Error = fmt.Sprintf("%s %s", err.Error(), body)
				// app.LogResult(output, encoder, &result)
				return result, nil
			}

			json.Unmarshal(*body, &bodyResponse)
			log.Printf("%s\n", body)
			doneLinks = append(doneLinks, linkedEntityId)
			result.LinkedEntityId = strings.Join(doneLinks, ",")
		}

		return result, nil
	}

	data, err := ioutil.ReadFile(*app.FileUpload)
	if err != nil {
		fmt.Print(err)
	}

	var obj []Download

	err = json.Unmarshal(data, &obj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	for _, v := range obj {
		thisFile := v
		gw.SubmitCheckResult(func() (interface{}, error) {
			log.Printf("FilePath: %s, Title: %s, ContentDocumentId: %s, Description: %s", thisFile.FilePath, thisFile.Title, thisFile.ContentDocumentId, thisFile.Description)
			r, _ := fn(thisFile.FilePath, thisFile.Title, thisFile.PathOnClient, thisFile.ContentDocumentId, thisFile.Description, thisFile.LinkedEntityId)
			return r, nil
		})
	}

	gw.Stop(true)
	wg.Wait()
	log.Printf("Finished")
}
