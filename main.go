package main

import (
	"bufio"
	"flag"
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

	// rat timer settings
	rat.chatDelay.mu.RLock()
	rat.chatDelay.duration = rat.ratSettings.chatDelay
	rat.chatDelay.ticker = time.NewTicker(rat.chatDelay.duration)
	rat.chatDelay.paused = false
	rat.chatDelay.mu.RUnlock()

	rat.emoteTimeout = 10 * time.Second
	rat.emoteSpamCooldown = 1 * time.Minute
	rat.emoteTimers = make([][]time.Time, len(rat.ratSettings.EmotesToSpam))
	rat.emoteLastTime = make([]time.Time, len(rat.ratSettings.EmotesToSpam))

	client := twitch.NewClient(rat.ratSettings.BotName, rat.ratSettings.Oauth)
	rat.client = client
	rat.client.OnPrivateMessage(func(message twitch.PrivateMessage) {
		if message.User.Name != "chatrat_" {
			rat.messageParser(message)
		}
	})
	//Loading the chat history to give the model something to go off of at the start.
	rat.loadChatLog()

	client.Join(rat.ratSettings.StreamName)
	defer client.Disconnect()
	defer client.Depart(rat.ratSettings.StreamName)
	rat.speak("Hi chat I'm back! =^.^=")
	if rat.ratSettings.VerboseLogging {
		log.Println("Chatrat starting in stream " + rat.ratSettings.StreamName + " running as " + rat.ratSettings.BotName)
	}
	go rat.speechHandler()
	err := client.Connect()

	if err != nil {
		panic(err)
	}
}

func (rat *ChatRat) speak(message string) {
	rat.client.Say(rat.ratSettings.StreamName, message)
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
		words := rat.graph.GenerateMarkovString()
		rat.speak(words)
		if rat.ratSettings.VerboseLogging {
			log.Println("Saying \"" + words + "\" from the routine speech handler")
		}
	}
}

func (rat *ChatRat) writeText(text string) {
	f, err := os.OpenFile(rat.ratSettings.ChatLog, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := f.Write([]byte(text + "\n")); err != nil {
		log.Fatal(err)
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
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
