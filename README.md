[![Go Version](https://img.shields.io/github/go-mod/go-version/chofnar/release-bot?logo=go)](go.mod)
[![Telegram Chat](https://img.shields.io/static/v1?label=Bot&message=chat&color=29a1d4&logo=telegram)](https://t.me/prgitrelbot)
[![Go Report](https://img.shields.io/badge/Go%20Report-A+-brightgreen.svg?style=flat)](https://goreportcard.com/report/github.com/chofnar/release-bot)

# release-bot - a telegram bot for Github releases

This is a Telegram bot that monitors the releases of given repos, sending messages upon a new release. 
It uses [mymmrac's Telegram Bot API implementation in Go](https://github.com/mymmrac/telego).

To store all the users and their repos, the bot uses AWS DynamoDB.

To access the details of the repos, the bot queries the Github GraphQL endpoint.

Right now, the bot is running using Google Cloud Run, with a scheduler to call the updateRepos endpoint.

This is basically a rewrite of my [other Telegram bot](https://github.com/chofnar/BasicGithubReleasesTelegramBot) after I became a lot wiser in the ways of software development.

Interact with the bot [here](https://t.me/prgitrelbot)

## Running it yourself
### DynamoDB
Create a table that has the primary key called "chatID" (string), and sort key called "repoID" (string).

### Set the necessary env vars
TELEGRAM_BOT_TOKEN - get this from [BotFather](https://t.me/botfather). You'll need to create a bot.

TELEGRAM_BOT_SITE_URL - URL used for listening for incoming requests from the Telegram servers. If running locally, you may want to use [ngrok](https://ngrok.com/)

PORT - the port that the application will listen to for requests

BOT_DYNAMODB_ENDPOINT - self explainatory. This bot uses DynamoDB. Put the endpoint that includes the region where your table is located

BOT_TABLE_NAME - the table name from DynamoDB

SUPER_SECRET_TOKEN - a random string. You must send this in the body of a post request to the /updateRepos endpoint, else the request will be dismissed.

GITHUB_GQL_TOKEN - you'll have to find out how to get this yourself.

AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, AWS_DEFAULT_REGION - you'll have to find out how to get these yourself.

### The Go part
Download the modules needed
```
go mod download
```

Build the binary
```
go build
```

Run the binary
```
./release-bot
```
