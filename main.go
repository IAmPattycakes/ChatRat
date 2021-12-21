package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	markov "github.com/IAmPattycakes/Go-Markov/v2"
	"github.com/gempir/go-twitch-irc/v2"
)

type ChatRat struct {
	graph           markov.Graph
	client          *twitch.Client
	trustedUsers    []string
	trustedUserFile string
	chatLog         string
	streamName      string
	oauth           string
	commandStarter  string
	botName         string
	ignoredUsers    []string
	ignoredUserFile string
	chatDelay       []string
	chatTrigger     time.Timer

	lastGoodTime time.Duration //The last time that was properly parsed. This shouldn't have to be used, but if the error checking fails for some reason, well it'll keep things running.
}

func main() {
	var rat ChatRat
	oauth := flag.String("oauth", "", "The oauth code for the twitch bot")
	streamName := flag.String("stream", "", "The name of the stream to join")
	botName := flag.String("botname", "", "The name of the bot")
	chatLog := flag.String("chatlog", "chat.log", "The name of the chat log to use. chat.log is used as the default.")
	trustFile := flag.String("trustfile", "trust.list", "The name of the list of trusted users")
	ignoreFile := flag.String("ignorefile", "block.list", "The name of the list of ignored users")
	commandStarter := flag.String("command", "!chatrat", "The word to get the bot's attention for commands")

	flag.Parse()
	rat.oauth = *oauth
	rat.streamName = *streamName
	rat.botName = *botName
	rat.chatLog = *chatLog
	rat.trustedUserFile = *trustFile
	rat.ignoredUserFile = *ignoreFile
	rat.commandStarter = *commandStarter

	// or client := twitch.NewAnonymousClient() for an anonymous user (no write capabilities)
	fmt.Println(rat.botName + " " + rat.oauth + " " + rat.streamName)
	client := twitch.NewClient(rat.botName, rat.oauth)
	rat.client = client
	client.OnPrivateMessage(func(message twitch.PrivateMessage) {
		if message.User.Name != "chatrat_" {
			rat.messageParser(message)
		}
	})
	//Loading the chat history to give the model something to go off of at the start.
	rat.loadChatLog()
	//Setting up the stuff for special users
	loadUserList(rat.trustedUserFile, &rat.trustedUsers)
	loadUserList(rat.ignoredUserFile, &rat.ignoredUsers)

	client.Join(rat.streamName)
	client.Say(rat.streamName, "ChatRat: online")
	go rat.speechHandler()
	err := client.Connect()

	if err != nil {
		panic(err)
	}
}

func loadUserList(filename string, list *[]string) {
	file, err := os.Open(filename)
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
		*list = append(*list, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func (rat *ChatRat) speak(message string) {
	rat.client.Say(rat.streamName, message)
}

func (rat *ChatRat) messageParser(message twitch.PrivateMessage) {
	messageStrings := strings.Split(message.Message, " ")
	log.Println(strconv.FormatBool(len(messageStrings) > 0) + ", " + strconv.FormatBool(messageStrings[0] == rat.commandStarter) + ", " + message.User.Name)
	fmt.Println(messageStrings)
	if (len(messageStrings) > 0) && (messageStrings[0] == rat.commandStarter) && rat.isUserTrusted(message.User.Name) { //Starting a chatrat command
		if len(messageStrings) > 1 {
			switch messageStrings[1] {
			case "set": //Setting ChatRat variables
				if len(messageStrings) > 2 {
					switch messageStrings[2] {
					case "delay": //Setting the delay between messages
						if len(messageStrings) > 4 {
							if s, err := strconv.ParseFloat(messageStrings[3], 32); err == nil {
								if s < 0 {
									rat.speak("@" + message.User.Name + " I don't understand how a delay can be negative.")
									return
								}
								var timeExtension string
								switch messageStrings[4] {
								case "seconds", "Seconds", "second", "Second":
									timeExtension = "s"
								case "minutes", "Minutes", "minute", "Minute":
									timeExtension = "m"
								case "hours", "Hours", "hour", "Hour":
									timeExtension = "h"
								default:
									rat.speak("@" + message.User.Name + "I don't understand what unit of time you're speaking about.")
								}
								_, err := time.ParseDuration(messageStrings[3] + timeExtension)
								if err == nil {
									rat.speak("@" + message.User.Name + " I will set the delay to " + messageStrings[3] + timeExtension + " when I get smart enough")
								} else {
									log.Println(err)
									rat.speak("@" + message.User.Name + " I don't know what went wrong here. Please screenshot what you said and send to IAmPattycakes on the discord.")
								}
							} else if err != nil {
								rat.speak("@" + message.User.Name + " I see you're trying to set the delay, but you gave me a weird number. ChatRat doesn't know math very well.")
							}
						} else {
							rat.speak("@" + message.User.Name + " I didn't hear any delay from you. I need a number and either hours, minutes, or seconds, like \"3 minutes\" or \"10 seconds\"")
						}
					}
				} else {
					rat.speak("@" + message.User.Name + " I couldn't understand you, I only saw you say \"" + rat.commandStarter + " set\" without anything else.")
					return
				}
			default:
				rat.speak("@" + message.User.Name + " I couldn't understand you, I only saw you say \"" + rat.commandStarter + "\" before I got confused.")
				return
			}
		}
	} else {
		rat.writeText(message.Message)
		rat.graph.LoadPhrase(message.Message)
	}
}

func (rat *ChatRat) isUserTrusted(username string) bool {
	for _, u := range rat.trustedUsers {
		if username == u {
			return true
		}
	}
	return false
}

func (rat *ChatRat) speechDelayPicker() time.Duration {
	switch len(rat.chatDelay) {
	case 1:
		t, err := time.ParseDuration(rat.chatDelay[0])
		if err == nil {
			rat.lastGoodTime = t
			return t
		} else {
			log.Println("Error parsing time: " + rat.chatDelay[0])
			return rat.lastGoodTime
		}
	case 2:
		t1, err := time.ParseDuration(rat.chatDelay[0])
		if err != nil {
			log.Println("Error parsing time: " + rat.chatDelay[0])
			return rat.lastGoodTime
		}
		t2, err := time.ParseDuration(rat.chatDelay[1])
		if err != nil {
			log.Println("Error parsing time: " + rat.chatDelay[1])
			return rat.lastGoodTime
		}
		if t1 > t2 {
			t1, t2 = t2, t1 //Swap the times to make the time randomization math work nicely without having to duplicate a bunch of crap.
		}
		return time.Duration(rand.Int63n(int64(t2-t1/time.Millisecond))) * time.Millisecond
	case 0:
		log.Println("I don't have a proper delay set up")
		return 5 * time.Minute
	}
	log.Print("The chatDelay array seems to have a bad amount of inputs, here it is")
	log.Println(rat.chatDelay)
	return 5 * time.Minute
}

func (rat *ChatRat) speechHandler() {
	rat.speak(rat.graph.GenerateMarkovString())
	rat.chatTrigger.Reset(rat.speechDelayPicker())
	<-rat.chatTrigger.C
}

func (rat *ChatRat) writeText(text string) {
	f, err := os.OpenFile(rat.chatLog, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
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
	file, err := os.Open(rat.chatLog)
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
		rat.graph.LoadPhrase(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}
