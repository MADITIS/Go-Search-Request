# Cloud Search Telegram Bot

This is a Telegram bot written in Go that allows users to request and search for files in Google Drive using the Google Drive API. The bot uses Redis to cache search results and MongoDB to store information about pending, completed, and picked-up requests. The bot also has a web UI that users can use to view search results. The web UI is built using the Gin web framework.

## Features

- Request files: Users can submit requests for files in Google Drive using the bot.
- Search files: Users can search for files in Google Drive using the bot.
- Redis integration: The bot uses Redis to cache search results for faster responses.
- MongoDB integration: The bot uses MongoDB to store information about pending, completed, and picked-up requests.
- Web UI: The bot has a web UI that users can use to view search results.
- Gin web framework: The web UI is built using the Gin web framework.
- Docker support: Docker support will be added in the future.
- Easy to start: To start the bot, simply run `go run main.go start`.

## Installation

1. Clone the repository: `git clone https://github.com/username/cloud-search-telegram-bot.git`
2. Install the required packages: `go mod download`
3. Set up Redis: Install Redis and set up the Redis connection in `config/`.
4. Set up MongoDB: Install MongoDB and set up the MongoDB connection in `config/`.
5. Set up Telegram bot API: Follow the instructions [here](https://core.telegram.org/bots#3-how-do-i-create-a-bot) to obtain a bot token and add it to `config/config.yaml`.
6. Set up Google Drive API: Follow the instructions [here](https://developers.google.com/drive/api/v3/quickstart/go) to obtain the required credentials and add them to `config/config.yaml`.
7. Start the bot: Run `go run main.go start`.

## Usage

1. Request files: Users can submit requests for files by sending a message to the bot with the format `/request filename`.
2. Search files: Users can search for files by sending a message to the bot with the format `/search query`.
3. View search results: Users can view search results by opening the web UI at `http://localhost:8080/search`.
4. Pick up requests: Other users can pick up requests to fill them by sending a message to the bot with the format `/pickup request_id`.

## Credits

This project uses the following libraries and APIs:

- Telegram Bot API: https://core.telegram.org/bots/api
- Redis: https://redis.io/
- MongoDB: https://www.mongodb.com/
- Gin web framework: https://github.com/gin-gonic/gin
- Google Drive API: https://developers.google.com/drive/api

## License

This project is licensed under the MIT License. See the [LICENSE](https://github.com/username/cloud-search-telegram-bot/blob/main/LICENSE) file for details.
