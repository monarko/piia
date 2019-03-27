package actions

import (
	"bytes"
	"context"
	"encoding/csv"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/gobuffalo/pop/nulls"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/pop"
	"github.com/monarko/piia/models"
)

// AnalyticsIndex default implementation.
func AnalyticsIndex(c buffalo.Context) error {
	download := false
	downloadType := "csv"
	dlValue := c.Params().Get("download")
	if dlValue == "Download CSV" {
		download = true
	} else if dlValue == "Download CSV for Veil" {
		download = true
		downloadType = "veil"
	}

	tx := c.Value("tx").(*pop.Connection)
	user := c.Value("current_user").(*models.User)
	analyticsScreenings := &models.AnalyticsScreenings{}

	query := `SELECT 
	s.created_at AS "CreatedDate",
	p.participant_id AS "ParticipantID",
	date_part('year', age(((p.dob->>'calculated_date'::text)::date)::timestamp with time zone)) AS "Age",
        CASE
            WHEN ((p.gender)::text = 'M'::text) THEN 'Male'::text
            ELSE 'Female'::text
        END AS "Gender",
	((s.eye->'right'::text)->>'visual_acuity'::text) AS "VAOD",
	((s.eye->'right'::text)->>'last_visual_acuity'::text) AS "VAPreviousOD",
	((s.eye->'right'::text)->>'dr'::text) AS "DRGradeOD",
	((s.eye->'right'::text)->>'dme'::text) AS "DMEOD",
	((s.eye->'left'::text)->>'visual_acuity'::text) AS "VAOS",
	((s.eye->'left'::text)->>'last_visual_acuity'::text) AS "VAPreviousOS",
	((s.eye->'left'::text)->>'dr'::text) AS "DRGradeOS",
	((s.eye->'left'::text)->>'dme'::text) AS "DMEOS",
        CASE
            WHEN ((s.referral->>'referred'::text) = 'true'::text) THEN 'Yes'::text
            ELSE 'No'::text
        END AS "DrReferral",
	(s.referral->>'additional_notes'::text) AS "ReferralNotes",
	((o.eye_assessment->'left'::text)->>'dr'::text) AS "DRGradeOVLeft",
	((o.eye_assessment->'left'::text)->>'dme'::text) AS "DMEOVLeft",
	((o.eye_assessment->'left'::text)->>'suspected_pathologies'::text) AS "SuspectedLeft",
	((o.eye_assessment->'right'::text)->>'dr'::text) AS "DRGradeOVRight",
	((o.eye_assessment->'right'::text)->>'dme'::text) AS "DMEOVRight",
	((o.eye_assessment->'right'::text)->>'suspected_pathologies'::text) AS "SuspectedRight",
        CASE
            WHEN ((o.referral->>'referred'::text) = 'true'::text) THEN 'Yes'::text
            ELSE 'No'::text
        END AS "OVReferral",
	(o.referral->>'additional_notes'::text) AS "OVReferralNotes"
FROM (
	participants p
	LEFT JOIN screenings s ON (p.id = s.participant_id)
	LEFT JOIN over_readings o ON (o.screening_id = s.id)
)
WHERE (
	s.id IS NOT NULL
)
ORDER BY s.created_at DESC`

	//  AND right(left(p.participant_id, 2), 1) = ?

	var q *pop.Query

	if download {
		q = tx.RawQuery(query)
	} else {
		q = tx.RawQuery(query).PaginateFromParams(c.Params())
	}

	// Retrieve all Posts from the DB
	if err := q.All(analyticsScreenings); err != nil {
		// return errors.WithStack(err)
		errStr := err.Error()
		errs := map[string][]string{
			"index_error": {errStr},
		}
		c.Set("errors", errs)
		if download {
			InsertLog("error", "User download analytics error", err.Error(), "", "", user.ID, c)
		} else {
			InsertLog("error", "User viewed analytics error", err.Error(), "", "", user.ID, c)
		}
		return c.Redirect(302, "/")
	}

	if download {
		var b *bytes.Buffer
		var err error
		filenamePart := "-analytics-"

		if downloadType == "veil" {
			b, err = downloadVeil(*analyticsScreenings)
			filenamePart = "-veil-"
		} else {
			b, err = downloadAnalytics(*analyticsScreenings)
		}

		if err != nil {
			errStr := err.Error()
			errs := map[string][]string{
				"index_error": {errStr},
			}
			c.Set("errors", errs)
			InsertLog("error", "User download analytics error", err.Error(), "", "", user.ID, c)
			return c.Redirect(302, "/")
		}

		InsertLog("error", "User download analytics", "", "", "", user.ID, c)
		appHost := envy.Get("APP_HOST", "http://127.0.0.1")
		hosts := strings.Split(strings.TrimSpace(strings.Replace(appHost, "/", "", -1)), ":")
		filename := hosts[1] + filenamePart + time.Now().Format("2006-01-02T15-04-05-0700") + ".csv"

		return c.Render(200, r.Download(c, filename, b))
	}

	// Make posts available inside the html template
	c.Set("participants", analyticsScreenings)
	// Add the paginator to the context so it can be used in the template.
	c.Set("pagination", q.Paginator)

	breadcrumbMap := make(map[string]interface{})
	breadcrumbMap["page_analytics_title"] = "/analytics/index"
	c.Set("breadcrumbMap", breadcrumbMap)
	c.Set("filterStatus", c.Params().Get("status"))
	logErr := InsertLog("view", "User viewed analytics", "", "", "", user.ID, c)
	if logErr != nil {
		// return errors.WithStack(logErr)
		errStr := logErr.Error()
		errs := map[string][]string{
			"index_error": {errStr},
		}
		c.Set("errors", errs)
		InsertLog("error", "User viewed analytics error", logErr.Error(), "", "", user.ID, c)
		return c.Render(422, r.HTML("analytics/index.html"))
	}
	return c.Render(200, r.HTML("analytics/index.html"))
}

