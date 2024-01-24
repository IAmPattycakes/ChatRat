package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
)

type settings struct {
	Oauth            string   `json:"oauth"` //Yeah I know oauth stuff probably shouldn't be sitting in a file naked.
	BotName          string   `json:"botName"`
	StreamName       string   `json:"streamName"`
	TrustedUsers     []string `json:"trustedUsers"`
	IgnoredUsers     []string `json:"ignoredUsers"`
	CommandStarter   string   `json:"commandStarter"`
	ChatLog          string   `json:"chatLog"`
	ChatContextDepth int      `json:"chatContextDepth"`
	ChatDelay        string   `json:"chatDelay"`
	chatDelay        time.Duration

	EmotesToSpam       []string `json:"emotesToSpam"`
	EmoteSpamThreshold int      `json:"emoteSpamThreshold"`
	EmoteSpamTimeout   string   `json:"emoteSpamTimeout"`
	EmoteSpamCooldown  string   `json:"emoteSpamCooldown"`
	emoteSpamTimeout   time.Duration
	emoteSpamCooldown  time.Duration

	BlacklistFileName      string `json:"blacklistFileName"`
	blacklist              []string
	RegexBlacklistFileName string `json:"regexBlacklistFileName"`
	regexBlacklist         []string
	settingsFileName       string

	LogType  string `json:"logType"`
	LogLevel string `json:"logLevel"`
	LogName  string `json:"logName"`
	logType  LogType
	logLevel LogSeverity
}

type blacklist struct {
	Blacklist      []string `json:"blacklist"`
	RegexBlacklist []string `json:"regexBlacklist"`
}

func NewSettings(filename string) *settings {
	var s settings
	s.loadSettings(filename)
	return &s
}

// loadSettings loads the settings from the provided filename string, and puts them into the settings struct.
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

	switch strings.ToLower(s.LogType) {
	case "console":
		s.logType = Console
	case "file":
		s.logType = File
	case "both":
		s.logType = File | Console
	default:
		s.logType = File & Console //No logger
	}
	switch strings.ToLower(s.LogLevel) {
	case "debug":
		s.logLevel = Debug
	case "info":
		s.logLevel = Info
	case "warning":
		s.logLevel = Warning
	case "critical":
		s.logLevel = Critical
	default:
		s.logLevel = Info
	}

	//In twitch the usernames are all lowercase in the backend. If the settings file includes names with uppercase characters, turn them lower.
	s.BotName = strings.ToLower(s.BotName)
	s.StreamName = strings.ToLower(s.StreamName)
	for i, v := range s.TrustedUsers {
		s.TrustedUsers[i] = strings.ToLower(v)
	}

	for i, v := range s.IgnoredUsers {
		s.IgnoredUsers[i] = strings.ToLower(v)
	}

	var b blacklist
	blacklistfile, err := os.Open(s.BlacklistFileName)
	if err != nil {
		log.Println("Could not open the blacklist, continuing with none.")
	} else {
		blacklistBytes, _ := ioutil.ReadAll(blacklistfile)
		blacklistParseError := json.Unmarshal(blacklistBytes, &b)
		if blacklistParseError != nil {
			log.Println("Error parsing blacklist. Continuing on with none. " + blacklistParseError.Error())
			s.blacklist = make([]string, 0)
		} else {
			s.blacklist = b.Blacklist
		}
	}
	timeout, err := time.ParseDuration(s.EmoteSpamTimeout)
	if err != nil {
		log.Println("Error parsing emote spam timeout. Continuing with default 10s")
		s.emoteSpamTimeout = 10 * time.Second
	} else {
		s.emoteSpamTimeout = timeout
	}
	cooldown, err := time.ParseDuration(s.EmoteSpamCooldown)
	if err != nil {
		log.Println("Error parsing emote spam cooldown. Continuing with default 1m")
		s.emoteSpamCooldown = 1 * time.Minute
	} else {
		s.emoteSpamCooldown = cooldown
	}
	delay, err := time.ParseDuration(s.ChatDelay)
	if err != nil {
		log.Println("Error parsing chat delay. Continuing with default 2m")
		s.chatDelay = 2 * time.Minute
	} else {
		s.chatDelay = delay
	}
}

// saveSettings saves all the JSON parse-able settings into the file listed in settings.settingsFileName
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

// removeStringFromList creates a new array without the input and sets the list to it. Returns whether or not an element was removed.
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

// reloadBlacklist reloads the blacklist from the file, returns false if it fails.
func (s *settings) reloadBlacklist() error {
	var b blacklist
	blacklistfile, err := os.Open(s.BlacklistFileName)
	if err != nil {
		return errors.New("could not open the blacklist, will not update")
	}
	blacklistBytes, _ := ioutil.ReadAll(blacklistfile)
	blacklistParseError := json.Unmarshal(blacklistBytes, &b)
	if blacklistParseError != nil {
		return errors.New("couldn't parse blacklist, not updating currently existing one. Error: " + blacklistParseError.Error())
	}
	s.blacklist = b.Blacklist
	s.regexBlacklist = b.RegexBlacklist
	if err != nil {
		return errors.New("could not open the blacklist, will not update")
	}
	return nil
}
