package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"
)

type settings struct {
	oauth          string   `json:"oauth"` //Yeah I know oauth stuff probably shouldn't be sitting in a file naked.
	botName        string   `json:"botName"`
	streamName     string   `json:"streamName"`
	trustedUsers   []string `json:"trustedUsers"`
	ignoredUsers   []string `json:"ignoredUsers"`
	commandStarter string   `json:"commandStarter"`
	chatLog        string   `json:"chatLog"`

	emotesToSpam       []string `json:"emotesToSpam"`
	emoteSpamThreshold int      `json:"emoteSpamThreshold"`
	emoteSpamTimeout   string   `json:"emoteSpamTimeout"`
	emoteSpamCooldown  string   `json:"emoteSpamCooldown"`
	emoteSentTime      [][]time.Time
	emoteCleanupMutex  sync.Mutex

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
		log.Println("Error loading settings: " + err.Error())
	}

	byteValue, _ := ioutil.ReadAll(jsonfile)
	json.Unmarshal(byteValue, s)
	s.settingsFileName = filename
}

func (s *settings) saveSettings() {
	f, err := os.OpenFile(s.settingsFileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatal(err)
	}
	str, err := json.Marshal(s)
	if err != nil {
		log.Fatal(err)
	}
	f.Write(str)
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
}
