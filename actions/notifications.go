package actions

import (
	"strings"
	"time"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v5"
	"github.com/gofrs/uuid"
	"github.com/monarko/piia/helpers"
	"github.com/monarko/piia/models"
	"github.com/pkg/errors"
)

// NotificationsIndex returns the logs list
func NotificationsIndex(c buffalo.Context) error {
	loggedInUser, ok := c.Value("current_user").(*models.User)
	if !ok {
		return c.Redirect(302, "/users/login")
	}
	tx := c.Value("tx").(*pop.Connection)
	// Notifications
	notifications := &models.Notifications{}
	var q *pop.Query
	openNotificationStatuses := []string{"open", "nurse-notified", "patient-contacted", "referral-arranged"}
	if len(strings.TrimSpace(loggedInUser.Site)) > 0 {
		q = tx.Eager().Where("site = ?", loggedInUser.Site).Where("status in (?)", openNotificationStatuses).PaginateFromParams(c.Params()).Order("created_at DESC")
	} else if loggedInUser.Admin || loggedInUser.Permission.StudyCoordinator {
		q = tx.Eager().Where("status in (?)", openNotificationStatuses).PaginateFromParams(c.Params()).Order("created_at DESC")
	} else {
		q = tx.Eager().Where("status != ?", "unknown").PaginateFromParams(c.Params()).Order("created_at DESC")
	}

	// Retrieve all Notifications from the DB
	if err := q.All(notifications); err != nil {
		return errors.WithStack(err)
	}

	// Make posts available inside the html template
	c.Set("notifications", notifications)
	// Add the paginator to the context so it can be used in the template.
	c.Set("pagination", q.Paginator)

	b := c.Value("breadcrumb").(helpers.Breadcrumbs)
	b = append(b, helpers.Breadcrumb{Title: "section_header_notifications", Path: "/notifications/index"})
	c.Set("breadcrumb", b)

	return c.Render(200, r.HTML("notifications/index.html"))
}

// InsertNotification inserts a notification into db
func InsertNotification(notificationType, message, status, site string, fromUserID uuid.UUID, participantID uuid.UUID, screeningID uuid.UUID, c buffalo.Context) error {
	notification := &models.Notification{}
	oldNotification := notification.Maps()

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

	newNotification := notification.Maps()
	auditErr := MakeAudit("Notification", notification.ID, oldNotification, newNotification, fromUserID, c)
	if auditErr != nil {
		return errors.WithStack(auditErr)
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

// NotificationsDestroy function
func NotificationsDestroy(c buffalo.Context) error {
	returnURL := "/notifications/index"
	user := c.Value("current_user").(*models.User)
	if !user.Admin {
		c.Flash().Add("danger", "Access denied")
		return c.Redirect(302, returnURL)
	}

	tx := c.Value("tx").(*pop.Connection)
	notification := &models.Notification{}
	if err := tx.Eager().Find(notification, c.Param("nid")); err != nil {
		return c.Error(404, err)
	}

	reason := c.Request().FormValue("reason")

	err := ArchiveMake(c, user.ID, notification.ID, "Notification", notification, reason)
	if err != nil {
		c.Flash().Add("danger", err.Error())
		return c.Redirect(302, returnURL)
	}

	// If there are no errors set a flash message
	c.Flash().Add("success", "Archived successfully")

	return c.Redirect(302, returnURL)
}
