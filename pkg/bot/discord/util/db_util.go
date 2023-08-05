package util

import (
	"github.com/disgoorg/log"
	"russian_losses/pkg/db"
	"strconv"
)

func ChangeDailyMode(chatId string) (*bool, error) {
	iChatId, parseError := strconv.ParseInt(chatId, 10, 64)
	if parseError != nil {
		return nil, parseError
	}

	service, dbError := db.GetDbService()
	if dbError != nil {
		return nil, dbError
	}

	chat, getError := service.GetChat(chatId, db.Discord)
	if getError != nil {
		return nil, getError
	}

	updateErr := updateDaily(iChatId, !chat.DailyNotification)
	if updateErr != nil {
		return nil, updateErr
	}

	return &chat.DailyNotification, nil
}

func updateDaily(chatId int64, daily bool) error {
	service, openDbErr := db.GetDbService()
	if openDbErr != nil {
		return openDbErr
	}

	id := strconv.FormatInt(chatId, 10)

	return service.UpdateDaily(db.ChatEntity{ChatId: id, BotPlatform: db.Discord}, daily)
}

func AddChat(channelId string) error {
	service, err := db.GetDbService()
	if err != nil {
		return err
	}

	return service.InsertChatId(channelId, db.Discord)
}

func RemoveChat(channelId string) {
	service, dbErr := db.GetDbService()
	if dbErr != nil {
		log.Error(dbErr)
	}

	removeErr := service.RemoveChat(channelId, db.Discord)
	if removeErr != nil {
		log.Error(removeErr)
	}
}