func downloadAnalytics(analytics []models.AnalyticsScreening) (*bytes.Buffer, error) {
	b := &bytes.Buffer{}
	w := csv.NewWriter(b)

	headers := []string{
		"Created Date",
		"Participant ID",
		"Age (Year)",
		"Gender",
		"VA (OD)",
		"VAPrevious (OD)",
		"DRGrade (OD)",
		"DME (OD)",
		"VA (OS)",
		"VAPrevious (OS)",
		"DRGrade (OS)",
		"DME (OS)",
		"DrReferral",
		"ReferralNotes",
		"DRGrade (Right) (Over Reader)",
		"DME (Right) (Over Reader)",
		"Suspected Pathologies (Right) (Over Reader)",
		"DRGrade (Left) (Over Reader)",
		"DME (Left) (Over Reader)",
		"Suspected Pathologies (Left) (Over Reader)",
		"Over Reader Referral",
		"Over Reader Notes",
	}

	if err := w.Write(headers); err != nil {
		return nil, err
	}

	for _, a := range analytics {
		var record []string
		record = append(record, a.CreatedDate.Format(time.RFC3339))
		record = append(record, a.ParticipantID)
		record = append(record, strconv.FormatInt(int64(a.Age), 10))
		record = append(record, a.Gender)
		record = append(record, a.VAOD.String)
		record = append(record, a.VAPreviousOD.String)
		record = append(record, a.DRGradeOD.String)
		record = append(record, a.DMEOD.String)
		record = append(record, a.VAOS.String)
		record = append(record, a.VAPreviousOS.String)
		record = append(record, a.DRGradeOS.String)
		record = append(record, a.DMEOS.String)
		record = append(record, a.DrReferral.String)
		record = append(record, a.ReferralNotes.String)
		record = append(record, a.DRGradeOVRight.String)
		record = append(record, a.DMEOVRight.String)
		sr := ""
		if len(a.SuspectedRight.String) > 0 {
			sr = a.SuspectedRight.String
		}
		record = append(record, SliceStringToCommaSeparatedValue(sr))
		record = append(record, a.DRGradeOVLeft.String)
		record = append(record, a.DMEOVLeft.String)
		sl := ""
		if len(a.SuspectedLeft.String) > 0 {
			sl = a.SuspectedLeft.String
		}
		record = append(record, SliceStringToCommaSeparatedValue(sl))
		record = append(record, a.OVReferral.String)
		record = append(record, a.OVReferralNotes.String)

		if err := w.Write(record); err != nil {
			return nil, err
		}
	}
	w.Flush()

	if err := w.Error(); err != nil {
		return nil, err
	}

	return b, nil
}

