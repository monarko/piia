package actions

import (
    "log"

    "github.com/gobuffalo/buffalo"
    "github.com/gobuffalo/envy"
    "github.com/gobuffalo/pop"

    "github.com/monarko/piia/helpers"
    "github.com/monarko/piia/models"
)

type ErrorResponse struct {
    Message string `json:"message"`
}

// ScreeningsDRAssessmentAPI function
func ScreeningsDRAssessmentAPI(c buffalo.Context) error {
    tx := c.Value("tx").(*pop.Connection)
    screening := &models.Screening{}
    if err := tx.Eager().Find(screening, c.Param("sid")); err != nil {
        er := ErrorResponse{Message: "Screening encounter not found"}
        return c.Render(404, r.JSON(er))
    }

    for _, s := range screening.ScreeningImages {
        surl, err := getEyeImage(s)
        if err != nil {
            er := ErrorResponse{Message: err.Error()}
            return c.Render(404, r.JSON(er))
        }
        s.Data["signed_url"] = surl.SignedURL
    }

    return c.Render(200, r.JSON(screening))
}

type postInput struct {
    Consent string `json:"consent"`
    Right   string `json:"right"`
    Left    string `json:"left"`
}

// ScreeningImagesSelected function
func ScreeningImagesSelected(c buffalo.Context) error {
    tx := c.Value("tx").(*pop.Connection)
    screening := &models.Screening{}
    if err := tx.Eager().Find(screening, c.Param("sid")); err != nil {
        er := ErrorResponse{Message: "Screening encounter not found"}
        return c.Render(404, r.JSON(er))
    }

    pi := &postInput{}
    if err := c.Bind(pi); err != nil {
        er := ErrorResponse{Message: err.Error()}
        return c.Render(500, r.JSON(er))
    }

    if len(pi.Right) > 0 && len(pi.Left) > 0 {
        selected := make([]map[string]interface{}, 0)
        for _, s := range screening.ScreeningImages {
            if s.ID.String() == pi.Right || s.ID.String() == pi.Left {
                s.Status.String = "selected"
                selected = append(selected, s.Data)
            } else {
                s.Status.String = "not selected"
            }
            _, err := tx.ValidateAndUpdate(&s)
            if err != nil {
                er := ErrorResponse{Message: err.Error()}
                return c.Render(500, r.JSON(er))
            }
        }
        screening.HubStatus.String = "processing"
        screening.HubStatus.Valid = true

        screening.Eyes.Consent.Valid = true
        screening.Eyes.Consent.Bool = false
        if pi.Consent == "Y" {
            screening.Eyes.Consent.Bool = true
        }

        if len(selected) == 2 {
            projectID := envy.Get("TOPIC_PROJECT", "")
            topicID := envy.Get("IMAGE_INGEST", "")
            type ingest struct {
                Consent string                   `json:"consent"`
                Images  []map[string]interface{} `json:"images"`
            }
            in := &ingest{}
            in.Consent = pi.Consent
            in.Images = selected
            id, err := helpers.PubSubPublish(projectID, topicID, in)
            if err != nil {
                er := ErrorResponse{Message: err.Error()}
                return c.Render(500, r.JSON(er))
            }
            log.Printf("Image Ingest PUB SUB Message ID for SID (%s): %s\n", screening.ID.String(), id)
        }

        verrs, err := tx.ValidateAndUpdate(screening)
        if err != nil {
            er := ErrorResponse{Message: err.Error()}
            return c.Render(500, r.JSON(er))
        }
        if verrs.HasAny() {
            er := ErrorResponse{Message: verrs.Error()}
            return c.Render(500, r.JSON(er))
        }
    }

    return c.Render(200, r.JSON(screening))
}
