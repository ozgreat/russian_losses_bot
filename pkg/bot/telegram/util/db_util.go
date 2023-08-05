package util

import (
	"russian_losses/pkg/db"
	"strconv"
)

func ChangeDailyMode(chat db.ChatEntity, daily bool) error {
	service, dbError := db.GetDbService()
	if dbError != nil {
		return dbError
	}

	return service.UpdateDaily(chat, daily)
}

func GetChatFromDb(id int64) (*db.ChatEntity, error) {
	service, dbError := db.GetDbService()
	if dbError != nil {
		return nil, dbError
	}

	return service.GetChat(strconv.FormatInt(id, 10), db.Telegram)
}

func AddChat(id string) error {
	service, dbError := db.GetDbService()
	if dbError != nil {
		return dbError
	}

	return service.InsertChatId(id, db.Telegram)
}

func RemoveChat(id string) error {
	service, dbError := db.GetDbService()
	if dbError != nil {
		return dbError
	}

	return service.RemoveChat(id, db.Telegram)
}