func downloadVeil(analytics []models.AnalyticsScreening) (*bytes.Buffer, error) {
	b := &bytes.Buffer{}
	w := csv.NewWriter(b)

	headers := []string{
		"StudyId",
		"Status",
		"LeftGradable",
		"RightGradable",
		"LeftDrGrade",
		"LeftDmeGrade",
		"RightDrGrade",
		"RightDmeGrade",
		"Referrable",
	}

	if err := w.Write(headers); err != nil {
		return nil, err
	}

	for _, a := range analytics {
		var record []string
		// StudyId / ParticipantId
		record = append(record, strings.Replace(a.ParticipantID, "-", "", -1))

		// Status
		if len(a.DRGradeOS.String) > 0 && len(a.DRGradeOD.String) > 0 {
			record = append(record, "COMPLETED")
		} else {
			record = append(record, "INCOMPLETE")
		}

		// LeftGradable
		if a.DRGradeOS.String == "Ungradeable" || a.DMEOS.String == "Ungradeable" {
			record = append(record, "FALSE")
		} else {
			record = append(record, "TRUE")
		}

		// RightGradable
		if a.DRGradeOD.String == "Ungradeable" || a.DMEOD.String == "Ungradeable" {
			record = append(record, "FALSE")
		} else {
			record = append(record, "TRUE")
		}

		// LeftDrGrade
		if a.DRGradeOS.String == "Normal" {
			record = append(record, "NO")
		} else {
			temp := strings.ToUpper(strings.TrimSuffix(a.DRGradeOS.String, " DR"))
			record = append(record, temp)
		}

		// LeftDmeGrade
		if a.DMEOS.String == "Not Present" {
			record = append(record, "NO")
		} else if a.DMEOS.String == "Present" {
			record = append(record, "YES")
		} else {
			record = append(record, strings.ToUpper(a.DMEOS.String))
		}

		// RightDrGrade
		if a.DRGradeOD.String == "Normal" {
			record = append(record, "NO")
		} else {
			temp := strings.ToUpper(strings.TrimSuffix(a.DRGradeOD.String, " DR"))
			record = append(record, temp)
		}

		// RightDmeGrade
		if a.DMEOD.String == "Not Present" {
			record = append(record, "NO")
		} else if a.DMEOD.String == "Present" {
			record = append(record, "YES")
		} else {
			record = append(record, strings.ToUpper(a.DMEOD.String))
		}

		// Referrable
		if a.DrReferral.String == "Yes" {
			record = append(record, "TRUE")
		} else {
			record = append(record, "FALSE")
		}

		if err := w.Write(record); err != nil {
			return nil, err
		}
	}
	w.Flush()

	if err := w.Error(); err != nil {
		return nil, err
	}

	return b, nil
}

