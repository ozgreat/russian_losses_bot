package command

import (
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/log"
	"github.com/disgoorg/snowflake/v2"
	"russian_losses/pkg/db"
	l "russian_losses/pkg/losses"
	"strconv"
	"strings"
)

const StatCommandName string = "stat"

func GetStatCommand() discord.ApplicationCommandCreate {
	return discord.SlashCommandCreate{
		Name:        StatCommandName,
		Description: "Відправляє свіжу статистику втрат окупантів",
	}
}

func GetStatListener() bot.EventListener {
	return &events.ListenerAdapter{OnApplicationCommandInteraction: handleStat}
}

func handleStat(event *events.ApplicationCommandInteractionCreate) {
	data := event.SlashCommandInteractionData()
	if data.CommandName() == StatCommandName {
		err := sendStatisticsToSingleChat(event.Client(), event.Channel().ID().String())
		if err != nil {
			log.Error(err)
		}
	}
}

func sendStatisticsToSingleChat(c bot.Client, chatId string) error {
	i, err := l.GetFreshInfo()
	if err != nil {
		return err
	}

	return sendStat(c, db.ChatEntity{ChatId: chatId, BotPlatform: db.Discord}, i)
}

func SendStatistics(client bot.Client, chats []db.ChatEntity, info *l.StatisticOfLoses) error {
	for _, chat := range chats {
		if chat.BotPlatform != db.Discord || !chat.DailyNotification {
			continue
		}

		err := sendStat(client, chat, info)
		if err != nil {
			return err
		}
	}

	return nil
}

func sendStat(client bot.Client, chat db.ChatEntity, info *l.StatisticOfLoses) error {
	iChatId, parseError := strconv.ParseInt(chat.ChatId, 10, 64)

	if parseError != nil {
		return parseError
	}

	go func() {
		message := info.ToMessage()

		// Fix for discord specific MarkDown
		message = strings.ReplaceAll(message, "*", "**")

		_, err := client.Rest().CreateMessage(
			snowflake.ID(iChatId),
			discord.NewMessageCreateBuilder().SetContent(message).Build(),
		)
		if err != nil {
			log.Error(err)
		}
	}()
	return nil
}
