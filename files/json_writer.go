package files

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

func ReadJsonFile(path string) map[string]interface{} {
	// Open the JSON file
	file, err := os.Open(path)
	if err != nil {
		log.Fatalln("error opening ", path, err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatalln("error closing ", path, err)
		}
	}(file)

	// Read the contents of the file into a byte array
	data, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatalln("error reading ", path, err)
	}

	// Unmarshal the JSON data into a map[string]interface{}
	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		log.Fatalln("error unmarshalling ", path, err)
	}

	// Print the data
	return result
}

func WriteToJsonFile(path string, data map[string]interface{}) {
	// Convert the map to a JSON string
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Fatalln("error marshalling ", path, err)
	}

	// Write the JSON string to a file
	err = ioutil.WriteFile(path, jsonData, 0644)
	if err != nil {
		log.Fatalln("error writing ", path, err)
	}

	fmt.Println("Data written to file successfully!")
}
