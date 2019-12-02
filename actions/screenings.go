package actions

import (
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/pop"
	"github.com/monarko/piia/models"
	"github.com/pkg/errors"
	"strings"
)

// ScreeningsIndex gets all Screenings. This function is mapped to the path
func ScreeningsIndex(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	participant := &models.Participant{}
	if err := tx.Find(participant, c.Param("pid")); err != nil {
		return c.Error(404, err)
	}
	c.Set("participant", participant)

	screenings := &models.Screenings{}
	q := tx.Eager("Screener").Where("participant_id = ?", c.Param("pid")).PaginateFromParams(c.Params()).Order("created_at DESC")
	// Retrieve all Screenings from the DB
	if err := q.All(screenings); err != nil {
		return errors.WithStack(err)
	}
	c.Set("screenings", screenings)

	// Add the paginator to the context so it can be used in the template.
	c.Set("pagination", q.Paginator)
	breadcrumbMap := make(map[string]interface{})
	breadcrumbMap["Participants"] = "/participants/index"
	// breadcrumbMap["Screenings"] = "/participants/" + c.Param("pid") + "/screenings/index"
	c.Set("breadcrumbMap", breadcrumbMap)
	return c.Render(200, r.HTML("screenings/index.html"))
}

func getDilatePupil(s models.Screening) string {
	returnText := ""
	if s.Eyes.LeftEye.DilatePupil.Valid && s.Eyes.RightEye.DilatePupil.Valid && s.Eyes.LeftEye.DilatePupil.Bool && s.Eyes.RightEye.DilatePupil.Bool {
		returnText = "both"
	} else if s.Eyes.LeftEye.DilatePupil.Valid && s.Eyes.RightEye.DilatePupil.Valid && !s.Eyes.LeftEye.DilatePupil.Bool && !s.Eyes.RightEye.DilatePupil.Bool {
		returnText = "no"
	} else if s.Eyes.RightEye.DilatePupil.Valid && s.Eyes.RightEye.DilatePupil.Bool {
		returnText = "right"
	} else if s.Eyes.LeftEye.DilatePupil.Valid && s.Eyes.LeftEye.DilatePupil.Bool {
		returnText = "left"
	}

	return returnText
}

// ScreeningsCreateGet renders the form for creating a new Screening.
func ScreeningsCreateGet(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	participant := &models.Participant{}
	if err := tx.Eager().Find(participant, c.Param("pid")); err != nil {
		return c.Error(404, err)
	}
	if len(participant.Screenings) > 0 {
		scr := participant.Screenings[0]
		red := "/participants/" + c.Param("pid") + "/screenings/edit/" + scr.ID.String()
		return c.Redirect(302, red)
	}
	hospitalNotReferralReasons := envy.Get("HOSPITAL_NOT_REFERRAL_REASONS", "")
	reasons := strings.SplitN(hospitalNotReferralReasons, ",", -1)
	c.Set("hospitalNotReferralReasons", reasons)
	hospitalNotReferralReasonsThai := envy.Get("HOSPITAL_NOT_REFERRAL_REASONS_THAI", "")
	reasonsThai := strings.SplitN(hospitalNotReferralReasonsThai, ",", -1)
	c.Set("hospitalNotReferralReasonsThai", reasonsThai)
	screening := &models.Screening{}
	c.Set("participant", participant)
	c.Set("screening", screening)
	c.Set("dilatePupil", getDilatePupil(*screening))
	breadcrumbMap := make(map[string]interface{})
	breadcrumbMap["Participants"] = "/participants/index"
	breadcrumbMap["New Screening"] = "/participants/" + c.Param("pid") + "/screenings/create"
	c.Set("breadcrumbMap", breadcrumbMap)
	return c.Render(200, r.HTML("screenings/create.html"))
}

