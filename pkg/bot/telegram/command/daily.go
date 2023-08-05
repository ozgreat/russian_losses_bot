package command

import (
	"github.com/disgoorg/log"
	tg "github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
	"russian_losses/pkg/bot/telegram/util"
)

func HandleChangeDailyMode(bot *tg.Bot, update tg.Update) {
	chatId := update.Message.Chat.ID
	chat, getFromDbErr := util.GetChatFromDb(chatId)
	if getFromDbErr != nil {
		log.Error(getFromDbErr)
	}

	daily := !chat.DailyNotification

	err := util.ChangeDailyMode(*chat, daily)
	if err != nil {
		log.Error(err)
		return
	}

	var message string
	if daily {
		message = "Щоденні звіщення увімкнуті"
	} else {
		message = "Щоденні звіщення вимкнуті"
	}
	_, err = bot.SendMessage(&tg.SendMessageParams{
		ChatID:    tu.ID(chatId),
		Text:      message,
		ParseMode: tg.ModeMarkdown,
	})
	if err != nil {
		log.Error(err)
	}
}
