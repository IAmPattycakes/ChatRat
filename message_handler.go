package main

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/gempir/go-twitch-irc/v2"
)

func (rat *ChatRat) messageParser(message twitch.PrivateMessage) {
	messageStrings := strings.Split(message.Message, " ")

	for i, v := range rat.ratSettings.EmotesToSpam {
		if contains(messageStrings, v) {
			rat.timerCleaner(i)
			rat.emoteTimers[i] = append(rat.emoteTimers[i], time.Now())
			if (len(rat.emoteTimers[i]) >= rat.ratSettings.EmoteSpamThreshold) && rat.emoteLastTime[i].Add(rat.emoteSpamCooldown).Before(time.Now()) {
				rat.speak(v)
				if rat.ratSettings.VerboseLogging {
					log.Println("Triggered " + v + " emote spam")
				}
				rat.emoteLastTime[i] = time.Now()
			}
		}
	}

	if rat.isUserIgnored(message.User.Name) && !strings.EqualFold(message.User.Name, rat.ratSettings.StreamName) {
		return
	}

	messageLength := len(messageStrings)
	//Checks if the message is a command. If it's not, then just listen and add it to what can be said if it should be.
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

	if !rat.isUserTrusted(message.User.Name) && !strings.EqualFold(message.User.Name, rat.ratSettings.StreamName) {
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
			dur, err := durationParse(messageStrings[2:])
			if err != nil {
				rat.speak(err.Error())
			}
			rat.chatDelay.mu.RLock()
			log.Printf("chat delay duration updating from [%s] to [%s]", rat.chatDelay.duration, dur)
			rat.chatDelay.ticker.Stop()
			rat.chatDelay.duration = dur
			rat.chatDelay.ticker.Reset(dur)
			rat.chatDelay.mu.RUnlock()
			rat.ratSettings.ChatDelay = dur.String()
			rat.ratSettings.saveSettings()
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
			rat.ratSettings.ChatContextDepth = int(num)
			rat.ratSettings.saveSettings()
		case "emoteSpamThreshold":
			if messageLength <= 3 {
				rat.speak(fmt.Sprintf("Current emote spam threshold is %d", rat.ratSettings.ChatContextDepth))
				return
			}
			num, err := strconv.ParseInt(messageStrings[3], 10, 0)
			if err != nil {
				log.Println("Couldn't read the emote spam threshold given. command was \"" + message.Message + "\" error given: " + err.Error())
				return
			}
			if num < 0 {
				rat.speak("I will respond after every emote sent I guess")
			}
			rat.ratSettings.EmoteSpamThreshold = int(num)
			rat.speak(fmt.Sprintf("I'll join in saying emotes I know after %d people have said them in %s", num, rat.ratSettings.emoteSpamTimeout.String()))
			rat.ratSettings.saveSettings()
		case "emoteSpamTimeout":
			dur, err := durationParse(messageStrings[2:])
			if err != nil {
				rat.speak(err.Error())
			}
			rat.speak(fmt.Sprintf("emote spam timeout changed from [%s] to [%s]", rat.ratSettings.emoteSpamTimeout, dur.String()))
			rat.ratSettings.emoteSpamTimeout = dur
			rat.ratSettings.EmoteSpamTimeout = dur.String()
			rat.ratSettings.saveSettings()
		case "emoteSpamCooldown":
			dur, err := durationParse(messageStrings[2:])
			if err != nil {
				rat.speak(err.Error())
			}
			rat.speak(fmt.Sprintf("emote spam cooldown changed from [%s] to [%s]", rat.ratSettings.emoteSpamTimeout, dur.String()))
			rat.ratSettings.emoteSpamCooldown = dur
			rat.ratSettings.EmoteSpamCooldown = dur.String()
			rat.ratSettings.saveSettings()
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
	case "spam":
		if len(messageStrings) > 2 {
			if !contains(rat.ratSettings.EmotesToSpam, messageStrings[2]) {
				rat.ratSettings.EmotesToSpam = append(rat.ratSettings.EmotesToSpam, messageStrings[2])
				rat.emoteTimers = append(rat.emoteTimers, make([]time.Time, 0))
				rat.emoteLastTime = append(rat.emoteLastTime, time.Now())
			}
		}
		rat.ratSettings.saveSettings()
		rat.speak("I will now spam " + messageStrings[2] + " with the crowd.")
		log.Println(message.User.Name + " requested spamming of " + messageStrings[2] + " with the crowd")
	case "dontspam", "dontSpam", "stopSpamming", "stopspamming":
		emotestospam := make([]string, 0)
		emotetimers := make([][]time.Time, 0)
		emotelasttime := make([]time.Time, 0)
		if len(messageStrings) > 2 {
			for i, v := range rat.ratSettings.EmotesToSpam {
				if v != messageStrings[2] {
					emotestospam = append(emotestospam, v)
					emotetimers = append(emotetimers, rat.emoteTimers[i])
					emotelasttime = append(emotelasttime, rat.emoteLastTime[i])
				}
			}
		}
		rat.emoteTimers = emotetimers
		rat.emoteLastTime = emotelasttime
		rat.ratSettings.EmotesToSpam = emotestospam
		rat.ratSettings.saveSettings()
		rat.speak("Okay, I'll stop spamming " + messageStrings[2] + " with the crowd.")
		log.Println(message.User.Name + " requested stopping spamming of " + messageStrings[2])
	default:
		rat.speak("@" + message.User.Name + " I couldn't understand you, I only saw you say \"" + rat.ratSettings.CommandStarter + "\" before I got confused.")
	}
}

func (rat *ChatRat) timerCleaner(index int) {
	arr := make([]time.Time, 0)
	for _, v := range rat.emoteTimers[index] {
		if v.Add(rat.emoteTimeout).After(time.Now()) {
			arr = append(arr, v)
		}
	}
	rat.emoteTimers[index] = arr
}

func durationParse(message []string) (time.Duration, error) {
	messageLength := len(message)
	if messageLength == 0 {
		return 9999 * time.Hour, errors.New("no input to duration parser")
	}
	if messageLength == 1 {
		dur, err := time.ParseDuration(message[0])
		if err != nil {
			return 9999 * time.Hour, errors.New("improperly formatted duration string")
		}
		return dur, nil
	}
	//Message length must be 2 or greater now.
	s, err := strconv.ParseFloat(message[1], 32)
	if err != nil {
		return 9999 * time.Hour, errors.New("couldn't parse " + message[1] + " as a number")
	}

	if s < 0 {
		return 9999 * time.Hour, errors.New("negative durations don't make sense in this context")
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
	timeExtension, err := parseTimeExtension(message[1])
	if err != nil {
		return 9999 * time.Hour, errors.New(message[1] + " is not a valid unit of time")
	}

	dur, err := time.ParseDuration(message[0] + timeExtension)
	return dur, err
}
