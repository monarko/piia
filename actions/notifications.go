package actions

import (
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/uuid"
	"github.com/monarko/piia/models"
)

// NotificationsIndex returns the logs list
func NotificationsIndex(c buffalo.Context) error {
	return nil
}

// InsertNotification inserts a notification into db
func InsertNotification(notificationType, message, status, site string, fromUserID uuid.UUID, participantID uuid.UUID, screeningID uuid.UUID, c buffalo.Context) error {
	notification := &models.Notification{}

	notification.Type = notificationType
	notification.Message = message
	notification.Status = status
	notification.Site = site

	notification.FromUserID = fromUserID
	notification.ParticipantID = participantID
	notification.ScreeningID = screeningID

	tx := c.Value("tx").(*pop.Connection)
	_, err := tx.ValidateAndCreate(notification)
	if err != nil {
		return err
	}

	return nil
}
