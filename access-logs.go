package main

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type AccessLog struct {
	UserID          string
	RoomID          string
	Method          string
	IsGrantedAccess bool
	Reason          string
}

type accessLogService struct {
	db *sqlx.DB
}

func NewAccessLogService(db *sqlx.DB) *accessLogService {
	return &accessLogService{db: db}
}

func (s *accessLogService) Create(ctx context.Context, log AccessLog) error {
	query := `INSERT INTO user_room_access_log (userId, method, roomId, isGrantedAccess, reason, createdAt) VALUES (?, ?, ?, ?, ?, NOW())`
	_, err := s.db.ExecContext(ctx, query, log.UserID, log.Method, log.RoomID, log.IsGrantedAccess, log.Reason)
	return err
}
