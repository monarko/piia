package actions

import (
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop"
	"github.com/monarko/piia/models"
	"github.com/pkg/errors"
)

// UpdateReferredMessage changes the site language
func UpdateReferredMessage(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	refer := &models.ReferredMessage{}
	user := c.Value("current_user").(*models.User)
	refer.UserID = user.ID

	participant := &models.Participant{}
	if err := tx.Eager("Screenings").Find(participant, c.Param("pid")); err != nil {
		return c.Error(404, err)
	}
	refer.ParticipantID = participant.ID
	refer.ScreeningID = participant.Screenings[0].ID
	refer.Message = c.Request().FormValue("Message")

	verrs, err := tx.ValidateAndCreate(refer)
	if err != nil {
		return errors.WithStack(err)
	}

	if verrs.HasAny() {
		c.Set("errors", verrs.Errors)
	} else {
		logErr := InsertLog("create", "User created a referral message: "+c.Request().FormValue("message"), "", refer.ID.String(), "referred_message", user.ID, c)
		if logErr != nil {
			return errors.WithStack(logErr)
		}
	}

	referrer := c.Request().Referer()

	return c.Redirect(302, referrer)
}
