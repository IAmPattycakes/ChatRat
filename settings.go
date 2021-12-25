package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

type settings struct {
	Oauth          string   `json:"oauth"` //Yeah I know oauth stuff probably shouldn't be sitting in a file naked.
	BotName        string   `json:"botName"`
	StreamName     string   `json:"streamName"`
	TrustedUsers   []string `json:"trustedUsers"`
	IgnoredUsers   []string `json:"ignoredUsers"`
	CommandStarter string   `json:"commandStarter"`
	ChatLog        string   `json:"chatLog"`
	VerboseLogging bool     `json:"verboseLogging"`

	EmotesToSpam       []string `json:"emotesToSpam"`
	EmoteSpamThreshold int      `json:"emoteSpamThreshold"`
	EmoteSpamTimeout   string   `json:"emoteSpamTimeout"`
	EmoteSpamCooldown  string   `json:"emoteSpamCooldown"`

	settingsFileName string
}

func NewSettings(filename string) *settings {
	var s settings
	s.loadSettings(filename)
	return &s
}

//loadSettings loads the settings from the provided filename string, and puts them into the settings struct.
func (s *settings) loadSettings(filename string) {
	jsonfile, err := os.Open(filename)
	if err != nil {
		log.Fatal("Error loading settings: " + err.Error())
	}
	byteValue, _ := ioutil.ReadAll(jsonfile)
	err2 := json.Unmarshal(byteValue, &s)
	if err2 != nil {
		log.Fatal("Error parsing settings: " + err2.Error())
	}
	s.settingsFileName = filename
}

//saveSettings saves all the JSON parse-able settings into the file listed in settings.settingsFileName
func (s *settings) saveSettings() {
	f, err := os.OpenFile(s.settingsFileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatal(err)
	}
	str, err := json.MarshalIndent(s, "", "\t")
	if err != nil {
		log.Fatal(err)
	}
	f.Write(str)
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
}

func (s *settings) trustUser(username string) {
	s.TrustedUsers = append(s.TrustedUsers, strings.ToLower(username))
	s.saveSettings()
}

func (s *settings) ignoreUser(username string) {
	s.IgnoredUsers = append(s.IgnoredUsers, strings.ToLower(username))
	s.saveSettings()
}

func (s *settings) untrustUser(username string) bool {
	removed := removeStringFromList(username, &s.TrustedUsers)
	s.saveSettings()
	return removed
}

func (s *settings) unignoreUser(username string) bool {
	removed := removeStringFromList(username, &s.IgnoredUsers)
	s.saveSettings()
	return removed
}

//removeUserFromList creates a new array and sets the list to it. Returns whether or not a user was removed.
func removeStringFromList(username string, list *[]string) bool {
	arr := make([]string, 0)
	ret := false
	for _, v := range *list {
		if !(strings.ToLower(username) == v) {
			arr = append(arr, v)
		} else {
			ret = true
		}
	}
	*list = arr
	return ret
}
