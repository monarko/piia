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
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop"
	"github.com/monarko/piia/helpers"
	"github.com/monarko/piia/models"
)

type fullRecord struct {
	CreatedDate                     time.Time    `json:"created_date" db:"created_date"`
	ScreeningID                     string       `json:"screening_id" db:"screening_id"`
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
	RightDilation                   nulls.String `json:"right_dilation" db:"right_dilation"`
	LeftDilation                    nulls.String `json:"left_dilation" db:"left_dilation"`
	ScreeningAssessmentDate         nulls.String `json:"screening_assessment_date" db:"screening_assessment_date"`
	DrReferral                      nulls.String `json:"dr_referral" db:"dr_referral"`
	HospitalReferral                nulls.String `json:"hospital_referral" db:"hospital_referral"`
	DrReferralRefused               nulls.String `json:"dr_referral_refused" db:"dr_referral_refused"`
	DrReferralReason                nulls.String `json:"dr_referral_reason" db:"dr_referral_reason"`
	HospitalNotReferralReason       nulls.String `json:"hospital_not_referral_reason" db:"hospital_not_referral_reason"`
	DrReferralNotes                 nulls.String `json:"dr_referral_notes" db:"dr_referral_notes"`
	RightDRGradeOver                nulls.String `json:"right_dr_grade_over" db:"right_dr_grade_over"`
	RightDMEOver                    nulls.String `json:"right_dme_over" db:"right_dme_over"`
	RightSuspectedOver              nulls.String `json:"right_suspected_over" db:"right_suspected_over"`
	LeftDRGradeOver                 nulls.String `json:"left_dr_grade_over" db:"left_dr_grade_over"`
	LeftDMEOver                     nulls.String `json:"left_dme_over" db:"left_dme_over"`
	LeftSuspectedOver               nulls.String `json:"left_suspected_over" db:"left_suspected_over"`
	OverReadingAssessmentDate       nulls.String `json:"over_assessment_date" db:"over_assessment_date"`
	OverReferral                    nulls.String `json:"over_referral" db:"over_referral"`
	OverReferralNotes               nulls.String `json:"over_referral_notes" db:"over_referral_notes"`
	ReferredMessageID               nulls.String `json:"referred_message_id" db:"referred_message_id"`
}

// DownloadFull function
func DownloadFull(c buffalo.Context) error {
	user := c.Value("current_user").(*models.User)
	filename, b, err := downloadFull(c)
	if err != nil {
		errs := map[string][]string{
			"index_error": {err.Error()},
		}
		c.Set("errors", errs)
		InsertLog("error", "User download analytics error", err.Error(), "", "", user.ID, c)
		return c.Redirect(302, "/analytics/index")
	}

	err = storeToGoogleCloudStorage(filename, b)
	if err != nil {
		errs := map[string][]string{
			"index_error": {err.Error()},
		}
		c.Set("errors", errs)
		c.Flash().Add("danger", "Error from GCS: "+err.Error())
		InsertLog("error", "User download analytics error", err.Error(), "", "", user.ID, c)
	} else {
		c.Flash().Add("success", "Successfully saved to your storage bucket")
	}

	return c.Redirect(302, "/analytics/index")
}

// DownloadFullCSV function
func DownloadFullCSV(c buffalo.Context) error {
	user := c.Value("current_user").(*models.User)
	filename, b, err := downloadFull(c)
	if err != nil {
		errs := map[string][]string{
			"index_error": {err.Error()},
		}
		c.Set("errors", errs)
		c.Flash().Add("danger", "Error: "+err.Error())
		InsertLog("error", "User download csv analytics error", err.Error(), "", "", user.ID, c)
		return c.Redirect(302, "/analytics/index")
	}

	return c.Render(200, r.Download(c, filename, b))
}

