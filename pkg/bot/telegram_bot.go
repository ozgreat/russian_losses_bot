package bot

import (
	"os"
	"strconv"

	"github.com/disgoorg/log"
	tg "github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"

	"russian_losses/pkg/db"
	"russian_losses/pkg/losses"
)

// TelegramBot singleton
var tgBot *TelegramBot = nil

type TelegramBot struct {
	client   *tg.Bot
	stopFunc func()
}

// GetTelegramBot Returns TelegramBot singleton
func GetTelegramBot() (*TelegramBot, error) {
	if tgBot == nil {
		var err error
		tgBot, err = newTelegramBot(os.Getenv("telegramToken"))
		return tgBot, err
	}

	return tgBot, nil
}

func (b TelegramBot) SendStatistics(chats []db.ChatEntity, info *losses.StatisticOfLoses) error {
	for _, chat := range chats {
		if chat.BotPlatform != db.Telegram {
			continue
		}
		chatId, parseError := strconv.ParseInt(chat.ChatId, 10, 64)
		if parseError != nil {
			return parseError
		}

		go func() { sendInfo(&db.ChatEntity{ChatId: strconv.FormatInt(chatId, 10)}, b.client, info) }()
	}

	return nil
}

func (TelegramBot) AddChat(chatId string) error {
	service, err := db.GetDbService()
	if err != nil {
		return err
	}

	return service.InsertChatId(chatId, db.Telegram)
}

func (b TelegramBot) StopBot() {
	b.stopFunc()
}

// Creates TelegramBot which uses tg.Bot with long polling bot mode
func newTelegramLongPollingBot(token string) (*TelegramBot, error) {
	return createTgBot(
		token,
		func(bot *tg.Bot) {},
		handleUpdatesLongPolling,
		func(bot *tg.Bot) func() { return bot.StopLongPolling },
	)
}

// Creates TelegramBot which uses tg.Bot with webhook bot mode
func newTelegramWebhookBot(token string, webHookUrl string) (*TelegramBot, error) {
	path := "/bot" + token
	return createTgBot(
		token,
		func(bot *tg.Bot) { startWebHookBot(bot, webHookUrl, path) },
		func(bot *tg.Bot) { handleUpdatesWebHookBot(bot, path) },
		func(bot *tg.Bot) func() { return bot.StopLongPolling },
	)
}

func handleUpdatesLongPolling(bot *tg.Bot) {
	updates, updatesErr := bot.UpdatesViaLongPolling(nil)
	if updatesErr != nil {
		log.Panic(updatesErr)
		os.Exit(1)
	}

	handleUpdates(bot, updates)
}

func startWebHookBot(bot *tg.Bot, webHookUrl string, path string) {
	err := bot.SetWebhook(&tg.SetWebhookParams{URL: webHookUrl + path})
	if err != nil {
		log.Panic(err)
		os.Exit(1)
	}

	go func() {
		err = bot.StartWebhook("localhost:8081")
		if err != nil {
			log.Error(err)
		}
	}()
}

func handleUpdatesWebHookBot(bot *tg.Bot, path string) {
	updates, updatesErr := bot.UpdatesViaWebhook(path)
	if updatesErr != nil {
		log.Panic(updatesErr)
		os.Exit(1)
	}

	handleUpdates(bot, updates)
}

// Creates and configure TelegramBot, uses function params to set onStart, onStop and onUpdate action
func createTgBot(token string, onStart func(bot *tg.Bot), onUpdate func(bot *tg.Bot),
	onStop func(bot *tg.Bot) func(),
) (*TelegramBot, error) {
	bot, err := tg.NewBot(token, tg.WithDefaultLogger(false, true))
	if err != nil {
		return nil, err
	}

	go onStart(bot)
	go onUpdate(bot)

	return &TelegramBot{bot, onStop(bot)}, nil
}

func newTelegramBot(token string) (*TelegramBot, error) {
	webHookUrl := os.Getenv("tgWebhookHost")

	if webHookUrl != "" {
		return newTelegramWebhookBot(token, webHookUrl)
	} else {
		return newTelegramLongPollingBot(token)
	}
}

func handleUpdates(bot *tg.Bot, updates <-chan tg.Update) {
	h, err := th.NewBotHandler(bot, updates)
	if err != nil {
		log.Error(err)
	}

	defer h.Stop()

	// Handle bot join
	h.Handle(handleAddRemoveChat, th.AnyMyChatMember())

	// Handle commands
	h.Handle(handleStat, th.CommandEqual("stat"))

	h.Start()
}

func handleAddRemoveChat(bot *tg.Bot, update tg.Update) {
	member := update.MyChatMember
	chatId := member.Chat.ID
	status := member.NewChatMember.MemberStatus()
	me, _ := bot.GetMe()
	isMe := me.ID == member.NewChatMember.MemberUser().ID
	chatFromDb, err := getChatFromDb(chatId)
	if err != nil {
		log.Error(err)
	}

	if chatFromDb != nil {
		return
	}

	if status == "member" && isMe {
		_ = addChat(strconv.FormatInt(chatId, 10))
	} else if status == "left" && isMe {
		_ = removeChat(strconv.FormatInt(chatId, 10))
	}
}

func handleStat(bot *tg.Bot, update tg.Update) {
	info, losesErr := losses.GetFreshInfo()
	if losesErr != nil {
		log.Error(losesErr)
	}

	i := strconv.FormatInt(update.Message.Chat.ID, 10)
	sendInfo(&db.ChatEntity{ChatId: i}, bot, info)
}

func getChatFromDb(id int64) (*db.ChatEntity, error) {
	service, dbError := db.GetDbService()
	if dbError != nil {
		return nil, dbError
	}

	return service.GetChat(strconv.FormatInt(id, 10), db.Telegram)
}

func addChat(id string) error {
	service, dbError := db.GetDbService()
	if dbError != nil {
		return dbError
	}

	return service.InsertChatId(id, db.Telegram)
}

func removeChat(id string) error {
	service, dbError := db.GetDbService()
	if dbError != nil {
		return dbError
	}

	return service.RemoveChat(id, db.Telegram)
}

func sendInfo(chat *db.ChatEntity, bot *tg.Bot, info *losses.StatisticOfLoses) {
	id, parseErr := strconv.ParseInt(chat.ChatId, 10, 64)
	if parseErr != nil {
		log.Error(parseErr)
		return
	}

	if chat.Format == "" {
		var getChatErr error
		chat, getChatErr = getChatFromDb(id)
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