// ScreeningsCreatePost renders the form for creating a new Screening.
func ScreeningsCreatePost(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	participant := &models.Participant{}
	if err := tx.Eager().Find(participant, c.Param("pid")); err != nil {
		return c.Error(404, err)
	}
	if len(participant.Screenings) > 0 {
		scr := participant.Screenings[0]
		red := "/participants/" + c.Param("pid") + "/screenings/edit/" + scr.ID.String()
		return c.Redirect(302, red)
	}
	hospitalNotReferralReasons := envy.Get("HOSPITAL_NOT_REFERRAL_REASONS", "")
	reasons := strings.SplitN(hospitalNotReferralReasons, ",", -1)
	c.Set("hospitalNotReferralReasons", reasons)
	hospitalNotReferralReasonsThai := envy.Get("HOSPITAL_NOT_REFERRAL_REASONS_THAI", "")
	reasonsThai := strings.SplitN(hospitalNotReferralReasonsThai, ",", -1)
	c.Set("hospitalNotReferralReasonsThai", reasonsThai)
	user := c.Value("current_user").(*models.User)
	screening := &models.Screening{}
	oldScreening := screening.Maps()
	if err := c.Bind(screening); err != nil {
		return errors.WithStack(err)
	}
	screening.Eyes.LeftEye.DilatePupil.Bool = false
	screening.Eyes.RightEye.DilatePupil.Bool = false
	screening.Eyes.LeftEye.DilatePupil.Valid = false
	screening.Eyes.RightEye.DilatePupil.Valid = false
	dilatePupil := c.Request().FormValue("dilatePupil")
	if dilatePupil == "both" {
		screening.Eyes.LeftEye.DilatePupil.Bool = true
		screening.Eyes.RightEye.DilatePupil.Bool = true
		screening.Eyes.LeftEye.DilatePupil.Valid = true
		screening.Eyes.RightEye.DilatePupil.Valid = true
	} else if dilatePupil == "left" {
		screening.Eyes.LeftEye.DilatePupil.Bool = true
		screening.Eyes.LeftEye.DilatePupil.Valid = true
	} else if dilatePupil == "right" {
		screening.Eyes.RightEye.DilatePupil.Bool = true
		screening.Eyes.RightEye.DilatePupil.Valid = true
	} else if dilatePupil == "no" {
		screening.Eyes.LeftEye.DilatePupil.Valid = true
		screening.Eyes.RightEye.DilatePupil.Valid = true
	}
	screening.ScreenerID = user.ID
	screening.ParticipantID = participant.ID

	screening.Referral.Referred.Bool = false
	screening.Referral.Referred.Valid = false
	screening.Referral.ReferralRefused.Bool = false
	screening.Referral.ReferralRefused.Valid = false
	referral := c.Request().FormValue("referral")
	if referral == "yes" {
		screening.Referral.Referred.Bool = true
		screening.Referral.Referred.Valid = true
		screening.Referral.ReferralRefused.Valid = true
	} else if referral == "no" {
		screening.Referral.Referred.Bool = false
		screening.Referral.Referred.Valid = true
	}
	hospitalReferral := c.Request().FormValue("hospital_referral")
	if hospitalReferral == "yes" {
		screening.Referral.HospitalReferred.Bool = true
		screening.Referral.HospitalReferred.Valid = true
	} else if hospitalReferral == "no" {
		screening.Referral.HospitalReferred.Bool = false
		screening.Referral.HospitalReferred.Valid = true
	}
	referralRefused := c.Request().FormValue("referral_refused")
	if referralRefused == "refused" {
		screening.Referral.ReferralRefused.Bool = true
		screening.Referral.ReferralRefused.Valid = true
	}
	hospitalNotReferredReason := c.Request().FormValue("hospital_not_referred_reason")
	hospitalNotReferredReasonText := c.Request().FormValue("hospital_not_referred_reason_text")
	if len(hospitalNotReferredReason) > 0 {
		screening.Referral.HospitalNotReferralReason.Valid = true
		if hospitalNotReferredReason == "Other" {
			screening.Referral.HospitalNotReferralReason.String = hospitalNotReferredReasonText
		} else {
			screening.Referral.HospitalNotReferralReason.String = hospitalNotReferredReason
			if strings.Contains(hospitalNotReferredReason, "refused") {
				screening.Referral.ReferralRefused.Bool = true
				screening.Referral.ReferralRefused.Valid = true
			}
		}
	}

	verrs, err := tx.ValidateAndCreate(screening)
	if err != nil {
		return errors.WithStack(err)
	}
	if verrs.HasAny() {
		c.Set("participant", participant)
		c.Set("screening", screening)
		c.Set("errors", verrs.Errors)
		c.Set("dilatePupil", getDilatePupil(*screening))
		breadcrumbMap := make(map[string]interface{})
		breadcrumbMap["Participants"] = "/participants/index"
		// breadcrumbMap["Screenings"] = "/participants/" + c.Param("pid") + "/screenings/index"
		breadcrumbMap["New Screening"] = "/participants/" + c.Param("pid") + "/screenings/create"
		c.Set("breadcrumbMap", breadcrumbMap)
		return c.Render(422, r.HTML("screenings/create.html"))
	}

	// if len(screening.Eyes.RightEye.VisualAcuity.String) > 0 && len(screening.Eyes.RightEye.DRGrading.String) > 0 && len(screening.Eyes.RightEye.DMEAssessment.String) > 0 && len(screening.Eyes.LeftEye.VisualAcuity.String) > 0 && len(screening.Eyes.LeftEye.DRGrading.String) > 0 && len(screening.Eyes.LeftEye.DMEAssessment.String) > 0 {
	if screening.Referral.Referred.Valid {
		participant.Status = "11"
		perrs, err := tx.ValidateAndUpdate(participant)
		if err != nil {
			return errors.WithStack(err)
		}
		if perrs.HasAny() {
			c.Set("participant", participant)
			c.Set("screening", screening)
			c.Set("errors", verrs.Errors)
			c.Set("dilatePupil", getDilatePupil(*screening))
			breadcrumbMap := make(map[string]interface{})
			breadcrumbMap["Participants"] = "/participants/index"
			// breadcrumbMap["Screenings"] = "/participants/" + c.Param("pid") + "/screenings/index"
			breadcrumbMap["New Screening"] = "/participants/" + c.Param("pid") + "/screenings/create"
			c.Set("breadcrumbMap", breadcrumbMap)
			return c.Render(422, r.HTML("screenings/create.html"))
		}
	}

	newScreening := screening.Maps()
	auditErr := MakeAudit("Screening", screening.ID, oldScreening, newScreening, user.ID, c)
	if auditErr != nil {
		return errors.WithStack(auditErr)
	}

	logErr := InsertLog("create", "User did a screening", "", screening.ID.String(), "screening", user.ID, c)
	if logErr != nil {
		return errors.WithStack(logErr)
	}

	// If there are no errors set a success message
	c.Flash().Add("success", "New screening added successfully.")

	return c.Redirect(302, "/participants/index")
}

