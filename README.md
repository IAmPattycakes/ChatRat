# ChatRat
ChatRat is a twitch chat bot built in Go that is a dedicated shitpost machine. Also does some other things, but for now the main thing is just a markov shitpost generator. He tries to act like a normal member of chat, but isn't the brightest so he can kinda go off the rails sometimes. 

## Commands
Commands are permissioned, so only trusted users are allowed to execute them. Currently the commands are: 
- `!chatrat stop` Stops the constant markov text chatting
- `!chatrat start` Starts the constant markov text chatting if it was previously stopped. 
- `!chatrat delay` Shows the current speech delay. 
- `!chatrat set` Sets variables inside of ChatRat. These include: 
    - `delay <duration>` to set the delay between each message of the markov speech
    - `contextDepth <non-negative integer>` to set the context depth of the markov chain. 
    - `emoteSpamThreshold <int>` to set the amount of emotes to have to be spammed in a given time to trigger a response. 
    - `emoteSpamTimeout <duration>` to set the amount of time the emotes have to reach the spam threshold. 
    - `emoteSpamCooldown <duration>` to set the amount of time between when each emote gets spammed. emote1 can only get spammed once every duration, but this isn't a global cooldown and is per emote.  
- `!chatrat trust <username>` Adds a user to the trusted user list, letting them execute commands
- `!chatrat untrust <username>` Removes a user from the trusted user list
- `!chatrat ignore <username>` Ignores a user so that their messages are no longer added to the model or the chatlog.  
- `!chatrat unignore <username>` removes a user from the ignore list. 
- `!chatrat spam <emote>` adds an emote to spam to the list of mob spamming emotes. 
- `!chatrat <stopspamming/dontspam> <emote>` removes the emote from the list to mob mentality spam. Either of the two commands work. 

`<duration>` is either one "word" formatted like `<number><letter>`, ex: 5m, 1h, 2m30s or `<number> <unit>` like "1 minute" or "30 seconds" 

## Other stuff
ChatRat will join in on emote spamming, so if a whole bunch of people are saying an emote he is configured to also speak, he will do so. The example settings file has catKiss and heCrazy as the emotes to spam, but more can be added/removed there or with the commands. 

## How to run
All you need is a folder that has a `settings.json` (see provided example) and an executable in it. You can get the executable by running "go build" or by downloading an executable from one of the releases. A better guide can be seen on the wiki https://github.com/IAmPattycakes/ChatRat/wiki

## Contributions
PRs welcome, any features you may want can also go in the issues. Literally anything that sounds good, or funny, is welcome. This isn't really meant to be a practical chatbot, but a thing to make people laugh. 
