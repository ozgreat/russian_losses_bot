package api

import (
	"github.com/disgoorg/log"
	"io"
	"net/http"

	"russian_losses/pkg/bot"
	"russian_losses/pkg/db"
	l "russian_losses/pkg/losses"
)

func HandleRequests() {
	bot.GetBots()
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
		return
	}

	sendStatistics(chats, info)
	_, err = io.WriteString(w, "OK")
	if err != nil {
		log.Error(err)
	}
}

// onStop call bot.IBot StopBot function before application stop
func onStop() {
	for _, statisticBot := range bot.GetBots() {
		statisticBot.StopBot()
	}
}

func sendStatistics(chats []db.ChatEntity, info *l.StatisticOfLoses) {
	for _, iBot := range bot.GetBots() {
		err := iBot.SendStatistics(chats, info)
		if err != nil {
			log.Error(err)
		}
	}
}
