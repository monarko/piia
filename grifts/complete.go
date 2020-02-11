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

            fullTopic := envy.Get("IMAGE_DIAGNOSE_SUB", "")
            br := strings.SplitN(fullTopic, "/", -1)
            projectID := br[1]
            subID := br[3]
            msgs, err := helpers.PubSubPullMessages(projectID, subID)
            if err != nil {
                return fmt.Errorf("image-diagnose-sub: %v", err)
            }

            for _, m := range msgs {
                p := map[string]interface{}{}
                err := json.Unmarshal(m, &p)
                if err != nil {
                    log.Printf("image-diagnose unmarshal: %v", err)
                    continue
                }
                images := p["images"].([]map[string]interface{})

                for _, i := range images {
                    screening := &models.Screening{}
                    q := tx.Where("accession_id = ?", i["accession_number"])
                    err = q.First(screening)
                    if err != nil {
                        log.Printf("image-diagnose find screening: %v", err)
                        continue
                    }

                    scrImage := &models.ScreeningImage{}
                    q = tx.Where("data->>'sop_instance_uid' = ?", i["sop_instance_uid"])
                    err = q.First(scrImage)
                    if err != nil {
                        log.Printf("image-diagnose find scr image: %v", err)
                        continue
                    }
                    data := scrImage.Data
                    data["is_success"] = i["is_success"]
                    if r, ok := i["reason"]; ok {
                        data["reason"] = r
                    }
                    scrImage.Data = data

                    verrs, err := tx.ValidateAndUpdate(scrImage)
                    if err != nil {
                        log.Printf("image-diagnose update image: %v", err)
                        continue
                    }
                    if verrs.HasAny() {
                        log.Printf("image-diagnose screening image: %v", verrs.Errors)
                        continue
                    }

                    if screening.HubStatus.String == "diagnosing" {
                        screening.HubStatus.String = "diagnosed"
                        verrs, err = tx.ValidateAndUpdate(screening)
                        if err != nil {
                            log.Printf("image-diagnose screening update: %v", err)
                            continue
                        }
                        if verrs.HasAny() {
                            log.Printf("image-diagnose screening: %v", verrs.Errors)
                            continue
                        }
                    }
                }
            }

            return nil
        })

    })

})
