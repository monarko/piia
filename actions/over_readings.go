package actions

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/gobuffalo/envy"
	"google.golang.org/api/iterator"

	"cloud.google.com/go/storage"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop"
	"github.com/monarko/piia/models"
	"github.com/pkg/errors"
)

// OverReadingsIndex gets all OverReadings. This function is mapped to the path
func OverReadingsIndex(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	participant := &models.Participant{}
	if err := tx.Find(participant, c.Param("pid")); err != nil {
		return c.Error(404, err)
	}
	c.Set("participant", participant)

	overReadings := &models.OverReadings{}
	q := tx.Eager("OverReader").Where("participant_id = ?", c.Param("pid")).PaginateFromParams(c.Params()).Order("created_at DESC")
	// Retrieve all OverReadings from the DB
	if err := q.All(overReadings); err != nil {
		return errors.WithStack(err)
	}
	c.Set("overReadings", overReadings)

	// Add the paginator to the context so it can be used in the template.
	c.Set("pagination", q.Paginator)
	breadcrumbMap := make(map[string]interface{})
	breadcrumbMap["Cases"] = "/cases/index"
	// breadcrumbMap["Over Readings"] = "/participants/" + c.Param("pid") + "/overreadings/index"
	c.Set("breadcrumbMap", breadcrumbMap)
	return c.Render(200, r.HTML("over_readings/index.html"))
}

// OverReadingsCreateGet renders the form for creating a new OverReading.
func OverReadingsCreateGet(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	participant := &models.Participant{}
	if err := tx.Eager("Screenings", "OverReadings").Find(participant, c.Param("pid")); err != nil {
		return c.Error(404, err)
	}
	screening := participant.Screenings[0]
	if len(participant.OverReadings) > 0 {
		return c.Redirect(302, "/cases/index")
	}
	if !(c.Param("pid") == participant.ID.String() && c.Param("sid") == screening.ID.String()) {
		c.Flash().Add("danger", "Not Found")
		return c.Redirect(302, "/cases/index")
	}
	c.Set("participant", participant)
	c.Set("screening", screening)
	c.Set("overReading", &models.OverReading{})

	right, left, err := getImage(participant.ParticipantID)
	if err != nil {
		left = ""
		right = ""
	}

	c.Set("leftEyeLink", left)
	c.Set("rightEyeLink", right)

	breadcrumbMap := make(map[string]interface{})
	breadcrumbMap["Cases"] = "/cases/index"
	breadcrumbMap["New Over Reading"] = "#"
	c.Set("breadcrumbMap", breadcrumbMap)
	return c.Render(200, r.HTML("over_readings/create.html"))
}

// OverReadingsCreatePost renders the form for creating a new OverReading.
func OverReadingsCreatePost(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	participant := &models.Participant{}
	if err := tx.Eager("Screenings", "OverReadings").Find(participant, c.Param("pid")); err != nil {
		return c.Error(404, err)
	}
	screening := participant.Screenings[0]
	if len(participant.OverReadings) > 0 {
		return c.Redirect(302, "/cases/index")
	}
	if !(c.Param("pid") == participant.ID.String() && c.Param("sid") == screening.ID.String()) {
		c.Flash().Add("danger", "Not Found")
		return c.Redirect(302, "/cases/index")
	}
	user := c.Value("current_user").(*models.User)
	overReading := &models.OverReading{}
	oldOverReading := overReading.Maps()
	if err := c.Bind(overReading); err != nil {
		return errors.WithStack(err)
	}

	overReading.OverReaderID = user.ID
	overReading.ParticipantID = participant.ID
	overReading.ScreeningID = screening.ID

	referral := c.Request().FormValue("referral")
	if referral == "yes" {
		overReading.Referral.Referred = true
	}

	// images
	right, left, err := getImage(participant.ParticipantID)
	if err != nil {
		left = ""
		right = ""
	}

	c.Set("leftEyeLink", left)
	c.Set("rightEyeLink", right)

	verrs, err := tx.ValidateAndCreate(overReading)
	if err != nil {
		return errors.WithStack(err)
	}
	if verrs.HasAny() {
		c.Set("participant", participant)
		c.Set("screening", screening)
		c.Set("overReading", overReading)
		c.Set("errors", verrs.Errors)
		breadcrumbMap := make(map[string]interface{})
		breadcrumbMap["Cases"] = "/cases/index"
		breadcrumbMap["New Over Reading"] = "#"
		c.Set("breadcrumbMap", breadcrumbMap)
		return c.Render(422, r.HTML("over_readings/create.html"))
	}

	participant.Status = "111"
	perrs, err := tx.ValidateAndUpdate(participant)
	if err != nil {
		return errors.WithStack(err)
	}
	if perrs.HasAny() {
		c.Set("participant", participant)
		c.Set("screening", screening)
		c.Set("overReading", overReading)
		c.Set("errors", perrs.Errors)
		breadcrumbMap := make(map[string]interface{})
		breadcrumbMap["Cases"] = "/cases/index"
		breadcrumbMap["New Over Reading"] = "#"
		c.Set("breadcrumbMap", breadcrumbMap)
		return c.Render(422, r.HTML("over_readings/create.html"))
	}

	newOverReading := overReading.Maps()
	auditErr := MakeAudit("OverReading", overReading.ID, oldOverReading, newOverReading, user.ID, c)
	if auditErr != nil {
		return errors.WithStack(auditErr)
	}

	logErr := InsertLog("create", "Case overread", "", overReading.ID.String(), "overReading", user.ID, c)
	if logErr != nil {
		return errors.WithStack(logErr)
	}

	checkReferral := checkScreeningAndOverReading(&screening, overReading)
	if checkReferral {
		notifErr := InsertNotification(
			"Referral Notification",
			"This participant should be referred. Please contact to arrange.",
			"open",
			string(participant.ParticipantID[1]),
			user.ID,
			participant.ID,
			screening.ID,
			c,
		)
		if notifErr != nil {
			return errors.WithStack(notifErr)
		}
	}

	// If there are no errors set a success message
	c.Flash().Add("success", "New over reading added successfully.")

	return c.Redirect(302, "/cases/index")
}

