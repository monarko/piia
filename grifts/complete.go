package grifts

import (
    "encoding/json"
    "fmt"
    "log"
    "os"
    "strings"

    "github.com/gobuffalo/envy"
    . "github.com/markbates/grift/grift"

    "github.com/monarko/piia/helpers"
    "github.com/monarko/piia/models"
)

var _ = Namespace("pubsub", func() {

    Namespace("pull", func() {

        Desc("complete", "Pull messages from diagnosis-complete")
        Add("complete", func(c *Context) error {
            envVar := envy.Get("HUB_SERVICE_ACCOUNT_PATH", "")
            err := os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", envVar)
            if err != nil {
                return fmt.Errorf("image-diagnose: %v", err)
            }

            fullTopic := envy.Get("IMAGE_DIAGNOSE_COMPLETE_SUB", "")
            br := strings.SplitN(fullTopic, "/", -1)
            projectID := br[1]
            subID := br[3]
            msgs, err := helpers.PubSubPullMessages(projectID, subID)
            if err != nil {
                return fmt.Errorf("image-diagnose-complete-sub: %v", err)
            }

            drOptions := []string{
                "Ungradeable",
                "Normal",
                "Mild DR",
                "Moderate DR",
                "Severe DR",
                "Proliferative DR",
            }
            dmeOptions := map[string]string{
                "yes": "Present",
                "no":  "Not Present",
            }

            for _, m := range msgs {
                var p map[string]interface{}
                err := json.Unmarshal(m, &p)
                if err != nil {
                    log.Printf("image-diagnose-complete unmarshal: %v", err)
                    continue
                }
                referral := p["referral"].(string)
                diagnosis := p["diagnosis"].([]interface{})

                for _, dg := range diagnosis {
                    d := dg.(map[string]interface{})
                    screening := &models.Screening{}
                    q := tx.Where("accession_id = ?", d["accession_number"])
                    err = q.First(screening)
                    if err != nil {
                        log.Printf("image-diagnose-complete find screening: %v", err)
                        continue
                    }

                    scrImage := &models.ScreeningImage{}
                    q = tx.Where("data->>'sop_instance_uid' = ?", d["sop_instance_uid"])
                    err = q.First(scrImage)
                    if err != nil {
                        log.Printf("image-diagnose-complete find scr image: %v", err)
                        continue
                    }
                    data := scrImage.Data
                    laterality := data["laterality"]
                    data["dr_grade"] = d["dr_grade"]
                    data["dme_grade"] = d["dme_grade"]
                    scrImage.Data = data

                    verrs, err := tx.ValidateAndUpdate(scrImage)
                    if err != nil {
                        log.Printf("image-diagnose-complete update image: %v", err)
                        continue
                    }
                    if verrs.HasAny() {
                        log.Printf("image-diagnose-complete screening image: %v", verrs.Errors)
                        continue
                    }

                    if screening.HubStatus.String == "diagnosing" {
                        screening.HubStatus.String = "diagnosed"
                        dr := d["dr_grade"].(map[string]interface{})
                        dme := d["dme_grade"].(map[string]interface{})
                        if laterality == "R" {
                            drGrade := dr["grade"].(string)
                            m, l := helpers.PartiallyMatch(drOptions, drGrade, 4)
                            if l >= 0 {
                                screening.Eyes.RightEye.DRGrading.String = m
                                screening.Eyes.RightEye.DRGrading.Valid = true
                            }

                            dmeGrade := dme["grade"].(string)
                            dm := "Ungradeable"
                            if !strings.HasPrefix(strings.ToLower(dmeGrade), "un") {
                                dm = dmeOptions[strings.ToLower(dmeGrade)]
                            }
                            screening.Eyes.RightEye.DMEAssessment.String = dm
                            screening.Eyes.RightEye.DMEAssessment.Valid = true
                        } else {
                            drGrade := dr["grade"].(string)
                            m, l := helpers.PartiallyMatch(drOptions, drGrade, 4)
                            if l >= 0 {
                                screening.Eyes.LeftEye.DRGrading.String = m
                                screening.Eyes.LeftEye.DRGrading.Valid = true
                            }

                            dmeGrade := dme["grade"].(string)
                            dm := "Ungradeable"
                            if !strings.HasPrefix(strings.ToLower(dmeGrade), "un") {
                                dm = dmeOptions[strings.ToLower(dmeGrade)]
                            }
                            screening.Eyes.LeftEye.DMEAssessment.String = dm
                            screening.Eyes.LeftEye.DMEAssessment.Valid = true
                        }
                        screening.Referral.Referred.Valid = true
                        if strings.HasPrefix(strings.ToLower(referral), "refer") {
                            screening.Referral.Referred.Bool = true
                        } else {
                            screening.Referral.Referred.Bool = false
                        }
                        screening.Referral.Notes.String = referral
                        screening.Referral.Notes.Valid = true

                        verrs, err = tx.ValidateAndUpdate(screening)
                        if err != nil {
                            log.Printf("image-diagnose-complete screening update: %v", err)
                            continue
                        }
                        if verrs.HasAny() {
                            log.Printf("image-diagnose-complete screening: %v", verrs.Errors)
                            continue
                        }
                    }
                }
            }

            return nil
        })

    })

})