type fullRecord struct {
	CreatedDate                     time.Time    `json:"created_date" db:"created_date"`
	ParticipantID                   string       `json:"participant_id" db:"participant_id"`
	Age                             int          `json:"age" db:"age"`
	Gender                          string       `json:"gender" db:"gender"`
	DiabetesType                    nulls.String `json:"diabetes_type" db:"diabetes_type"`
	DiabetesDuration                nulls.String `json:"diabetes_duration" db:"diabetes_duration"`
	DiabetesDurationType            nulls.String `json:"diabetes_duration_type" db:"diabetes_duration_type"`
	MedicalHistorySmoker            nulls.String `json:"medical_history_smoker" db:"medical_history_smoker"`
	MedicalHistoryMorbidities       nulls.String `json:"medical_history_morbidities" db:"medical_history_morbidities"`
	MedicationsOnInsulin            nulls.String `json:"medications_on_insulin" db:"medications_on_insulin"`
	MedicationsTakingMedications    nulls.String `json:"medications_taking_medications" db:"medications_taking_medications"`
	MeasurementsBPSBP               nulls.String `json:"measurements_bp_sbp" db:"measurements_bp_sbp"`
	MeasurementsBPDBP               nulls.String `json:"measurements_bp_dbp" db:"measurements_bp_dbp"`
	MeasurementsBPAssessmentSate    nulls.String `json:"measurements_bp_assessment_date" db:"measurements_bp_assessment_date"`
	PathologyHba1cValue             nulls.String `json:"pathology_hba1c_value" db:"pathology_hba1c_value"`
	PathologyHba1cUnit              nulls.String `json:"pathology_hba1c_unit" db:"pathology_hba1c_unit"`
	PathologyHba1cAssessmentDate    nulls.String `json:"pathology_hba1c_assessment_date" db:"pathology_hba1c_assessment_date"`
	PathologyLipidsTotalCholesterol nulls.String `json:"pathology_lipids_total_cholesterol" db:"pathology_lipids_total_cholesterol"`
	PathologyLipidsUnit             nulls.String `json:"pathology_lipids_unit" db:"pathology_lipids_unit"`
	PathologyLipidsAssessmentDate   nulls.String `json:"pathology_lipids_assessment_date" db:"pathology_lipids_assessment_date"`
	RightVisualAcuity               nulls.String `json:"right_visual_acuity" db:"right_visual_acuity"`
	RightPreviousVisualAcuity       nulls.String `json:"right_previous_visual_acuity" db:"right_previous_visual_acuity"`
	RightDRGrade                    nulls.String `json:"right_dr_grade" db:"right_dr_grade"`
	RightDME                        nulls.String `json:"right_dme" db:"right_dme"`
	LeftVisualAcuity                nulls.String `json:"left_visual_acuity" db:"left_visual_acuity"`
	LeftPreviousVisualAcuity        nulls.String `json:"left_previous_visual_acuity" db:"left_previous_visual_acuity"`
	LeftDRGrade                     nulls.String `json:"left_dr_grade" db:"left_dr_grade"`
	LeftDME                         nulls.String `json:"left_dme" db:"left_dme"`
	DrReferral                      nulls.String `json:"dr_referral" db:"dr_referral"`
	DrReferralNotes                 nulls.String `json:"dr_referral_notes" db:"dr_referral_notes"`
	RightDRGradeOver                nulls.String `json:"right_dr_grade_over" db:"right_dr_grade_over"`
	RightDMEOver                    nulls.String `json:"right_dme_over" db:"right_dme_over"`
	RightSuspectedOver              nulls.String `json:"right_suspected_over" db:"right_suspected_over"`
	LeftDRGradeOver                 nulls.String `json:"left_dr_grade_over" db:"left_dr_grade_over"`
	LeftDMEOver                     nulls.String `json:"left_dme_over" db:"left_dme_over"`
	LeftSuspectedOver               nulls.String `json:"left_suspected_over" db:"left_suspected_over"`
	OverReferral                    nulls.String `json:"over_referral" db:"over_referral"`
	OverReferralNotes               nulls.String `json:"over_referral_notes" db:"over_referral_notes"`
}

