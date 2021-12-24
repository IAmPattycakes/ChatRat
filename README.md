# ChatRat
ChatRat is a twitch chat bot built in Go that is a dedicated shitpost machine. Also does some other things, but for now the main thing is just a markov shitpost generator. 

## Commands
Commands are permissioned, so only trusted users are allowed to execute them. Currently the commands are: 
- `!chatrat stop` Stops the constant markov text chatting
- `!chatrat start` Starts the constant markov text chatting if it was previously stopped. 
- `!chatrat set` Sets variables inside of ChatRat. These include: 
    - `delay <non-negative number> <minutes/seconds/hours>` to set the delay between each message of the markov speech
- `!chatrat trust <username>` Adds a user to the trusted user list, letting them execute commands
- `!chatrat untrust <username>` Removes a user from the trusted user list
- `!chatrat ignore <username>` Ignores a user 
- `!chatrat unignore <username>`
More to come, just gotta work on it

## Other stuff
Currently two emotes are set to join in a spam, catKiss and heCrazy. 

## How to run
Run it in CLI with some of the flags set, `-oauth=oauth:code_here -botname=name_of_twitch_user -stream=name_of_stream_to_join`
