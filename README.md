### Prerequisites

Before you begin, ensure you have the following software installed:

- [Go](https://golang.org/dl/) (version 1.20 or higher)
- Docker (optional, for containerized deployment)

### Installing

1. Install Go dependencies
   ```sh
   go mod tidy
   ```

2. Configuring Environment Variables
    Obtain Telegram Bot API at [BotFather](https://t.me/BotFather)

    Get Gemini API keys from [Google AI Studio](https://makersuite.google.com/app/apikey)
    ```sh
    export GOOGLE_GEMINI_KEY='your_google_gemini_key'
    export TELEGRAM_BOT_TOKEN='your_telegram_bot_token
    ```
3. Running the Application
    ```sh
    go run main.go
    ```
### Run from docker
    ```sh
    docker build -t tg_gemini_bot .
    docker run -e GOOGLE_GEMINI_KEY='your_google_gemini_key' -e TELEGRAM_BOT_TOKEN='your_telegram_bot_token' tg_gemini_bot

    ```

### Deploy

You can click on the button below to deploy the bot to Zeabur.

[![Deploy on Zeabur](https://zeabur.com/button.svg)](https://zeabur.com/templates/Z8668N)

### Reference

https://github.com/yihong0618/tg_bot_collections
