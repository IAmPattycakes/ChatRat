package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
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
}

func main() {
	var rat ChatRat
	oauth := flag.String("oauth", "", "The oauth code for the twitch bot")
	streamName := flag.String("stream", "", "The name of the stream to join")
	botName := flag.String("botname", "", "The name of the bot")
	chatLog := flag.String("chatlog", "chat.log", "The name of the chat log to use. chat.log is used as the default.")
	trustFile := flag.String("trustfile", "trust.list", "The name of the list of trusted users")
	ignoreFile := flag.String("ignorefile", "block.list", "THe name of the list of ignored users")

	flag.Parse()
	rat.oauth = *oauth
	rat.streamName = *streamName
	rat.botName = *botName
	rat.chatLog = *chatLog
	rat.trustedUserFile = *trustFile
	rat.ignoredUserFile = *ignoreFile

	fmt.Println(rat)

	// or client := twitch.NewAnonymousClient() for an anonymous user (no write capabilities)
	fmt.Println(rat.botName + " " + rat.oauth + " " + rat.streamName)
	client := twitch.NewClient(rat.botName, rat.oauth)
	rat.client = client
	client.OnPrivateMessage(func(message twitch.PrivateMessage) {
		if message.User.Name != "chatrat_" {
			rat.messageParser(message)
		}
	})
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

func (rat *ChatRat) messageParser(message twitch.PrivateMessage) {
	messageStrings := strings.Split(message.Message, " ")
	log.Println(strconv.FormatBool(len(messageStrings) > 0) + ", " + strconv.FormatBool(messageStrings[0] == "!chatrat") + ", " + message.User.Name)
	fmt.Println(messageStrings)
	if (len(messageStrings) > 0) && (messageStrings[0] == "!chatrat") && rat.isUserTrusted(message.User.Name) { //Starting a chatrat command
		if len(messageStrings) > 1 {
			switch messageStrings[1] {
			case "set": //Setting ChatRat variables
				if len(messageStrings) > 2 {
					switch messageStrings[2] {
					case "delay": //Setting the delay between messages
						if len(messageStrings) > 4 {
							if s, err := strconv.ParseFloat(messageStrings[3], 32); err == nil {
								if s < 0 {
									rat.client.Say("iampattycakes", "@"+message.User.Name+" I don't understand how a delay can be negative.")
									return
								}
								switch messageStrings[4] {
								case "seconds", "Seconds", "second", "Second":
									_, err := time.ParseDuration(messageStrings[3] + "s")
									if err == nil {
										rat.client.Say("iampattycakes", "@"+message.User.Name+" I will set the delay to "+messageStrings[3]+"s"+" when I get smart enough")
									} else {
										log.Println(err)
										rat.client.Say("iampattycakes", "@"+message.User.Name+" I don't know what went wrong here. Please screenshot what you said and send to IAmPattycakes on the discord.")
									}
								case "minutes", "Minutes", "minute", "Minute":
									timeToParse := messageStrings[3] + "m"
									_, err := time.ParseDuration(timeToParse)
									if err == nil {
										rat.client.Say("iampattycakes", "@"+message.User.Name+" I will set the delay to "+messageStrings[3]+"m"+" when I get smart enough")
									} else {
										rat.client.Say("iampattycakes", "@"+message.User.Name+" I don't know what went wrong here. Please screenshot what you said and send to IAmPattycakes on the discord.")
									}
								}
							} else if err != nil {
								rat.client.Say("iampattycakes", "@"+message.User.Name+" I see you're trying to set the delay, but you gave me a weird number. ChatRat doesn't know math very well.")
							}
						} else {
							rat.client.Say("iampattycakes", "@"+message.User.Name+" I didn't hear any delay from you. I need a number and either minutes or seconds, like \"3 minutes\" or \"10 seconds\"")
						}
					}
				} else {
					rat.client.Say("iampattycakes", "@"+message.User.Name+" I couldn't understand you, I only saw you say \"!chatrat set\" without anything else.")
					return
				}
			default:
				rat.client.Say("iampattycakes", "@"+message.User.Name+" I couldn't understand you, I only saw you say \"!chatrat\" before I got confused.")
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

func (rat *ChatRat) speechHandler() {
	time.Sleep(10 * time.Second) //Wait for some connections
	for {
		rat.client.Say("iampattycakes", rat.graph.GenerateMarkovString())
		time.Sleep(2 * time.Minute)
	}
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