// OverReadingsEditGet renders the form for creating a new OverReading.
func OverReadingsEditGet(c buffalo.Context) error {
	user := c.Value("current_user").(*models.User)
	tx := c.Value("tx").(*pop.Connection)
	overReading := &models.OverReading{}
	if err := tx.Eager().Find(overReading, c.Param("oid")); err != nil {
		return c.Error(404, err)
	}
	if !(overReading.OverReaderID == user.ID || user.Admin) {
		c.Flash().Add("danger", "Access denied")
		return c.Redirect(302, "/cases/index")
	}
	participant := overReading.Participant
	screening := overReading.Screening
	if !(c.Param("pid") == participant.ID.String() && c.Param("sid") == screening.ID.String()) {
		c.Flash().Add("danger", "Not Found")
		return c.Redirect(302, "/cases/index")
	}
	c.Set("participant", participant)
	c.Set("screening", screening)
	c.Set("overReading", overReading)

	right, left, err := getImage(participant.ParticipantID)
	if err != nil {
		left = ""
		right = ""
	}

	c.Set("leftEyeLink", left)
	c.Set("rightEyeLink", right)

	breadcrumbMap := make(map[string]interface{})
	breadcrumbMap["Cases"] = "/cases/index"
	breadcrumbMap["Edit Over Reading"] = "#"
	c.Set("breadcrumbMap", breadcrumbMap)
	return c.Render(200, r.HTML("over_readings/edit.html"))
}

// OverReadingsEditPost renders the form for creating a new OverReading.
func OverReadingsEditPost(c buffalo.Context) error {
	user := c.Value("current_user").(*models.User)
	tx := c.Value("tx").(*pop.Connection)
	overReading := &models.OverReading{}
	if err := tx.Eager().Find(overReading, c.Param("oid")); err != nil {
		return c.Error(404, err)
	}
	if !(overReading.OverReaderID == user.ID || user.Admin) {
		c.Flash().Add("danger", "Access denied")
		return c.Redirect(302, "/cases/index")
	}
	participant := overReading.Participant
	screening := overReading.Screening
	if !(c.Param("pid") == participant.ID.String() && c.Param("sid") == screening.ID.String()) {
		c.Flash().Add("danger", "Not Found")
		return c.Redirect(302, "/cases/index")
	}
	c.Set("participant", participant)
	c.Set("screening", screening)
	oldOverReading := overReading.Maps()
	if err := c.Bind(overReading); err != nil {
		return errors.WithStack(err)
	}

	referral := c.Request().FormValue("referral")
	if referral == "yes" {
		overReading.Referral.Referred = true
	}

	// images
	right, left, err := getImage(participant.ParticipantID)
	if err != nil {
		left = ""
		right = ""
	}
	c.Set("leftEyeLink", left)
	c.Set("rightEyeLink", right)

	verrs, err := tx.ValidateAndUpdate(overReading)
	if err != nil {
		return errors.WithStack(err)
	}
	if verrs.HasAny() {
		c.Set("participant", participant)
		c.Set("screening", screening)
		c.Set("overReading", overReading)
		c.Set("errors", verrs.Errors)
		breadcrumbMap := make(map[string]interface{})
		breadcrumbMap["Cases"] = "/cases/index"
		breadcrumbMap["Edit Over Reading"] = "#"
		c.Set("breadcrumbMap", breadcrumbMap)
		return c.Render(422, r.HTML("over_readings/edit.html"))
	}

	newOverReading := overReading.Maps()
	auditErr := MakeAudit("OverReading", overReading.ID, oldOverReading, newOverReading, user.ID, c)
	if auditErr != nil {
		return errors.WithStack(auditErr)
	}

	logErr := InsertLog("update", "Case overread update", "", overReading.ID.String(), "overReading", user.ID, c)
	if logErr != nil {
		return errors.WithStack(logErr)
	}

	checkReferral := checkScreeningAndOverReading(&screening, overReading)
	if checkReferral {
		notifErr := InsertNotification(
			"Referral Notification",
			"This participant should be referred. Please contact to arrange.",
			"open",
			string(participant.ParticipantID[1]),
			user.ID,
			participant.ID,
			screening.ID,
			c,
		)
		if notifErr != nil {
			return errors.WithStack(notifErr)
		}
	}

	// If there are no errors set a success message
	c.Flash().Add("success", "Over reading edited successfully.")

	return c.Redirect(302, "/cases/index")
}

