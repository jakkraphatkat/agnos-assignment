package main

import (
	"encoding/json"
	"log"
)

func logJSON(level, message string, fields map[string]any) {
	payload := map[string]any{
		"level":   level,
		"message": message,
	}

	for key, value := range fields {
		payload[key] = value
	}

	line, err := json.Marshal(payload)
	if err != nil {
		log.Printf(`{"level":"error","message":"failed to marshal log","error":%q}`, err.Error())
		return
	}

	log.Println(string(line))
}
