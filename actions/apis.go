package actions

import (
    "github.com/gobuffalo/buffalo"
    "github.com/gobuffalo/pop"

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
