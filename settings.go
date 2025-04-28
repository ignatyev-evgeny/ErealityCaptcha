package main

import (
	"encoding/json"
	"io/ioutil"
)

type Settings struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	ChatID   string `json:"chat_id"`
}

func saveSettings(settings Settings) error {
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile("settings.json", data, 0644)
}

func loadSettings() (Settings, error) {
	var settings Settings

	data, err := ioutil.ReadFile("settings.json")
	if err != nil {
		return settings, err
	}

	err = json.Unmarshal(data, &settings)
	return settings, err
}
