package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
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

	catKisses        []time.Time
	catKissTimeout   time.Duration
	catKissThreshold int
	catKissCooldown  time.Duration
	catKissLastTime  time.Time

	heCrazies        []time.Time
	heCrazyTimeout   time.Duration
	heCrazyThreshold int
	heCrazyCooldown  time.Duration
	heCrazyLastTime  time.Time

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

	// rat timer settings
	rat.chatDelay.mu.RLock()
	rat.chatDelay.duration = 2 * time.Minute
	rat.chatDelay.ticker = time.NewTicker(rat.chatDelay.duration)
	rat.chatDelay.paused = false
	rat.chatDelay.mu.RUnlock()
	rat.graph = *markov.NewGraph(rat.ratSettings.ChatContextDepth)

	rat.catKissTimeout = 10 * time.Second
	rat.catKissThreshold = 3
	rat.catKissCooldown = 1 * time.Minute

	rat.heCrazyTimeout = 10 * time.Second
	rat.heCrazyThreshold = 3
	rat.heCrazyCooldown = 1 * time.Minute

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
	// log.Println("saying" + message)
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

func (rat *ChatRat) catKissCleaner() {
	arr := make([]time.Time, 0)
	for _, v := range rat.catKisses {
		if v.Add(rat.catKissTimeout).After(time.Now()) {
			arr = append(arr, v)
		}
	}
	rat.catKisses = arr
}

func (rat *ChatRat) heCrazyCleaner() {
	arr := make([]time.Time, 0)
	for _, v := range rat.heCrazies {
		if v.Add(rat.heCrazyTimeout).After(time.Now()) {
			arr = append(arr, v)
		}
	}
	rat.heCrazies = arr
}

