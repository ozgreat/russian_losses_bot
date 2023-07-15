package api

import (
	"io"
	"net/http"
	"os"

	"github.com/disgoorg/log"

	"russian_losses/pkg/bot"
	"russian_losses/pkg/db"
	l "russian_losses/pkg/losses"
)

func HandleRequests() {
	getBots()
	defer onStop()
	http.HandleFunc("/stat", handleStat)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleStat(w http.ResponseWriter, _ *http.Request) {
	service, err := db.GetDbService()
	if err != nil {
		log.Error(err)
		return
	}

	chats, err := service.GetAllChats()
	if err != nil {
		log.Error(err)
		return
	}

	info, err := l.GetFreshInfo()
	if err != nil {
		log.Error(err)
	}

	sendStatistics(chats, info)
	io.WriteString(w, "OK")
}

// onStop call bot.IBot StopBot function before application stop
func onStop() {
	for _, statisticBot := range getBots() {
		statisticBot.StopBot()
	}
}

// getBots returns bots
func getBots() []bot.IBot {
	discordBot, discordError := bot.GetDiscordBot()
	if discordError != nil {
		log.Panic(discordError)
		os.Exit(1)
	}

	telegramBot, telergramError := bot.GetTelegramBot()
	if telergramError != nil {
		log.Panic(telergramError)
		os.Exit(1)
	}

	return []bot.IBot{discordBot, telegramBot}
}

func sendStatistics(chats []db.ChatEntity, info *l.StatisticOfLoses) {
	for _, iBot := range getBots() {
		err := iBot.SendStatistics(chats, info)
		if err != nil {
			log.Error(err)
		}
	}
}