// ScreeningsEditGet renders the form for creating a new Screening.
func ScreeningsEditGet(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	screening := &models.Screening{}
	if err := tx.Eager().Find(screening, c.Param("sid")); err != nil {
		return c.Error(404, err)
	}
	participant := screening.Participant
	if c.Param("pid") != participant.ID.String() {
		c.Flash().Add("danger", "Not Found")
		return c.Redirect(302, "/participants/index")
	}
	hospitalNotReferralReasons := envy.Get("HOSPITAL_NOT_REFERRAL_REASONS", "")
	reasons := strings.SplitN(hospitalNotReferralReasons, ",", -1)
	c.Set("hospitalNotReferralReasons", reasons)
	hospitalNotReferralReasonsThai := envy.Get("HOSPITAL_NOT_REFERRAL_REASONS_THAI", "")
	reasonsThai := strings.SplitN(hospitalNotReferralReasonsThai, ",", -1)
	c.Set("hospitalNotReferralReasonsThai", reasonsThai)
	c.Set("participant", participant)
	c.Set("screening", screening)
	c.Set("dilatePupil", getDilatePupil(*screening))
	// statuses := screening.StatusesMap()
	// c.Set("screeningStatuses", statuses)
	breadcrumbMap := make(map[string]interface{})
	breadcrumbMap["Participants"] = "/participants/index"
	breadcrumbMap["Edit Screening"] = "/participants/" + c.Param("pid") + "/screenings/edit"
	c.Set("breadcrumbMap", breadcrumbMap)
	return c.Render(200, r.HTML("screenings/edit.html"))
}

