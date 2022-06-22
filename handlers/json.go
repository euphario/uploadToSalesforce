package handlers

import (
	"encoding/json"
	"os"
)

func writeAnyJSON(filename string, data interface{}) error {
	file, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		panic(err)
	}

	err = os.WriteFile(filename+".json", file, 0644)
	if err != nil {
		panic(err)
	}

	return nil
}