func downloadFull(c buffalo.Context) (string, *bytes.Buffer, error) {
	tx := c.Value("tx").(*pop.Connection)
	fullRecords := make([]fullRecord, 0)

	query := `SELECT
	s.created_at AS "created_date",
	s.id AS "screening_id",
	p.participant_id AS "participant_id",
	date_part('year', age(((p.dob->>'calculated_date'::text)::date)::timestamp with time zone)) AS "age",
        CASE
            WHEN ((p.gender)::text = 'M'::text) THEN 'Male'::text
	    WHEN ((p.gender)::text = 'F'::text) THEN 'Female'::text
            ELSE ''::text
        END AS "gender",

	(s.diabetes->>'diabetes_type'::text) AS "diabetes_type",
	(s.diabetes->>'duration'::text) AS "diabetes_duration",
	(s.diabetes->>'duration_type'::text) AS "diabetes_duration_type",

	CASE
            WHEN ((s.medical_history->>'smoker'::text) = 'true'::text) THEN 'Yes'::text
	    WHEN ((s.medical_history->>'smoker'::text) = 'false'::text) THEN 'No'::text
            ELSE NULL
        END AS "medical_history_smoker",
	(s.medical_history->>'morbidities'::text) AS "medical_history_morbidities",

	CASE
            WHEN ((s.medications->>'on_insulin'::text) = 'true'::text) THEN 'Yes'::text
            WHEN ((s.medications->>'on_insulin'::text) = 'false'::text) THEN 'No'::text
            ELSE NULL
        END AS "medications_on_insulin",
	CASE
            WHEN ((s.medications->>'taking_medications'::text) = 'true'::text) THEN 'Yes'::text
	    WHEN ((s.medications->>'taking_medications'::text) = 'false'::text) THEN 'No'::text
            ELSE NULL
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
	((s.eye->'right'::text)->>'dilate_pupil'::text) AS "right_dilation",
	((s.eye->'left'::text)->>'dilate_pupil'::text) AS "left_dilation",
	CASE
	    WHEN ((s.eye->'assessment_date'::text)->>'calculated_date'::text) IS NULL THEN TO_CHAR(s.created_at, 'YYYY-MM-DD')
	    ELSE
		CASE 
			WHEN (LEFT(((s.eye->'assessment_date'::text)->>'calculated_date'::text), 10) = '0001-01-01'::text) THEN NULL
			ELSE LEFT(((s.eye->'assessment_date'::text)->>'calculated_date'::text), 10)
		END 
	END AS screening_assessment_date,
	CASE
		WHEN ((s.referral->>'referred'::text) = 'true'::text) THEN 'Yes'::text
		WHEN ((s.referral->>'referred'::text) = 'false'::text) THEN 'No'::text
		ELSE NULL
	END AS "dr_referral",
	CASE
		WHEN ((s.referral->>'hospital_referred'::text) = 'true'::text) THEN 'Yes'::text
		WHEN ((s.referral->>'hospital_referred'::text) = 'false'::text) THEN 'No'::text
		ELSE NULL
	END AS "hospital_referral",
	CASE
		WHEN ((s.referral->>'referral_refused'::text) = 'true'::text) THEN 'Yes'::text
		WHEN ((s.referral->>'referral_refused'::text) = 'false'::text) THEN 'No'::text
		ELSE NULL
	END AS "dr_referral_refused",
	(s.referral->>'referral_reason'::text) AS "dr_referral_reason",
	(s.referral->>'hospital_not_referral_reason'::text) AS "hospital_not_referral_reason",
	(s.referral->>'additional_notes'::text) AS "dr_referral_notes",

	((o.eye_assessment->'right'::text)->>'dr'::text) AS "right_dr_grade_over",
	((o.eye_assessment->'right'::text)->>'dme'::text) AS "right_dme_over",
	((o.eye_assessment->'right'::text)->>'suspected_pathologies'::text) AS "right_suspected_over",
	((o.eye_assessment->'left'::text)->>'dr'::text) AS "left_dr_grade_over",
	((o.eye_assessment->'left'::text)->>'dme'::text) AS "left_dme_over",
	((o.eye_assessment->'left'::text)->>'suspected_pathologies'::text) AS "left_suspected_over",
	TO_CHAR(o.created_at, 'YYYY-MM-DD') AS over_assessment_date,
        CASE
            WHEN ((o.referral->>'referred'::text) = 'true'::text) THEN 'Yes'::text
	    WHEN ((o.referral->>'referred'::text) = 'false'::text) THEN 'No'::text
            ELSE NULL
        END AS "over_referral",
	(o.referral->>'additional_notes'::text) AS "over_referral_notes",
	r.id AS "referred_message_id"
FROM (
	participants p
	LEFT JOIN screenings s ON (p.id = s.participant_id)
	LEFT JOIN over_readings o ON (o.screening_id = s.id)
	LEFT JOIN referred_messages r ON (r.screening_id = s.id)
)
WHERE (
	s.id IS NOT NULL
)
ORDER BY s.created_at`

	q := tx.RawQuery(query)

	if err := q.All(&fullRecords); err != nil {
		return "", nil, err
	}

	appHost := envy.Get("APP_HOST", "http://127.0.0.1")
	hosts := strings.Split(strings.TrimSpace(strings.Replace(appHost, "/", "", -1)), ":")
	filename := hosts[1] + "-full-record" + ".csv"

	b, err := downloadAllRecords(fullRecords)
	if err != nil {
		return "", nil, err
	}

	return filename, b, nil
}

