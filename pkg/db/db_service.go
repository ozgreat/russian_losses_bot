package db

import (
	"database/sql"
	"os"

	"github.com/disgoorg/log"
)

var service *Service = nil

type Service struct {
	conn *sql.DB
}

func GetDbService() (*Service, error) {
	if service == nil {
		var err error
		service, err = newDbService()
		if err != nil {
			return nil, err
		}
	}

	return service, nil
}

func newDbService() (*Service, error) {
	conn, err := openConnection()
	if err != nil {
		return nil, err
	}

	return &Service{conn}, nil
}

func (s Service) InsertChatId(chatId string, platform Platform) error {
	return s.insertChatId(ChatEntity{ChatId: chatId, BotPlatform: platform}, nil)
}

func (s Service) insertChatId(chat ChatEntity, prevErr error) error {
	_, queryErr := s.conn.Query("insert into chats (chat_id, platform) values ($1, $2);",
		chat.ChatId,
		chat.BotPlatform,
	)

	if queryErr != nil {
		return s.handleQueryErrorAndRetry(chat, queryErr, prevErr, s.insertChatId)
	}

	return nil
}

func (s Service) UpdateFormat(chat *ChatEntity, format StatFormat) error {
	_, err := s.conn.Query(`
			update chats 
			set format = $1  
			where chat_id = $2 and platform = $3;`,
		format,
		chat.ChatId,
		chat.BotPlatform,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s Service) GetChat(chatId string, platform Platform) (*ChatEntity, error) {
	return s.getChat(ChatEntity{ChatId: chatId, BotPlatform: platform}, nil)
}

func (s Service) getChat(chat ChatEntity, prevErr error) (*ChatEntity, error) {
	var chatId string
	var platform Platform
	var format StatFormat

	row := s.conn.QueryRow(
		"select chat_id, platform, format from chats where chat_id = $1 and platform = $2 limit 1;",
		chat.ChatId,
		chat.BotPlatform,
	)
	queryErr := row.Scan(&chatId, &platform, &format)

	if queryErr != nil {
		if queryErr == sql.ErrNoRows {
			return nil, nil
		}
		log.Warn(queryErr)
		if prevErr != nil {
			return nil, queryErr
		}

		var openErr error
		s.conn, openErr = openConnection()
		if openErr != nil {
			return nil, openErr
		}

		return s.getChat(chat, queryErr)
	}

	return &ChatEntity{ChatId: chatId, BotPlatform: platform, Format: format}, nil
}

func (s Service) RemoveChat(chatId string, platform Platform) error {
	return s.removeChat(ChatEntity{ChatId: chatId, BotPlatform: platform}, nil)
}

func (s Service) removeChat(chat ChatEntity, prevErr error) error {
	_, queryErr := s.conn.Query(
		"delete from chats where chat_id = $1 and platform = $2",
		chat.ChatId,
		chat.BotPlatform,
	)

	if queryErr != nil {
		return s.handleQueryErrorAndRetry(chat, queryErr, prevErr, s.removeChat)
	}

	return nil
}

func (s Service) GetAllChats() ([]ChatEntity, error) {
	return s.getAllChats(nil)
}

func (s Service) getAllChats(prevErr error) ([]ChatEntity, error) {
	rows, queryError := s.conn.Query("select chat_id, platform, format from chats;")
	if queryError != nil {
		if prevErr != nil {
			return nil, queryError
		}

		var openError error
		s.conn, openError = openConnection()
		if openError != nil {
			return nil, openError
		}

		return s.getAllChats(queryError)
	}

	var result []ChatEntity
	for rows.Next() {
		var chatId string
		var platform Platform
		var format StatFormat

		scanErr := rows.Scan(&chatId, &platform, &format)
		if scanErr != nil {
			return nil, scanErr
		}

		result = append(
			result,
			ChatEntity{ChatId: chatId, BotPlatform: platform, Format: format},
		)
	}

	return result, nil
}

func (s Service) handleQueryErrorAndRetry(chat ChatEntity, queryErr error, prevErr error,
	retryFunc dbQuery,
) error {
	err := s.handleErrorAndOpenDb(queryErr, prevErr)
	if err != nil {
		return err
	}

	return retryFunc(chat, queryErr)
}

func (s Service) handleErrorAndOpenDb(queryErr error, prevErr error) error {
	log.Warn(queryErr)
	if prevErr != nil {
		return queryErr
	}

	var openErr error
	s.conn, openErr = openConnection()
	if openErr != nil {
		return openErr
	}

	if s.conn != nil {
		log.Error("Connection is nil")
	}

	return nil
}

func openConnection() (*sql.DB, error) {
	connStr := os.Getenv("DATABASE_URL")
	db, connErr := sql.Open("postgres", connStr)
	db.SetMaxIdleConns(20)
	db.SetConnMaxIdleTime(30000000000)
	db.SetConnMaxLifetime(30000000000)
	if connErr != nil {
		return nil, connErr
	}

	return db, nil
}