// DownloadFull function
func DownloadFull(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	user := c.Value("current_user").(*models.User)

	fullRecords := make([]fullRecord, 0)

	query := `SELECT
	s.created_at AS "created_date",
	p.participant_id AS "participant_id",
	date_part('year', age(((p.dob->>'calculated_date'::text)::date)::timestamp with time zone)) AS "age",
        CASE
            WHEN ((p.gender)::text = 'M'::text) THEN 'Male'::text
            ELSE 'Female'::text
        END AS "gender",

	(s.diabetes->>'diabetes_type'::text) AS "diabetes_type",
	(s.diabetes->>'duration'::text) AS "diabetes_duration",
	(s.diabetes->>'duration_type'::text) AS "diabetes_duration_type",

	CASE
            WHEN ((s.medical_history->>'smoker'::text) = 'true'::text) THEN 'Yes'::text
            ELSE 'No'::text
        END AS "medical_history_smoker",
	(s.medical_history->>'morbidities'::text) AS "medical_history_morbidities",

	CASE
            WHEN ((s.medications->>'on_insulin'::text) = 'true'::text) THEN 'Yes'::text
            ELSE 'No'::text
        END AS "medications_on_insulin",
	CASE
            WHEN ((s.medications->>'taking_medications'::text) = 'true'::text) THEN 'Yes'::text
            ELSE 'No'::text
        END AS "medications_taking_medications",

	((s.measurements->'blood_pressure'::text)->>'sbp'::text) AS measurements_bp_sbp,
	((s.measurements->'blood_pressure'::text)->>'dbp'::text) AS measurements_bp_dbp,
	CASE 
	    WHEN (LEFT((((s.measurements->'blood_pressure'::text)->'assessment_date'::text)->>'calculated_date'::text), 10) = '0001-01-01'::text) THEN NULL
	    ELSE LEFT((((s.measurements->'blood_pressure'::text)->'assessment_date'::text)->>'calculated_date'::text), 10)
	END AS measurements_bp_assessment_date,

	((s.pathology->'hba1c'::text)->>'value'::text) AS pathology_hba1c_value,
	((s.pathology->'hba1c'::text)->>'unit'::text) AS pathology_hba1c_unit,
	CASE 
	    WHEN (LEFT((((s.pathology->'hba1c'::text)->'assessment_date'::text)->>'calculated_date'::text), 10) = '0001-01-01'::text) THEN NULL
	    ELSE LEFT((((s.pathology->'hba1c'::text)->'assessment_date'::text)->>'calculated_date'::text), 10)
	END AS pathology_hba1c_assessment_date,

	((s.pathology->'lipids'::text)->>'total_cholesterol'::text) AS pathology_lipids_total_cholesterol,
	((s.pathology->'lipids'::text)->>'unit'::text) AS pathology_lipids_unit,
	CASE 
	    WHEN (LEFT((((s.pathology->'lipids'::text)->'assessment_date'::text)->>'calculated_date'::text), 10) = '0001-01-01'::text) THEN NULL
	    ELSE LEFT((((s.pathology->'lipids'::text)->'assessment_date'::text)->>'calculated_date'::text), 10)
	END AS pathology_lipids_assessment_date,

	((s.eye->'right'::text)->>'visual_acuity'::text) AS "right_visual_acuity",
	((s.eye->'right'::text)->>'last_visual_acuity'::text) AS "right_previous_visual_acuity",
	((s.eye->'right'::text)->>'dr'::text) AS "right_dr_grade",
	((s.eye->'right'::text)->>'dme'::text) AS "right_dme",
	((s.eye->'left'::text)->>'visual_acuity'::text) AS "left_visual_acuity",
	((s.eye->'left'::text)->>'last_visual_acuity'::text) AS "left_previous_visual_acuity",
	((s.eye->'left'::text)->>'dr'::text) AS "left_dr_grade",
	((s.eye->'left'::text)->>'dme'::text) AS "left_dme",
        CASE
            WHEN ((s.referral->>'referred'::text) = 'true'::text) THEN 'Yes'::text
            ELSE 'No'::text
        END AS "dr_referral",
	(s.referral->>'additional_notes'::text) AS "dr_referral_notes",
	((o.eye_assessment->'right'::text)->>'dr'::text) AS "right_dr_grade_over",
	((o.eye_assessment->'right'::text)->>'dme'::text) AS "right_dme_over",
	((o.eye_assessment->'right'::text)->>'suspected_pathologies'::text) AS "right_suspected_over",
	((o.eye_assessment->'left'::text)->>'dr'::text) AS "left_dr_grade_over",
	((o.eye_assessment->'left'::text)->>'dme'::text) AS "left_dme_over",
	((o.eye_assessment->'left'::text)->>'suspected_pathologies'::text) AS "left_suspected_over",
        CASE
            WHEN ((o.referral->>'referred'::text) = 'true'::text) THEN 'Yes'::text
            ELSE 'No'::text
        END AS "over_referral",
	(o.referral->>'additional_notes'::text) AS "over_referral_notes"
FROM (
	participants p
	LEFT JOIN screenings s ON (p.id = s.participant_id)
	LEFT JOIN over_readings o ON (o.screening_id = s.id)
)
WHERE (
	s.id IS NOT NULL
)
ORDER BY s.created_at`

	q := tx.RawQuery(query)

	if err := q.All(&fullRecords); err != nil {
		errStr := err.Error()
		errs := map[string][]string{
			"index_error": {errStr},
		}
		c.Set("errors", errs)
		InsertLog("error", "User download analytics error", err.Error(), "", "", user.ID, c)
		return c.Redirect(302, "/analytics/index")
	}

	appHost := envy.Get("APP_HOST", "http://127.0.0.1")
	hosts := strings.Split(strings.TrimSpace(strings.Replace(appHost, "/", "", -1)), ":")
	// filename := hosts[1] + "-full-record-" + time.Now().Format("2006-01-02T15-04-05") + ".csv"
	filename := hosts[1] + "-full-record" + ".csv"

	b, err := downloadAllRecords(fullRecords)
	if err != nil {
		errStr := err.Error()
		errs := map[string][]string{
			"index_error": {errStr},
		}
		c.Set("errors", errs)
		InsertLog("error", "User download analytics error", err.Error(), "", "", user.ID, c)
		return c.Redirect(302, "/analytics/index")
	}

	err = storeToGoogleCloudStorage(filename, b)
	if err != nil {
		errStr := err.Error()
		errs := map[string][]string{
			"index_error": {errStr},
		}
		c.Set("errors", errs)
		c.Flash().Add("danger", "Error from GCS: "+errStr)
		InsertLog("error", "User download analytics error", err.Error(), "", "", user.ID, c)
	} else {
		c.Flash().Add("success", "Successfully saved to your storage bucket")
	}

	return c.Redirect(302, "/analytics/index")
}

