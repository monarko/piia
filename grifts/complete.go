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

            projectID := envy.Get("SUB_PROJECT", "")
            subID := envy.Get("IMAGE_DIAGNOSE_COMPLETE", "") + "-sub"
            msgs, err := helpers.PubSubPullMessages(projectID, subID)
            if err != nil {
                return fmt.Errorf("image-diagnose-complete-sub: %v", err)
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

                    screening.HubStatus.String = "diagnosed"
                    dr := d["dr_grade"].(map[string]interface{})
                    dme := d["dme_grade"].(map[string]interface{})
                    if laterality == "R" {
                        drGrade := dr["grade"].(string)
                        screening.Eyes.RightEye.DRGrading.String = drGrade
                        screening.Eyes.RightEye.DRGrading.Valid = true

                        dmeGrade := dme["grade"].(string)
                        screening.Eyes.RightEye.DMEAssessment.String = dmeGrade
                        screening.Eyes.RightEye.DMEAssessment.Valid = true
                    } else {
                        drGrade := dr["grade"].(string)
                        screening.Eyes.LeftEye.DRGrading.String = drGrade
                        screening.Eyes.LeftEye.DRGrading.Valid = true

                        dmeGrade := dme["grade"].(string)
                        screening.Eyes.LeftEye.DMEAssessment.String = dmeGrade
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
                    participant := screening.Participant
                    if participant.Status == "1" && screening.Referral.Referred.Valid {
                        participant.Status = "11"
                        perrs, err := tx.ValidateAndUpdate(&participant)
                        if err != nil {
                            log.Printf("image-diagnose-complete participant update: %v", err)
                            continue
                        }
                        if perrs.HasAny() {
                            log.Printf("image-diagnose-complete participant: %v", verrs.Errors)
                            continue
                        }
                    }
                }
            }

            return nil
        })

    })

})
