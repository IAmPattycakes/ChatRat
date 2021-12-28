# ChatRat
ChatRat is a twitch chat bot built in Go that is a dedicated shitpost machine. Also does some other things, but for now the main thing is just a markov shitpost generator. He tries to act like a normal member of chat, but isn't the brightest so he can kinda go off the rails sometimes. 

## Commands
Commands are permissioned, so only trusted users are allowed to execute them. Currently the commands are: 
- `!chatrat stop` Stops the constant markov text chatting
- `!chatrat start` Starts the constant markov text chatting if it was previously stopped. 
- `!chatrat delay` Shows the current speech delay. 
- `!chatrat set` Sets variables inside of ChatRat. These include: 
    - `delay <non-negative number> <minutes/seconds/hours>` to set the delay between each message of the markov speech
    - `contextDepth <non-negative integer>` to set the context depth of the markov chain. 
- `!chatrat trust <username>` Adds a user to the trusted user list, letting them execute commands
- `!chatrat untrust <username>` Removes a user from the trusted user list
- `!chatrat ignore <username>` Ignores a user so that their messages are no longer added to the model or the chatlog.  
- `!chatrat unignore <username>` removes a user from the ignore list. 
More to come, just gotta work on it

## Other stuff
Currently two emotes are set to join in a spam, catKiss and heCrazy. These are not currently very configurable, The only thing you can do is set the threshold/delay/cooldown in the settings.json, but these will be refactored eventually. 

## How to run
All you need is a folder that has a `settings.json` (see provided example) and an executable in it. You can get the executable by running "go build" or by downloading an executable from one of the releases. A better guide can be seen on the wiki https://github.com/IAmPattycakes/ChatRat/wiki

## Contributions
PRs welcome, any features you may want can also go in the issues. Literally anything that sounds good, or funny, is welcome. This isn't really meant to be a practical chatbot, but a thing to make people laugh. 