func checkScreeningAndOverReading(screening *models.Screening, overReading *models.OverReading) bool {
	if !screening.Referral.Referred && overReading.Referral.Referred {
		return true
	}
	return false
}

func shouldBeReferred(overReading *models.OverReading) bool {
	refer := false

	if overReading.Eyes.LeftEye.DRGrading.String == "Ungradeable" || overReading.Eyes.LeftEye.DRGrading.String == "Severe DR" || overReading.Eyes.RightEye.DRGrading.String == "Ungradeable" || overReading.Eyes.RightEye.DRGrading.String == "Severe DR" {
		refer = true
	}

	if overReading.Eyes.LeftEye.DMEAssessment.String == "Present" || overReading.Eyes.RightEye.DMEAssessment.String == "Present" {
		refer = true
	}

	return refer
}

func getImage(participantID string) (string, string, error) {
	// If getting host is down, check network filter on little snitch
	// return "", "", nil
	// fmt.Println("---- GET IMAGE ----")
	envVar := envy.Get("GOOGLE_APPLICATION_CREDENTIALS_PATH", "")
	err := os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", envVar)
	if err != nil {
		return "", "", err
	}
	// fmt.Println("---- ENV SET ----")

	credentialFile, err := ioutil.ReadFile(envVar)
	if err != nil {
		return "", "", err
	}
	// fmt.Println("---- READ CREDENTIAL FILE ----")

	credentialContent := make(map[string]string)
	if err := json.Unmarshal(credentialFile, &credentialContent); err != nil {
		return "", "", err
	}
	// fmt.Println("---- MAKE CREDENTIAL CONTENT ----", credentialContent)

	// fmt.Printf("\n%#v\n", credentialContent)

	pID := participantID
	cleanPID := strings.Replace(pID, "-", "", -1)

	// fileNames := map[string]string{"right": pID + "_RIGHT_1543476715_40971517.png", "left": pID + "_LEFT_1543476746_40971518.png"}
	fileNames := make(map[string]string)
	right := ""
	left := ""

	ctx := context.Background()

	// Sets your Google Cloud Platform project ID.
	// projectID := "piia-project"

	// Creates a client.
	client, err := storage.NewClient(ctx)
	if err != nil {
		return right, left, err
	}

	// // Sets the name for the new bucket.
	bucketName := envy.Get("GOOGLE_STORAGE_BUCKET_NAME", "piia_images")

	// Creates a Bucket instance.
	bucket := client.Bucket(bucketName)

	idsToCheck := []string{strings.ToUpper(pID), strings.ToUpper(cleanPID), strings.ToLower(pID), strings.ToLower(cleanPID)}
	// fmt.Println("---- IDS TO CHECK ----", idsToCheck)

	for _, id := range idsToCheck {
		// fmt.Println(id)
		objs := bucket.Objects(ctx, &storage.Query{
			Prefix:    id,
			Delimiter: "",
		})
		// fmt.Println("---- ID ----", id)
		i := 0
		for {
			if i > 2 {
				break
			}
			i++
			attrs, err := objs.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				continue
			}
			// fmt.Println("---- ID Filename ----", attrs.Name)
			name := strings.ToLower(attrs.Name)
			if strings.Contains(name, "right") {
				fileNames["right"] = attrs.Name
			} else if strings.Contains(name, "left") {
				fileNames["left"] = attrs.Name
			}
		}
		if len(fileNames) != 0 {
			break
		}
	}

	// fmt.Println("---- FINAL ----", fileNames)

	if len(fileNames) == 0 {
		return right, left, errors.New("no file found for the participant id")
	}

	// // rc, err := bucket.Object(fileNames[0]). .NewReader(ctx)
	// // if err != nil {
	// // 	return "", "", err
	// // }
	// // defer rc.Close()
	// // body, err := ioutil.ReadAll(rc)
	// // if err != nil {
	// // 	return "", "", err
	// // }

	method := "GET"
	expires := time.Now().Add(time.Second * 60 * 10)

	// googleStorageEmail := envy.Get("GOOGLE_STORAGE_SERVICE_EMAIL", "")
	// googleStoragePrivateKey := envy.Get("GOOGLE_STORAGE_SERVICE_PRIVATE_KEY", "")

	googleStorageEmail := credentialContent["client_email"]
	googleStoragePrivateKey := credentialContent["private_key"]

	// fmt.Println(googleStorageEmail)
	// fmt.Println(googleStoragePrivateKey)

	for k, v := range fileNames {
		url, err := storage.SignedURL(bucketName, v, &storage.SignedURLOptions{
			GoogleAccessID: googleStorageEmail,
			PrivateKey:     []byte(googleStoragePrivateKey),
			Method:         method,
			Expires:        expires,
		})
		if err != nil {
			continue
		}
		if k == "right" {
			right = url
		} else if k == "left" {
			left = url
		}
	}

	return right, left, nil
}

