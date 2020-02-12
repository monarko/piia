package grifts

import (
    "encoding/json"
    "fmt"
    "log"
    "os"

    "github.com/gobuffalo/envy"
    . "github.com/markbates/grift/grift"

    "github.com/monarko/piia/helpers"
    "github.com/monarko/piia/models"
)

var _ = Namespace("pubsub", func() {

    Namespace("pull", func() {

        Desc("converted", "Pull messages from image-converted")
        Add("converted", func(c *Context) error {
            envVar := envy.Get("HUB_SERVICE_ACCOUNT_PATH", "")
            err := os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", envVar)
            if err != nil {
                return fmt.Errorf("image-converted: %v", err)
            }

            projectID := envy.Get("SUB_PROJECT", "")
            subID := envy.Get("IMAGE_CONVERTED", "") + "-sub"
            msgs, err := helpers.PubSubPullMessages(projectID, subID)
            if err != nil {
                return fmt.Errorf("image-converted-sub: %v", err)
            }

            for _, m := range msgs {
                p := map[string]interface{}{}
                err := json.Unmarshal(m, &p)
                if err != nil {
                    log.Printf("image-converted unmarshal: %v", err)
                    continue
                }
                screening := &models.Screening{}
                q := tx.Where("accession_id = ?", p["accession_number"])
                err = q.First(screening)
                if err != nil {
                    log.Printf("image-converted find screening: %v", err)
                    continue
                }

                scrImage := &models.ScreeningImage{}
                scrImage.ScreeningID = screening.ID
                scrImage.Status.String = "unselected"
                scrImage.Status.Valid = true
                scrImage.Data = p

                verrs, err := tx.ValidateAndCreate(scrImage)
                if err != nil {
                    log.Printf("image-converted create image: %v", err)
                    continue
                }
                if verrs.HasAny() {
                    log.Printf("image-converted screening image: %v", verrs.Errors)
                    continue
                }
            }

            return nil
        })

    })

})
