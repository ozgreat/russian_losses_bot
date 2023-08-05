package command

import (
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/log"
	"russian_losses/pkg/bot/discord/util"
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
		err = sendToChat(*daily, event)
		if err != nil {
			log.Error(err)
		}
	}

}

func sendToChat(daily bool, event *events.ApplicationCommandInteractionCreate) error {
	var message string
	if !daily {
		message = "Щоденні звіщення увімкнуті"
	} else {
		message = "Щоденні звіщення вимкнуті"
	}

	return event.CreateMessage(discord.NewMessageCreateBuilder().SetContent(message).Build())
}