// OverReadingsDetails renders the form for creating a new OverReading.
func OverReadingsDetails(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	participant := &models.Participant{}
	if err := tx.Eager("Screenings").Find(participant, c.Param("pid")); err != nil {
		return c.Error(404, err)
	}
	screening := participant.Screenings[0]
	c.Set("participant", participant)
	c.Set("screening", screening)
	overReadings := &models.OverReading{}
	if err := tx.Eager().Find(overReadings, c.Param("oid")); err != nil {
		return c.Error(404, err)
	}
	c.Set("overReading", overReadings)

	// images
	// response, err := http.Get("http://localhost:8080/" + participant.ParticipantID)
	// if err != nil {
	// 	// If there are no errors set a success message
	// 	c.Flash().Add("danger", "Error from the image server")

	// 	return c.Redirect(302, "/cases/index")
	// }
	// defer response.Body.Close()
	// data, _ := ioutil.ReadAll(response.Body)
	// respData := map[string]string{}
	// uerr := json.Unmarshal(data, &respData)
	// if uerr != nil {
	// 	// If there are no errors set a success message
	// 	c.Flash().Add("danger", "Error from the image server")

	// 	return c.Redirect(302, "/cases/index")
	// }

	right, left, err := getImage(participant.ParticipantID)
	if err != nil {
		left = ""
		right = ""
	}

	c.Set("leftEyeLink", left)
	c.Set("rightEyeLink", right)

	breadcrumbMap := make(map[string]interface{})
	breadcrumbMap["Cases"] = "/cases/index"
	// breadcrumbMap["Over Readings"] = "/participants/" + c.Param("pid") + "/overreadings/index"
	breadcrumbMap["Over Reading"] = "/cases/" + c.Param("pid") + "/overreadings/" + c.Param("oid")
	c.Set("breadcrumbMap", breadcrumbMap)
	return c.Render(200, r.HTML("over_readings/details.html"))
}

// OverReadingDestroy function
func OverReadingDestroy(c buffalo.Context) error {
	returnURL := "/cases/index"
	user := c.Value("current_user").(*models.User)
	if !user.Admin {
		c.Flash().Add("danger", "Access denied")
		return c.Redirect(302, returnURL)
	}

	tx := c.Value("tx").(*pop.Connection)
	overReading := &models.OverReading{}
	if err := tx.Eager().Find(overReading, c.Param("oid")); err != nil {
		return c.Error(404, err)
	}
	participant := overReading.Participant
	screening := overReading.Screening
	if !(c.Param("pid") == participant.ID.String() && c.Param("sid") == screening.ID.String()) {
		c.Flash().Add("danger", "Not Found")
		return c.Redirect(302, returnURL)
	}

	reason := c.Request().FormValue("reason")

	err := ArchiveMake(c, user.ID, overReading.ID, "OverReading", overReading, reason)
	if err != nil {
		c.Flash().Add("danger", err.Error())
		return c.Redirect(302, returnURL)
	}

	prt := &models.Participant{}
	if err := tx.Find(prt, participant.ID); err != nil {
		return c.Error(404, err)
	}
	prt.Status = "11"
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
