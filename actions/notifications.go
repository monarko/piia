package actions

import (
	"strings"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/uuid"
	"github.com/monarko/piia/models"
	"github.com/pkg/errors"
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

// ChangeNotificationStatus changes the site language
func ChangeNotificationStatus(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	notification := &models.Notification{}
	if err := tx.Find(notification, string(c.Request().FormValue("notificationId"))); err != nil {
		return c.Error(404, err)
	}
	user := c.Value("current_user").(*models.User)

	notification.Status = c.Request().FormValue("status")

	verrs, err := tx.ValidateAndUpdate(notification)
	if err != nil {
		return errors.WithStack(err)
	}

	if verrs.HasAny() {
		c.Set("errors", verrs.Errors)
	} else {
		logErr := InsertLog("update", "User updated the notification with status: "+strings.ToUpper(c.Request().FormValue("status")), "", notification.ID.String(), "notification", user.ID, c)
		if logErr != nil {
			return errors.WithStack(logErr)
		}
	}

	referrer := c.Request().Referer()

	return c.Redirect(302, referrer)
}
