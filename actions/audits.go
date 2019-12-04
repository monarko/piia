package actions

import (
	"encoding/json"
	"strings"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop"
	"github.com/gofrs/uuid"
	jd "github.com/josephburnett/jd/lib"
	"github.com/monarko/piia/models"
)

// MakeAudit function
func MakeAudit(modelType string, modelID uuid.UUID, oldData map[string]interface{}, newData map[string]interface{}, userID uuid.UUID, c buffalo.Context) error {
	audit := &models.Audit{}

	audit.ModelType = modelType
	audit.ModelID = modelID

	audit.OldData = oldData
	audit.NewData = newData

	b, _ := json.Marshal(oldData)
	n, _ := json.Marshal(newData)

	bb, _ := jd.ReadJsonString(string(b))
	nn, _ := jd.ReadJsonString(string(n))
	dff := bb.Diff(nn).Render()

	strs := strings.Split(dff, "\n")
	changes := make(map[string]interface{})
	jsonPath := ""
	temp := make(map[string]interface{})
	for _, s := range strs {
		if strings.HasPrefix(s, "@") {
			// path
			path := strings.TrimPrefix(s, "@ [")
			path = strings.TrimSuffix(path, "]")
			path = strings.Replace(path, "\"", "", -1)
			paths := strings.Split(path, ",")
			jsonPath = strings.Join(paths, "->")
		}

		if strings.HasPrefix(s, "-") {
			temp["from"] = strings.TrimPrefix(s, "- ")
		}

		if strings.HasPrefix(s, "+") {
			temp["to"] = strings.TrimPrefix(s, "+ ")
			changes[jsonPath] = temp
			temp = make(map[string]interface{})
		}
	}

	// fmt.Printf("\n\n%s\n\n", dff)

	// diff, equal := messagediff.DeepDiff(oldData, newData)
	// changes := make(map[string]interface{})
	// if !equal {
	// 	for k, v := range diff.Added {
	// 		changes[k.String()] = v
	// 	}

	// 	for k, v := range diff.Modified {
	// 		changes[k.String()] = v
	// 	}

	// 	for k, v := range diff.Removed {
	// 		changes[k.String()] = v
	// 	}
	// }
	audit.Changes = changes
	audit.UserID = userID

	tx := c.Value("tx").(*pop.Connection)
	_, err := tx.ValidateAndCreate(audit)
	if err != nil {
		return err
	}

	return nil
}
