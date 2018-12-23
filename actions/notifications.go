package actions

import (
	"strings"
	"time"

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
	oldNotification := notification.Maps()
	user := c.Value("current_user").(*models.User)

	notification.Status = c.Request().FormValue("status")
	notification.Message = c.Request().FormValue("notes")

	eventDate := c.Request().FormValue("eventDate")
	eventDateLaguage := c.Request().FormValue("eventDateLanguage")

	notification.EventDate.Calendar = eventDateLaguage
	notification.EventDate.GivenDate, _ = time.Parse("2006-01-02", eventDate)
	notification.EventDate.CalculatedDate = notification.EventDate.GivenDate
	if notification.EventDate.Calendar == "thai" {
		notification.EventDate.CalculatedDate = notification.EventDate.CalculatedDate.AddDate(-543, 0, 0)
	}

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

	newNotification := notification.Maps()
	auditErr := MakeAudit("Notification", notification.ID, oldNotification, newNotification, user.ID, c)
	if auditErr != nil {
		return errors.WithStack(auditErr)
	}

	referrer := c.Request().Referer()

	return c.Redirect(302, referrer)
}