// ScreeningsEditPost renders the form for creating a new Screening.
func ScreeningsEditPost(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	user := c.Value("current_user").(*models.User)
	screening := &models.Screening{}
	if err := tx.Eager().Find(screening, c.Param("sid")); err != nil {
		return c.Error(404, err)
	}
	participant := screening.Participant
	if c.Param("pid") != participant.ID.String() {
		c.Flash().Add("danger", "Not Found")
		return c.Redirect(302, "/participants/index")
	}
	hospitalNotReferralReasons := envy.Get("HOSPITAL_NOT_REFERRAL_REASONS", "")
	reasons := strings.SplitN(hospitalNotReferralReasons, ",", -1)
	c.Set("hospitalNotReferralReasons", reasons)
	hospitalNotReferralReasonsThai := envy.Get("HOSPITAL_NOT_REFERRAL_REASONS_THAI", "")
	reasonsThai := strings.SplitN(hospitalNotReferralReasonsThai, ",", -1)
	c.Set("hospitalNotReferralReasonsThai", reasonsThai)
	oldScreening := screening.Maps()
	if err := c.Bind(screening); err != nil {
		return errors.WithStack(err)
	}

	screening.Eyes.LeftEye.DilatePupil.Bool = false
	screening.Eyes.RightEye.DilatePupil.Bool = false
	screening.Eyes.LeftEye.DilatePupil.Valid = false
	screening.Eyes.RightEye.DilatePupil.Valid = false
	dilatePupil := c.Request().FormValue("dilatePupil")
	if dilatePupil == "both" {
		screening.Eyes.LeftEye.DilatePupil.Bool = true
		screening.Eyes.RightEye.DilatePupil.Bool = true
		screening.Eyes.LeftEye.DilatePupil.Valid = true
		screening.Eyes.RightEye.DilatePupil.Valid = true
	} else if dilatePupil == "left" {
		screening.Eyes.LeftEye.DilatePupil.Bool = true
		screening.Eyes.LeftEye.DilatePupil.Valid = true
	} else if dilatePupil == "right" {
		screening.Eyes.RightEye.DilatePupil.Bool = true
		screening.Eyes.RightEye.DilatePupil.Valid = true
	} else if dilatePupil == "no" {
		screening.Eyes.LeftEye.DilatePupil.Valid = true
		screening.Eyes.RightEye.DilatePupil.Valid = true
	}

	screening.Referral.Referred.Bool = false
	screening.Referral.Referred.Valid = false
	screening.Referral.ReferralRefused.Bool = false
	screening.Referral.ReferralRefused.Valid = false
	referral := c.Request().FormValue("referral")
	if referral == "yes" {
		screening.Referral.Referred.Bool = true
		screening.Referral.Referred.Valid = true
		screening.Referral.ReferralRefused.Valid = true
	} else if referral == "no" {
		screening.Referral.Referred.Bool = false
		screening.Referral.Referred.Valid = true
	}
	hospitalReferral := c.Request().FormValue("hospital_referral")
	if hospitalReferral == "yes" {
		screening.Referral.HospitalReferred.Bool = true
		screening.Referral.HospitalReferred.Valid = true
	} else if hospitalReferral == "no" {
		screening.Referral.HospitalReferred.Bool = false
		screening.Referral.HospitalReferred.Valid = true
	}
	referralRefused := c.Request().FormValue("referral_refused")
	if referralRefused == "refused" {
		screening.Referral.ReferralRefused.Bool = true
		screening.Referral.ReferralRefused.Valid = true
	}
	hospitalNotReferredReason := c.Request().FormValue("hospital_not_referred_reason")
	hospitalNotReferredReasonText := c.Request().FormValue("hospital_not_referred_reason_text")
	if len(hospitalNotReferredReason) > 0 {
		screening.Referral.HospitalNotReferralReason.Valid = true
		if hospitalNotReferredReason == "Other" {
			screening.Referral.HospitalNotReferralReason.String = hospitalNotReferredReasonText
		} else {
			screening.Referral.HospitalNotReferralReason.String = hospitalNotReferredReason
			if strings.Contains(hospitalNotReferredReason, "refused") {
				screening.Referral.ReferralRefused.Bool = true
				screening.Referral.ReferralRefused.Valid = true
			}
		}
	}

	verrs, err := tx.ValidateAndUpdate(screening)
	if err != nil {
		return errors.WithStack(err)
	}
	if verrs.HasAny() {
		c.Set("participant", participant)
		c.Set("screening", screening)
		c.Set("errors", verrs.Errors)
		c.Set("dilatePupil", getDilatePupil(*screening))
		breadcrumbMap := make(map[string]interface{})
		breadcrumbMap["Participants"] = "/participants/index"
		breadcrumbMap["Edit Screening"] = "/participants/" + c.Param("pid") + "/screenings/edit"
		c.Set("breadcrumbMap", breadcrumbMap)
		return c.Render(422, r.HTML("screenings/edit.html"))
	}

	if participant.Status == "1" && screening.Referral.Referred.Valid {
		participant.Status = "11"
		perrs, err := tx.ValidateAndUpdate(&participant)
		if err != nil {
			return errors.WithStack(err)
		}
		if perrs.HasAny() {
			c.Set("participant", participant)
			c.Set("screening", screening)
			c.Set("errors", verrs.Errors)
			c.Set("dilatePupil", getDilatePupil(*screening))
			breadcrumbMap := make(map[string]interface{})
			breadcrumbMap["Participants"] = "/participants/index"
			breadcrumbMap["Edit Screening"] = "/participants/" + c.Param("pid") + "/screenings/edit"
			c.Set("breadcrumbMap", breadcrumbMap)
			return c.Render(422, r.HTML("screenings/edit.html"))
		}
	}

	newScreening := screening.Maps()
	auditErr := MakeAudit("Screening", screening.ID, oldScreening, newScreening, user.ID, c)
	if auditErr != nil {
		return errors.WithStack(auditErr)
	}

	logErr := InsertLog("update", "User updated a screening", "", screening.ID.String(), "screening", user.ID, c)
	if logErr != nil {
		return errors.WithStack(logErr)
	}

	// If there are no errors set a success message
	c.Flash().Add("success", "Screening updated successfully.")

	return c.Redirect(302, "/participants/index")
}