func downloadAllRecords(records []fullRecord) (*bytes.Buffer, error) {
	b := &bytes.Buffer{}
	w := csv.NewWriter(b)

	headers := []string{
		"created_date",
		"study_site",
		"study_id",
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
		"pupil_dilation",
		"screening_assessment_date",
		"screening_referral",
		"screening_hospital_referral",
		"screening_referral_refused",
		"screening_referral_reason",
		"screening_hospital_not_referral_reason",
		"screening_referral_notes",
		"overreading_right_dr_grade",
		"overreading_right_dme",
		"overreading_right_suspected_pathology",
		"overreading_left_dr_grade",
		"overreading_left_dme",
		"overreading_left_suspected_pathology",
		"overreading_gradeability",
		"overreading_assessment_date",
		"overreading_referral",
		"overreading_referral_notes",
		"overreading_referral_details",
		"referral_tracked",
	}

	if err := w.Write(headers); err != nil {
		return nil, err
	}

	sites := GetSystemSites()

	for _, a := range records {
		var rc []string

		rc = append(rc, a.CreatedDate.Format(time.RFC3339))
		rc = append(rc, sites[a.ParticipantID[1:2]])
		rc = append(rc, a.ScreeningID)
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
		rc = append(rc, helpers.SliceStringToCommaSeparatedValue(sr))
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

		pupilDilation := ""
		if a.RightDilation.Valid && a.RightDilation.String == "true" && a.LeftDilation.Valid && a.LeftDilation.String == "true" {
			pupilDilation = "BOTH EYES"
		} else if a.RightDilation.Valid && a.RightDilation.String == "false" && a.LeftDilation.Valid && a.LeftDilation.String == "false" {
			pupilDilation = "NO"
		} else if a.RightDilation.Valid && a.RightDilation.String == "true" {
			pupilDilation = "RIGHT EYE"
		} else if a.LeftDilation.Valid && a.LeftDilation.String == "true" {
			pupilDilation = "LEFT EYE"
		}
		rc = append(rc, pupilDilation)

		rc = append(rc, a.ScreeningAssessmentDate.String)
		rc = append(rc, a.DrReferral.String)
		rc = append(rc, a.HospitalReferral.String)
		rc = append(rc, a.DrReferralRefused.String)
		screeningReasons := ""
		if a.DrReferral.String == "Yes" && len(a.DrReferralReason.String) == 0 {
			screeningReasons = screeningDetails(
				a.RightVisualAcuity.String,
				a.RightPreviousVisualAcuity.String,
				a.RightDRGrade.String,
				a.RightDME.String,
				a.LeftVisualAcuity.String,
				a.LeftPreviousVisualAcuity.String,
				a.LeftDRGrade.String,
				a.LeftDME.String,
			)
		} else {
			screeningReasons = a.DrReferralReason.String
		}
		rc = append(rc, screeningReasons)
		rc = append(rc, a.HospitalNotReferralReason.String)
		haveNotes := "FALSE"
		if a.DrReferralNotes.Valid && len(a.DrReferralNotes.String) > 0 {
			haveNotes = "TRUE"
		}
		rc = append(rc, haveNotes)
		rc = append(rc, a.RightDRGradeOver.String)
		rc = append(rc, a.RightDMEOver.String)
		srs := ""
		if len(a.RightSuspectedOver.String) > 0 {
			srs = a.RightSuspectedOver.String
		}
		rc = append(rc, helpers.SliceStringToCommaSeparatedValue(srs))
		rc = append(rc, a.LeftDRGradeOver.String)
		rc = append(rc, a.LeftDMEOver.String)
		sls := ""
		if len(a.LeftSuspectedOver.String) > 0 {
			sls = a.LeftSuspectedOver.String
		}
		rc = append(rc, helpers.SliceStringToCommaSeparatedValue(sls))
		overReadingGradeability := ""
		if a.RightDRGradeOver.Valid && a.RightDMEOver.Valid && a.LeftDRGradeOver.Valid && a.LeftDMEOver.Valid {
			overReadingGradeability = "GRADEABLE"
			if a.RightDRGradeOver.String == "Ungradeable" || a.RightDMEOver.String == "Ungradeable" || a.LeftDRGradeOver.String == "Ungradeable" || a.LeftDMEOver.String == "Ungradeable" {
				overReadingGradeability = "UNGRADEABLE"
			}
		}
		rc = append(rc, overReadingGradeability)
		rc = append(rc, a.OverReadingAssessmentDate.String)
		rc = append(rc, a.OverReferral.String)
		rc = append(rc, a.OverReferralNotes.String)
		overreadingReasons := ""
		if a.OverReferral.String == "Yes" {
			overreadingReasons = overreadingDetails(
				a.RightDRGrade.String,
				a.RightDME.String,
				helpers.SliceStringToCommaSeparatedValue(srs),
				a.LeftDRGrade.String,
				a.LeftDME.String,
				helpers.SliceStringToCommaSeparatedValue(sls),
			)
		}
		rc = append(rc, overreadingReasons)
		referralPresent := "NO"
		if a.ReferredMessageID.Valid && len(a.ReferredMessageID.String) > 0 {
			referralPresent = "YES"
		}
		rc = append(rc, referralPresent)

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
	bucketName := envy.Get("GOOGLE_STORAGE_EXPORT_BUCKET_NAME", "")

	wc := client.Bucket(bucketName).Object(filename).NewWriter(ctx)
	wc.ContentType = "text/csv"
	if _, err = io.Copy(wc, bytesBuffer); err != nil {
		return err
	}

	if err := wc.Close(); err != nil {
		return err
	}

	return nil
}

func screeningDetails(rightVA, rightLastVA, rightDR, rightDME, leftVA, leftLastVA, leftDR, leftDME string) string {
	reasons := make([]string, 0)

	right := screeningDetailsForEye(rightVA, rightLastVA, rightDR, rightDME, "R")
	if len(right) > 0 {
		reasons = append(reasons, right...)
	}
	left := screeningDetailsForEye(leftVA, leftLastVA, leftDR, leftDME, "L")
	if len(left) > 0 {
		reasons = append(reasons, left...)
	}

	return strings.Join(reasons, ", ")
}

func screeningDetailsForEye(va, lastVA, dr, dme, eye string) []string {
	reasons := make([]string, 0)

	vaValues := []string{"20/20", "20/30", "20/40", "20/50", "20/70", "20/100", "20/200", "CF", "HM", "LP", "NLP"}
	vaIndex := helpers.IndexInSlice(vaValues, va)
	lastVAIndex := helpers.IndexInSlice(vaValues, lastVA)

	if dr == "Ungradeable" {
		reasons = append(reasons, "Ungradeable for DR ("+eye+")")
	} else if dr == "Severe DR" || dr == "Proliferative DR" {
		reasons = append(reasons, dr+" ("+eye+")")
	}

	if dme == "Ungradeable" {
		reasons = append(reasons, "Ungradeable for DME ("+eye+")")
	} else if dme == "Present" {
		reasons = append(reasons, "DME Present ("+eye+")")
	}

	if vaIndex >= 4 {
		reasons = append(reasons, "V/A is "+va+" ("+eye+")")
	} else if vaIndex < 4 && (lastVAIndex-vaIndex) >= 2 {
		reasons = append(reasons, "V/A from "+va+" to "+lastVA+" ("+eye+")")
	}

	return reasons
}

func overreadingDetails(rightDR, rightDME, rightSuspected, leftDR, leftDME, leftSuspected string) string {
	reasons := make([]string, 0)

	right := overreadingDetailsForEye(rightDR, rightDME, rightSuspected, "R")
	if len(right) > 0 {
		reasons = append(reasons, right...)
	}

	left := overreadingDetailsForEye(leftDR, leftDME, leftSuspected, "L")
	if len(left) > 0 {
		reasons = append(reasons, left...)
	}

	return strings.Join(reasons, ", ")
}

func overreadingDetailsForEye(dr, dme, suspected, eye string) []string {
	reasons := make([]string, 0)

	if dr == "Ungradeable" {
		reasons = append(reasons, "Ungradeable for DR ("+eye+")")
	} else if dr == "Severe DR" || dr == "Proliferative DR" {
		reasons = append(reasons, dr+" ("+eye+")")
	}

	if dme == "Ungradeable" {
		reasons = append(reasons, "Ungradeable for DME ("+eye+")")
	} else if dme == "Present" {
		reasons = append(reasons, "DME Present ("+eye+")")
	}

	if len(suspected) > 0 {
		sv := strings.Split(suspected, ", ")
		for _, s := range sv {
			reasons = append(reasons, s+" ("+eye+")")
		}
	}

	return reasons
}