func (rat *ChatRat) messageParser(message twitch.PrivateMessage) {
	messageStrings := strings.Split(message.Message, " ")
	//CatKissies
	if contains(messageStrings, "catKiss") {
		rat.catKissCleaner()
		rat.catKisses = append(rat.catKisses, time.Now())
		if len(rat.catKisses) > rat.catKissThreshold {
			if rat.catKissLastTime.Add(rat.catKissCooldown).Before(time.Now()) {
				rat.speak("catKiss")
				if rat.ratSettings.VerboseLogging {
					log.Println("Triggered catKiss emote spam")
				}
			}
		}
	}
	if contains(messageStrings, "heCrazy") {
		rat.heCrazyCleaner()
		rat.heCrazies = append(rat.heCrazies, time.Now())
		if len(rat.heCrazies) > rat.heCrazyThreshold {
			if rat.heCrazyLastTime.Add(rat.heCrazyCooldown).Before(time.Now()) {
				rat.speak("heCrazy")
				if rat.ratSettings.VerboseLogging {
					log.Println("Triggered heCrazy emote spam")
				}
			}
		}
	}

	if rat.isUserIgnored(message.User.Name) {
		return
	}

	messageLength := len(messageStrings)
	if messageLength <= 0 || messageStrings[0] != rat.ratSettings.CommandStarter {
		rat.writeText(message.Message)
		loaded, badword := rat.LoadPhrase(message.Message)
		if rat.ratSettings.VerboseLogging {
			if loaded {
				log.Println("Heard \"" + message.Message + "\" and added it to the model")
			} else {
				log.Println("Heard \"" + message.Message + "\" and didn't add it to the model because I saw \"" + badword + "\"")
			}
		}
		return
	}

	if messageLength <= 1 {
		return
	}

	if !rat.isUserTrusted(message.User.Name) {
		rat.speak("Hi I'm ChatRat, I only let trusted people tell me what to do, but I guess you can say my name if you like =^.^=")
		return
	}

	switch messageStrings[1] {
	case "delay":
		rat.speak(fmt.Sprintf("@%s the current delay is set to %s", message.User.Name, rat.chatDelay.duration))
	case "set": //Setting ChatRat variables
		if messageLength <= 2 {
			rat.speak("@" + message.User.Name + " I couldn't understand you, I only saw you say \"" + rat.ratSettings.CommandStarter + " set\" without anything else.")
			return
		}

		switch messageStrings[2] {
		case "delay": //Setting the delay between messages
			if messageLength <= 4 {
				rat.speak("@" + message.User.Name + " I didn't hear any delay from you. I need a number and either hours, minutes, or seconds, like \"3 minutes\" or \"10 seconds\"")
				return
			}

			s, err := strconv.ParseFloat(messageStrings[3], 32)
			if err != nil {
				rat.speak("@" + message.User.Name + " I see you're trying to set the delay, but you gave me a weird number. ChatRat doesn't know math very well.")
			}

			if s < 0 {
				rat.speak("@" + message.User.Name + " I don't understand how a delay can be negative.")
				return
			}

			parseTimeExtension := func(message string) (string, error) {
				switch message {
				case "seconds", "Seconds", "second", "Second":
					return "s", nil
				case "minutes", "Minutes", "minute", "Minute":
					return "m", nil
				case "hours", "Hours", "hour", "Hour":
					return "h", nil
				default:
					return "", errors.New("unknown time extension format")
				}
			}

			timeExtension, err := parseTimeExtension(messageStrings[4])
			if err != nil {
				rat.speak("@" + message.User.Name + "I don't understand what unit of time you're speaking about.")
				return
			}

			dur, err := time.ParseDuration(messageStrings[3] + timeExtension)
			if err != nil {
				log.Println(err)
				rat.speak("@" + message.User.Name + " I don't know what went wrong here. Please screenshot what you said and send to the #chatrat channel on the discord.")
				return
			}

			rat.chatDelay.mu.RLock()
			log.Printf("chat delay duration updating from [%s] to [%s]", rat.chatDelay.duration, dur)
			rat.chatDelay.ticker.Stop()
			rat.chatDelay.duration = dur
			rat.chatDelay.ticker.Reset(dur)
			rat.chatDelay.mu.RUnlock()
		case "contextDepth": //Setting the context depth of the markov chain text generation
			if messageLength <= 3 {
				rat.speak(fmt.Sprintf("Current context depth is %d", rat.ratSettings.ChatContextDepth))
				return
			}
			num, err := strconv.ParseInt(messageStrings[3], 10, 0)
			if err != nil {
				log.Println("Couldn't read the context depth given. command was \"" + message.Message + "\" error given: " + err.Error())
				return
			}
			if num < 0 {
				rat.speak("I can't have less than 0 context")
				return
			}
			rat.speak("@" + message.User.Name + " I'm re-learning what to say, this may take a bit...")
			rat.reloadGraph(int(num))
			rat.speak("Okay, I know how to talk again.")
		}

	case "stop":
		rat.chatDelay.mu.RLock()
		rat.chatDelay.ticker.Stop()
		rat.chatDelay.paused = true
		rat.chatDelay.mu.RUnlock()

		rat.speak("Alright, I'll stop talking for now")
		if rat.ratSettings.VerboseLogging {
			log.Println("Chatrat was stopped by " + message.User.DisplayName)
		}
	case "start":
		rat.chatDelay.mu.RLock()
		rat.chatDelay.ticker.Reset(rat.chatDelay.duration)
		rat.chatDelay.paused = false
		rat.chatDelay.mu.RUnlock()

		rat.speak("Yay I get to talk again!")
		if rat.ratSettings.VerboseLogging {
			log.Println("Chatrat was stopped by " + message.User.DisplayName)
		}
	case "ignore":
		if messageLength > 2 {
			rat.speak("Sorry @" + messageStrings[2] + ", I can't talk to you anymore")
			rat.ratSettings.ignoreUser(messageStrings[2])
			if rat.ratSettings.VerboseLogging {
				log.Println(message.User.DisplayName + " ignored " + messageStrings[2])
			}
			return
		}

		rat.speak("@" + message.User.Name + " I didn't see a user to ignore.")
	case "unignore":
		if messageLength <= 2 {
			return
		}
		unignored := rat.ratSettings.unignoreUser(messageStrings[2])
		if unignored {
			rat.speak("Okay, I'll listen to what @" + messageStrings[2] + " has to say again.")
			if rat.ratSettings.VerboseLogging {
				log.Println(message.User.DisplayName + " unignored " + messageStrings[2])
			}
		} else {
			rat.speak("@" + message.User.Name + ", " + messageStrings[2] + " wasn't ignored before.")
		}
	case "trust":
		if messageLength > 2 {
			rat.speak("Okay @" + messageStrings[2] + ", I'll let you tell me things to do")
			rat.ratSettings.trustUser(messageStrings[2])
			if rat.ratSettings.VerboseLogging {
				log.Println(message.User.DisplayName + " trusted " + messageStrings[2])
			}
			return
		}
		rat.speak("@" + message.User.Name + " I didn't see a user to trust.")
	case "untrust":
		if messageLength <= 2 {
			return
		}
		untrusted := rat.ratSettings.untrustUser(messageStrings[2])
		if untrusted {
			rat.speak("Sorry @" + messageStrings[2] + ", I can't listen to commands from you anymore")
			if rat.ratSettings.VerboseLogging {
				log.Println(message.User.DisplayName + " untrusted " + messageStrings[2])
			}
		} else {
			rat.speak("@" + message.User.Name + ", " + messageStrings[2] + " wasn't trusted before.")
		}
		rat.speak("Sorry @" + messageStrings[2] + ", I can't listen to commands from you anymore")
	case "speak":
		words := rat.graph.GenerateMarkovString()
		rat.speak(words)
		if rat.ratSettings.VerboseLogging {
			log.Println("Saying \"" + words + "\" after being told to speak")
		}
	case "reloadBlacklist":
		err := rat.ratSettings.reloadBlacklist()
		if err != nil {
			rat.speak("I couldn't understand the blacklist anymore")
			log.Println(err)
			return
		}
		rat.speak("I'm re-learning what to say while ignoring the new bad words. This may take a bit.")
		rat.reloadGraph(rat.ratSettings.ChatContextDepth)
		rat.speak("Okay, I know how to talk now.")
	default:
		rat.speak("@" + message.User.Name + " I couldn't understand you, I only saw you say \"" + rat.ratSettings.CommandStarter + "\" before I got confused.")
	}
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
