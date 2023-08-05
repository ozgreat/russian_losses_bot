package bot

import (
	"log"
	"os"
	"russian_losses/pkg/bot/discord"
	"russian_losses/pkg/bot/telegram"
	"russian_losses/pkg/db"
	"russian_losses/pkg/losses"
)

type IBot interface {
	// SendStatistics Sends formatted losses.StatisticOfLoses to all chats from chatIds array
	SendStatistics(chats []db.ChatEntity, info *losses.StatisticOfLoses) error
	// AddChat Adds chat to db to save it for later to scheduled static sending
	AddChat(chatId string) error
	// StopBot Stops bot, called before application stop
	StopBot()
}

// getBots returns bots
func GetBots() []IBot {
	discordBot, discordError := discord.GetDiscordBot()
	if discordError != nil {
		log.Panic(discordError)
		os.Exit(1)
	}

	telegramBot, telergramError := telegram.GetTelegramBot()
	if telergramError != nil {
		log.Panic(telergramError)
		os.Exit(1)
	}

	return []IBot{discordBot, telegramBot}
}
