package telegram

import (
	"os"
	"russian_losses/pkg/bot/telegram/command"
	"russian_losses/pkg/bot/telegram/util"
	"strconv"

	"github.com/disgoorg/log"
	tg "github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	"russian_losses/pkg/db"
	"russian_losses/pkg/losses"
)

// Bot singleton
var tgBot *Bot = nil

type Bot struct {
	client   *tg.Bot
	stopFunc func()
}

// GetTelegramBot Returns Bot singleton
func GetTelegramBot() (*Bot, error) {
	if tgBot == nil {
		var err error
		tgBot, err = newTelegramBot(os.Getenv("telegramToken"))
		return tgBot, err
	}

	return tgBot, nil
}

func (b Bot) SendStatistics(chats []db.ChatEntity, info *losses.StatisticOfLoses) error {
	for _, chat := range chats {
		if chat.BotPlatform != db.Telegram || !chat.DailyNotification {
			continue
		}
		chatId, parseError := strconv.ParseInt(chat.ChatId, 10, 64)
		if parseError != nil {
			return parseError
		}

		go func() { command.SendInfo(&db.ChatEntity{ChatId: strconv.FormatInt(chatId, 10)}, b.client, info) }()
	}

	return nil
}

func (Bot) AddChat(chatId string) error {
	service, err := db.GetDbService()
	if err != nil {
		return err
	}

	return service.InsertChatId(chatId, db.Telegram)
}

func (b Bot) StopBot() {
	b.stopFunc()
}

// Creates Bot which uses tg.Bot with long polling bot mode
func newTelegramBot(token string) (*Bot, error) {
	bot, err := tg.NewBot(token, tg.WithDefaultLogger(false, true))
	if err != nil {
		return nil, err
	}

	go handleUpdatesLongPolling(bot)

	return &Bot{bot, bot.StopLongPolling}, nil
}

func handleUpdatesLongPolling(bot *tg.Bot) {
	updates, updatesErr := bot.UpdatesViaLongPolling(nil)
	if updatesErr != nil {
		log.Panic(updatesErr)
		os.Exit(1)
	}

	handleUpdates(bot, updates)
}

func handleUpdates(bot *tg.Bot, updates <-chan tg.Update) {
	h, err := th.NewBotHandler(bot, updates)
	if err != nil {
		log.Error(err)
	}

	defer h.Stop()

	// Handle bot join
	h.Handle(handleAddRemoveChat, th.AnyMyChatMember())

	// Handle command
	h.Handle(command.HandleStat, th.CommandEqual("stat"))
	h.Handle(command.HandleChangeDailyMode, th.CommandEqual("daily"))

	h.Start()
}

func handleAddRemoveChat(bot *tg.Bot, update tg.Update) {
	member := update.MyChatMember
	chatId := member.Chat.ID
	status := member.NewChatMember.MemberStatus()
	me, _ := bot.GetMe()
	isMe := me.ID == member.NewChatMember.MemberUser().ID
	chatFromDb, err := util.GetChatFromDb(chatId)
	if err != nil {
		log.Error(err)
	}

	if chatFromDb != nil {
		return
	}

	if status == "member" && isMe {
		_ = util.AddChat(strconv.FormatInt(chatId, 10))
	} else if status == "left" && isMe {
		_ = util.RemoveChat(strconv.FormatInt(chatId, 10))
	}
}
