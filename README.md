# Mintlayer Telegram Bot

Mintlayer's Telegram bot is a powerful tool designed to interact with Mintlayer directly from Telegram. It allows users to manage pools, check balances, and receive notifications on balance changes.

## Available Commands

- `/pool_add <poolID>` - Add a pool
- `/pool_remove <poolID>` - Remove a pool
- `/pool_list` - List your pools
- `/balance` - Get the total balance of your pools
- `/notify_start` - Notify on balance change
- `/notify_stop` - Stop balance change notifications

## Installation

To set up the Mintlayer Telegram Bot, follow these steps:

1. Clone the repository to your local machine.
2. Ensure you have Go installed on your system. You can download it from [https://golang.org/dl/](https://golang.org/dl/).
3. Navigate to the cloned repository directory.
4. Run `go build` to compile the bot executable.
5. Create a `config.json` file in the same directory as the executable with the following content:

```
{
  "bot_token": "<TELEGRAM_TOKEN_ID>",
  "api_base_url": "https://api-server.mintlayer.org",
  "admin_user": "<TELEGRAM_USER_ID>"
}
```

Replace `<TELEGRAM_TOKEN_ID>` with your actual bot token from Telegram. Set `api_base_url` if you run a local api-server, otherwise keep the default. Set `admin_user` to your Telegram numeric user ID to enable admin-only commands like `/broadcast`.

### Generating Telegram Bot Token

To generate a Telegram bot token:

1. Open Telegram and search for the `BotFather`.
2. Send `/newbot` command and follow the instructions to create your bot.
3. Once created, `BotFather` will give you a token. Use this token in your `config.json`.

## Security

Ensure your `config.json` is properly secured and not accessible by unauthorized individuals. Never share your bot token.

## License

This project is licensed under the MIT License - see the LICENSE file for details.

