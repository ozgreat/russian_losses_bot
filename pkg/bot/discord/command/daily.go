package command

import (
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/log"
	"github.com/disgoorg/snowflake/v2"
	"russian_losses/pkg/bot/discord/util"
	"strconv"
)

const ChangeDailyModeCommandName string = "daily"

func GetDailyCommand() discord.ApplicationCommandCreate {
	return discord.SlashCommandCreate{
		Name:        ChangeDailyModeCommandName,
		Description: "Увімкнути/Вимкнути щоденне відправлення статистики",
	}
}

func GetDailyListener() bot.EventListener {
	return &events.ListenerAdapter{OnApplicationCommandInteraction: handleDaily}
}

func handleDaily(event *events.ApplicationCommandInteractionCreate) {
	data := event.SlashCommandInteractionData()
	if data.CommandName() == ChangeDailyModeCommandName {
		chatId := event.Channel().ID().String()
		daily, err := util.ChangeDailyMode(chatId)
		if err != nil {
			log.Error(err)
			return
		}
		err = sendToChat(chatId, *daily, event.Client())
		if err != nil {
			log.Error(err)
		}
	}

}

func sendToChat(chatId string, daily bool, c bot.Client) error {
	iChatId, parseError := strconv.ParseInt(chatId, 10, 64)
	if parseError != nil {
		return parseError
	}

	var message string
	if !daily {
		message = "Щоденні звіщення увімкнуті"
	} else {
		message = "Щоденні звіщення вимкнуті"
	}

	_, createMessageErr := c.Rest().CreateMessage(
		snowflake.ID(iChatId),
		discord.NewMessageCreateBuilder().SetContent(message).Build(),
	)

	return createMessageErr
}