// ScreeningsDestroy renders the form for creating a new Screening.
func ScreeningsDestroy(c buffalo.Context) error {
	returnURL := "/participants/" + c.Param("pid")
	user := c.Value("current_user").(*models.User)
	if !user.Admin {
		c.Flash().Add("danger", "Access denied")
		return c.Redirect(302, returnURL)
	}

	tx := c.Value("tx").(*pop.Connection)
	screening := &models.Screening{}
	if err := tx.Eager().Find(screening, c.Param("sid")); err != nil {
		return c.Error(404, err)
	}
	participant := screening.Participant
	if c.Param("pid") != participant.ID.String() {
		c.Flash().Add("danger", "Not Found")
		return c.Redirect(302, "/participants/index")
	}

	reason := c.Request().FormValue("reason")

	for _, o := range screening.OverReadings {
		err := ArchiveMake(c, user.ID, o.ID, "OverReading", &o, reason)
		if err != nil {
			c.Flash().Add("danger", err.Error())
			return c.Redirect(302, returnURL)
		}
	}

	for _, o := range screening.Notifications {
		err := ArchiveMake(c, user.ID, o.ID, "Notification", &o, reason)
		if err != nil {
			c.Flash().Add("danger", err.Error())
			return c.Redirect(302, returnURL)
		}
	}

	for _, o := range screening.ReferredMessages {
		err := ArchiveMake(c, user.ID, o.ID, "ReferredMessage", &o, reason)
		if err != nil {
			c.Flash().Add("danger", err.Error())
			return c.Redirect(302, returnURL)
		}
	}

	err := ArchiveMake(c, user.ID, screening.ID, "Screening", screening, reason)
	if err != nil {
		c.Flash().Add("danger", err.Error())
		return c.Redirect(302, returnURL)
	}

	prt := &models.Participant{}
	if err := tx.Find(prt, participant.ID); err != nil {
		return c.Error(404, err)
	}
	prt.Status = "1"
	perrs, err := tx.ValidateAndUpdate(prt)
	if err != nil {
		c.Flash().Add("danger", err.Error())
		return c.Redirect(302, returnURL)
	}
	if perrs.HasAny() {
		c.Set("errors", perrs.Errors)
		return c.Redirect(302, returnURL)
	}

	// If there are no errors set a flash message
	c.Flash().Add("success", "Archived successfully")

	return c.Redirect(302, returnURL)
}
