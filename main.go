package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	markov "github.com/IAmPattycakes/Go-Markov"
	"github.com/gempir/go-twitch-irc/v2"
)

type ChatRat struct {
	graph       markov.Graph
	client      *twitch.Client
	ratSettings settings

	emoteTimers       [][]time.Time
	emoteTimeout      time.Duration
	emoteLastTime     []time.Time
	emoteSpamCooldown time.Duration

	chatDelay chatDelay
	logger RatLogger
}

type chatDelay struct {
	mu       sync.RWMutex
	ticker   *time.Ticker
	duration time.Duration
	paused   bool
}

func main() {
	var rat ChatRat

	settingsFile := flag.String("settings", "settings.json", "The name of the settings json file")
	flag.Parse()
	rat.ratSettings = *NewSettings(*settingsFile)
	rat.graph = *markov.NewGraph(rat.ratSettings.ChatContextDepth)
	rat.logger = *NewLogger(rat.ratSettings.logType, rat.ratSettings.LogName, rat.ratSettings.logLevel)
	go rat.logger.HandleLogs()
	defer rat.logger.Close()

	//Timer settings
	rat.chatDelay.mu.RLock()
	rat.chatDelay.duration = rat.ratSettings.chatDelay
	rat.chatDelay.ticker = time.NewTicker(rat.chatDelay.duration)
	rat.chatDelay.paused = false
	rat.log(Debug, fmt.Sprintf("Setting new chat delay to %s", rat.ratSettings.chatDelay.String()))
	rat.chatDelay.mu.RUnlock()

	//Emote spam settings
	rat.emoteTimeout = rat.ratSettings.emoteSpamTimeout
	rat.emoteSpamCooldown = rat.ratSettings.emoteSpamCooldown
	rat.emoteTimers = make([][]time.Time, len(rat.ratSettings.EmotesToSpam))
	rat.emoteLastTime = make([]time.Time, len(rat.ratSettings.EmotesToSpam))
	rat.log(Debug, fmt.Sprintf("Setting emoteTimeout to %s and emoteSpamCooldown to %s", rat.ratSettings.emoteSpamTimeout.String(), rat.ratSettings.emoteSpamCooldown.String()))

	client := twitch.NewClient(rat.ratSettings.BotName, rat.ratSettings.Oauth)
	rat.client = client
	rat.client.OnPrivateMessage(func(message twitch.PrivateMessage) {
		rat.log(Debug, fmt.Sprintf("Passing message to messageParser, raw: %s", message.Raw))
		rat.messageParser(message)
	})
	//Loading the chat history to give the model something to go off of at the start.
	rat.log(Debug, fmt.Sprintf("Starting loading chat log at %s", time.Now().String()))
	rat.loadChatLog()
	rat.log(Debug, fmt.Sprintf("Finished loading chat log at %s", time.Now().String()))

	client.Join(rat.ratSettings.StreamName)
	defer client.Disconnect()
	defer client.Depart(rat.ratSettings.StreamName)
	rat.speak("Hi chat I'm back! =^.^=")
	rat.logger.Log(Info, "Chatrat starting in stream " + rat.ratSettings.StreamName + " running as " + rat.ratSettings.BotName)

	go rat.speechHandler()
	err := client.Connect()

	if err != nil {
		panic(err)
	}
}

//speak checks to see if the message given is able to be said in chat, says it, and returns true if it can. Returns false if it can't.
func (rat *ChatRat) speak(message string) bool {
	if len(message) > 512 {
		rat.log(Debug, fmt.Sprintf("Failed to speak message: %s, too long", message))
		return false
	}
	rat.client.Say(rat.ratSettings.StreamName, message)
	return true
}

func (rat *ChatRat) log(sev LogSeverity, m string) {
	rat.logger.Log(sev, m)
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func (rat *ChatRat) isUserTrusted(username string) bool {
	return contains(rat.ratSettings.TrustedUsers, username)
}

func (rat *ChatRat) isUserIgnored(username string) bool {
	return contains(rat.ratSettings.IgnoredUsers, username)
}

func (rat *ChatRat) speechHandler() {
	for range rat.chatDelay.ticker.C {
		if rat.chatDelay.paused {
			continue
		}
		spoken := false
		for !spoken {
			rat.log(Debug, "Trying to generate speech for routine speech handler")
			words := rat.graph.GenerateMarkovString()
			spoken = rat.speak(words)
			if spoken {
				rat.log(Info, "Saying \"" + words + "\" from the routine speech handler")
			}
		}
	}
}

func (rat *ChatRat) writeText(text string) {
	f, err := os.OpenFile(rat.ratSettings.ChatLog, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		rat.log(Critical, "Couldn't open chat log to write, " + err.Error())
	}
	if _, err := f.Write([]byte(text + "\n")); err != nil {
		rat.log(Critical, "Couldn't write to chat log, " + err.Error())
	}
	if err := f.Close(); err != nil {
		rat.log(Critical, "Couldn't close chat log, " + err.Error())
	}
}

func (rat *ChatRat) loadChatLog() {
	file, err := os.Open(rat.ratSettings.ChatLog)
	quitnow := false
	if err != nil {
		log.Print(err)
		quitnow = true
	}
	defer file.Close()
	if quitnow {
		return
	}
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		rat.LoadPhrase(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func (rat *ChatRat) reloadGraph(depth int) {
	//Pause chatting so it doesn't try to talk with a bad/incomplete graph
	rat.chatDelay.mu.RLock()
	rat.chatDelay.ticker.Stop()
	rat.chatDelay.paused = true

	//Set the context depth in the settings,
	rat.ratSettings.ChatContextDepth = int(depth)
	rat.graph = *markov.NewGraph(depth)
	rat.loadChatLog()

	//Unpause the chatting and release the lock
	rat.chatDelay.ticker.Reset(rat.chatDelay.duration)
	rat.chatDelay.paused = false
	rat.chatDelay.mu.RUnlock()
}

// LoadPhrase attemptes to load the phrase into the markov chain. If the phrase has a blacklisted word/phrase in it
// then the whole phrase gets ignored and the function returns false. If it passes, then returns true.
func (rat *ChatRat) LoadPhrase(s string) (bool, string) {
	for _, v := range rat.ratSettings.blacklist {
		if strings.Contains(strings.ToLower(s), strings.ToLower(v)) {
			return false, v
		}
	}
	rat.graph.LoadPhrase(s)
	return true, ""
}
