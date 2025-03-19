package parser

import (
	"daespuor91/choose-your-adventure/internal/model"
	"encoding/json"
	"log"
	"os"
)

func ParseJSON() (model.Story, error) {
	// read JSON

	content, err := os.ReadFile("./internal/data/gopher.json")

	if err != nil {
		log.Printf("Error occurr reading the json file! %v", err)
		return nil, err
	}

	// parse JSON
	var stories model.Story

	err = json.Unmarshal(content, &stories)

	if err != nil {
		log.Printf("Error parsing the JSON file! %v", err)
		return nil, err
	}

	return stories, nil
}
