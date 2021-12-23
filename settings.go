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
}
