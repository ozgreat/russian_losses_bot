package command

import (
	"github.com/disgoorg/log"
	tg "github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
	"russian_losses/pkg/bot/telegram/util"
	"russian_losses/pkg/db"
	"russian_losses/pkg/losses"
	"strconv"
)

func HandleStat(bot *tg.Bot, update tg.Update) {
	info, losesErr := losses.GetFreshInfo()
	if losesErr != nil {
		log.Error(losesErr)
	}

	i := strconv.FormatInt(update.Message.Chat.ID, 10)
	SendInfo(&db.ChatEntity{ChatId: i}, bot, info)
}

func SendInfo(chat *db.ChatEntity, bot *tg.Bot, info *losses.StatisticOfLoses) {
	id, parseErr := strconv.ParseInt(chat.ChatId, 10, 64)
	if parseErr != nil {
		log.Error(parseErr)
		return
	}

	if chat.Format == "" {
		var getChatErr error
		chat, getChatErr = util.GetChatFromDb(id)
		if getChatErr != nil {
			log.Error(getChatErr)
			return
		}
	}

	var err error
	format := chat.Format
	if format == db.Text {
		err = sendTextInfo(id, bot, info)
	} else {
		err = db.FormatError{Msg: "Unknown format"}
	}

	if err != nil {
		log.Error(err)
	}
}

func sendTextInfo(chatId int64, bot *tg.Bot, info *losses.StatisticOfLoses) error {
	message := info.ToMessage()
	_, err := bot.SendMessage(&tg.SendMessageParams{
		ChatID:    tu.ID(chatId),
		Text:      message,
		ParseMode: tg.ModeMarkdown,
	})
	return err
}
