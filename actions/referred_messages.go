package actions

import (
	"time"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop"
	"github.com/monarko/piia/models"
	"github.com/pkg/errors"
)

// UpdateReferredMessage changes the site language
func UpdateReferredMessage(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	refer := &models.ReferredMessage{}
	oldRefer := refer.Maps()
	user := c.Value("current_user").(*models.User)
	refer.UserID = user.ID

	participant := &models.Participant{}
	if err := tx.Eager("Screenings").Find(participant, c.Param("pid")); err != nil {
		return c.Error(404, err)
	}
	refer.ParticipantID = participant.ID
	refer.ScreeningID = participant.Screenings[0].ID
	refer.Message = c.Request().FormValue("Message")

	if c.Request().FormValue("Attended") == "yes" {
		refer.ReferralData.Attended = true
	} else {
		refer.ReferralData.Attended = false
	}

	if c.Request().FormValue("Treatment") == "yes" {
		refer.ReferralData.ReferredForTreatment = true
	} else {
		refer.ReferralData.ReferredForTreatment = false
	}

	err := c.Request().ParseForm()
	if err != nil {
		return c.Error(404, err)
	}
	refer.ReferralData.Plans = c.Request().Form["Plans"]
	refer.ReferralData.FollowUpPlan = c.Request().FormValue("FollowUp")
	refer.ReferralData.DateOfAttendance.Calendar = c.Request().FormValue("Calendar")
	refer.ReferralData.DateOfAttendance.GivenDate, _ = time.Parse("2006-01-02", c.Request().FormValue("GivenDate"))
	if len(c.Request().FormValue("HospitalName")) > 0 {
		refer.ReferralData.HospitalName.String = c.Request().FormValue("HospitalName")
		refer.ReferralData.HospitalName.Valid = true
	}

	// fmt.Printf("\n\n%#v\n\n", c.Request().FormValue("HospitalName"))

	verrs, err := tx.ValidateAndCreate(refer)
	if err != nil {
		return errors.WithStack(err)
	}

	if verrs.HasAny() {
		c.Set("errors", verrs.Errors)
	} else {
		newRefer := refer.Maps()
		auditErr := MakeAudit("ReferredMessage", refer.ID, oldRefer, newRefer, user.ID, c)
		if auditErr != nil {
			return errors.WithStack(auditErr)
		}

		logErr := InsertLog("create", "User created a referral message: "+c.Request().FormValue("message"), "", refer.ID.String(), "referred_message", user.ID, c)
		if logErr != nil {
			return errors.WithStack(logErr)
		}

		logErr = InsertLog("update", "User marked a participant for appointment completed", "", participant.ID.String(), "participant", user.ID, c)
		if logErr != nil {
			return errors.WithStack(logErr)
		}
	}

	return c.Redirect(302, "/referrals/index")
}
