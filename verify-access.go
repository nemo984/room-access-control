package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/jmoiron/sqlx"
)

type AccessType string

const (
	AccessTypeFP  AccessType = "FINGERPRINT"
	AccessTypeNFC AccessType = "NFC"
)

type AccessRequest struct {
	SensorID string     `json:"sensorId"`
	Key      string     `json:"key"`
	Type     AccessType `json:"type"`
}

type QueryResult struct {
	ScheduleID        string  `db:"scheduleId"`
	ScheduleName      string  `db:"scheduleName"`
	RoleID            string  `db:"roleId"`
	RoleName          string  `db:"roleName"`
	UserID            string  `db:"userId"`
	Username          string  `db:"username"`
	UserFingerprintID *string `db:"userFingerprintId"`
	UserNfcID         *string `db:"userNfcId"`
	RoomID            string  `db:"roomId"`
	RoomSnsTopicARN   *string `db:"roomSnsTopicArn"`
	SensorID          string  `db:"sensorId"`
	Type              string  `db:"type"`
}

type AccessLogService interface {
	Create(context context.Context, log AccessLog) error
}

type handler struct {
	db               *sqlx.DB
	accessLogService AccessLogService
	snsClient        *sns.Client
}

func NewHandler(db *sqlx.DB, accessLogService AccessLogService, snsClient *sns.Client) *handler {
	return &handler{
		db:               db,
		accessLogService: accessLogService,
		snsClient:        snsClient,
	}
}

func (h handler) VerifyAccess(w http.ResponseWriter, req *http.Request) {
	r := AccessRequest{}
	if err := json.NewDecoder(req.Body).Decode(&r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	currentDate := time.Now()
	currentTime := currentDate.Format("15:04")
	currentDay := currentDate.Weekday().String()

	var result QueryResult
	query := `
		SELECT
			a.id AS "scheduleId",
			a.name AS "scheduleName",
			ro.id AS "roleId",
			ro.name AS "roleName",
			us.id AS "userId",
			us.name AS "username",
			us."fingerprintId" AS "userFingerprintId",
			us."nfcId" AS "userNfcId",
			roo.id AS "roomId",
			roo."snsTopicArn" AS "roomSnsTopicArn",
			s.id AS "sensorId",
			s.type AS "type"
		FROM public."AccessSchedule" AS a
		INNER JOIN public."AccessScheduleTime" AS t ON t."accessScheduleId" = a.id
		INNER JOIN public."_AccessScheduleToRole" AS r ON r."A" = a.id
		INNER JOIN public."Role" AS ro ON ro.id = r."B"
		INNER JOIN public."_RoleToUser" AS u ON u."A" = ro.id
		INNER JOIN public."User" AS us ON us.id = u."B"
		INNER JOIN public."_AccessScheduleToRoom" AS rm ON rm."A" = a.id
		INNER JOIN public."Room" AS roo ON roo.id = rm."B"
		INNER JOIN public."DoorAccessSensor" AS s ON s."roomId" = roo.id
		LEFT JOIN public."Suspension" AS susp ON susp."userId" = us.id
		WHERE a.active = 'true'
		  AND t.day = $1
		  AND t.from <= $2
		  AND t.to >= $3
		  AND (
			us."fingerprintId" = $4
			OR us."nfcId" = $5
		  )
		  AND (
			susp.id IS NULL
			OR (
				susp."isPermanent" = 'false'
		  		AND (susp."startDate" > $6 OR susp."endDate" < $7)
			)
		  )
		  AND s.id = $8
  		LIMIT 1`

	err := h.db.Get(&result, query, currentDay, currentTime, currentTime, r.Key, r.Key, currentDate, currentDate, r.SensorID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "No matching schedule found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	slog.Info("Access granted", "scheduleId", result.ScheduleID, "scheduleName", result.ScheduleName, "roleID", result.RoleID, "roleName", result.RoleName, "userID", result.UserID, "userFingerprintID", result.UserFingerprintID, "userNfcID", result.UserNfcID, "roomID", result.RoomID, "roomSNSTopicARN", result.RoomSnsTopicARN, "sensorID", result.SensorID, "type", result.Type)
	fmt.Fprintf(w, "Access granted\n")
	go func() {
		if err := h.accessLogService.Create(context.TODO(), AccessLog{
			UserID:          result.UserID,
			RoomID:          result.RoomID,
			Method:          string(result.Type),
			IsGrantedAccess: true,
			Reason:          "Access granted",
		}); err != nil {
			slog.Error("Error creating access log", "error", err)
		}

		if result.RoomSnsTopicARN != nil && *result.RoomSnsTopicARN != "" {
			// 	_, err := SNSClient.Publish(context.TODO(), &sns.PublishInput{
			// 		Message:  &result.Username,
			// 		TopicArn: result.RoomSnsTopicARN,
			// 	})
			// 	if err != nil {
			// 		slog.Error("Error publishing to SNS", "error", err)
			// 	}
		}
	}()

}