func downloadAllRecords(records []fullRecord) (*bytes.Buffer, error) {
	b := &bytes.Buffer{}
	w := csv.NewWriter(b)

	headers := []string{
		"created_date",
		"participant_id",
		"age",
		"gender",
		"diabetes_type",
		"diabetes_duration",
		"diabetes_duration_type",
		"medical_history_smoker",
		"medical_history_morbidities",
		"medications_on_insulin",
		"medications_taking_medications",
		"measurements_bp_sbp",
		"measurements_bp_dbp",
		"measurements_bp_assessment_date",
		"pathology_hba1c_value",
		"pathology_hba1c_unit",
		"pathology_hba1c_assessment_date",
		"pathology_lipids_total_cholesterol",
		"pathology_lipids_unit",
		"pathology_lipids_assessment_date",
		"right_visual_acuity",
		"right_previous_visual_acuity",
		"right_dr_grade",
		"right_dme",
		"left_visual_acuity",
		"left_previous_visual_acuity",
		"left_dr_grade",
		"left_dme",
		"dr_referral",
		"dr_referral_notes",
		"right_dr_grade_over",
		"right_dme_over",
		"right_suspected_over",
		"left_dr_grade_over",
		"left_dme_over",
		"left_suspected_over",
		"over_referral",
		"over_referral_notes",
	}

	if err := w.Write(headers); err != nil {
		return nil, err
	}

	for _, a := range records {
		var rc []string

		rc = append(rc, a.CreatedDate.Format(time.RFC3339))
		rc = append(rc, a.ParticipantID)
		rc = append(rc, strconv.FormatInt(int64(a.Age), 10))
		rc = append(rc, a.Gender)
		rc = append(rc, a.DiabetesType.String)
		rc = append(rc, a.DiabetesDuration.String)
		rc = append(rc, a.DiabetesDurationType.String)
		rc = append(rc, a.MedicalHistorySmoker.String)
		sr := ""
		if len(a.MedicalHistoryMorbidities.String) > 0 {
			sr = a.MedicalHistoryMorbidities.String
		}
		rc = append(rc, SliceStringToCommaSeparatedValue(sr))
		rc = append(rc, a.MedicationsOnInsulin.String)
		rc = append(rc, a.MedicationsTakingMedications.String)
		rc = append(rc, a.MeasurementsBPSBP.String)
		rc = append(rc, a.MeasurementsBPDBP.String)
		rc = append(rc, a.MeasurementsBPAssessmentSate.String)
		rc = append(rc, a.PathologyHba1cValue.String)
		rc = append(rc, a.PathologyHba1cUnit.String)
		rc = append(rc, a.PathologyHba1cAssessmentDate.String)
		rc = append(rc, a.PathologyLipidsTotalCholesterol.String)
		rc = append(rc, a.PathologyLipidsUnit.String)
		rc = append(rc, a.PathologyLipidsAssessmentDate.String)
		rc = append(rc, a.RightVisualAcuity.String)
		rc = append(rc, a.RightPreviousVisualAcuity.String)
		rc = append(rc, a.RightDRGrade.String)
		rc = append(rc, a.RightDME.String)
		rc = append(rc, a.LeftVisualAcuity.String)
		rc = append(rc, a.LeftPreviousVisualAcuity.String)
		rc = append(rc, a.LeftDRGrade.String)
		rc = append(rc, a.LeftDME.String)
		rc = append(rc, a.DrReferral.String)
		rc = append(rc, a.DrReferralNotes.String)
		rc = append(rc, a.RightDRGradeOver.String)
		rc = append(rc, a.RightDMEOver.String)
		sr = ""
		if len(a.RightSuspectedOver.String) > 0 {
			sr = a.RightSuspectedOver.String
		}
		rc = append(rc, SliceStringToCommaSeparatedValue(sr))
		rc = append(rc, a.LeftDRGradeOver.String)
		rc = append(rc, a.LeftDMEOver.String)
		sr = ""
		if len(a.LeftSuspectedOver.String) > 0 {
			sr = a.LeftSuspectedOver.String
		}
		rc = append(rc, SliceStringToCommaSeparatedValue(sr))
		rc = append(rc, a.OverReferral.String)
		rc = append(rc, a.OverReferralNotes.String)

		if err := w.Write(rc); err != nil {
			return nil, err
		}
	}
	w.Flush()

	if err := w.Error(); err != nil {
		return nil, err
	}

	return b, nil
}

func storeToGoogleCloudStorage(filename string, bytesBuffer *bytes.Buffer) error {
	envVar := envy.Get("GOOGLE_APPLICATION_CREDENTIALS_PATH_FOR_EXPORT", "")
	err := os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", envVar)
	if err != nil {
		return err
	}

	ctx := context.Background()

	// Creates a client.
	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}

	// Sets the name for the new bucket.
	bucketName := envy.Get("GOOGLE_STORAGE_EXPORT_BUCKET_NAME", "piia_project_exports")

	wc := client.Bucket(bucketName).Object(filename).NewWriter(ctx)
	wc.ContentType = "text/csv"
	if _, err = io.Copy(wc, bytesBuffer); err != nil {
		return err
	}

	if err := wc.Close(); err != nil {
		return err
	}

	// googleStorageEmail := envy.Get("GOOGLE_STORAGE_SERVICE_EMAIL", "")
	// googleStoragePrivateKey := envy.Get("GOOGLE_STORAGE_SERVICE_PRIVATE_KEY", "")

	// googleStorageEmail := credentialContent["client_email"]
	// googleStoragePrivateKey := credentialContent["private_key"]

	// fmt.Println(googleStorageEmail)
	// fmt.Println(googleStoragePrivateKey)

	return nil
}
