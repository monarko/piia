package models

import (
	"encoding/json"
	"time"

	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/validate"
	"github.com/gofrs/uuid"
)

var (
	dependency = map[string]map[string]string{
		"OverReading":     map[string]string{"User": "over_reader_id", "Participant": "participant_id", "Screening": "screening_id"},
		"Notification":    map[string]string{"User": "from_user_id", "Participant": "participant_id", "Screening": "screening_id"},
		"ReferredMessage": map[string]string{"User": "user_id", "Participant": "participant_id", "Screening": "screening_id"},
		"Screening":       map[string]string{"User": "screener_id", "Participant": "participant_id"},
		"Participant":     map[string]string{"User": "author_id"},
	}
)

// Archive type
type Archive struct {
	ID          uuid.UUID `json:"id" db:"id"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	ArchiveType string    `json:"archive_type" db:"archive_type"`
	ModelID     uuid.UUID `json:"model_id" db:"model_id"`
	Archiver    User      `belongs_to:"user" json:"archiver"`
	ArchiverID  uuid.UUID `json:"archiver_id" db:"archiver_id"`
	Data        []byte    `json:"data" db:"data"`
	Dependency  Mapping   `json:"dependency" db:"dependency"`
	Reason      string    `json:"reason" db:"reason"`
}

// Archives is not required by pop and may be deleted
type Archives []Archive

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (a *Archive) Validate(tx *pop.Connection) (*validate.Errors, error) {
	dep := dependency[a.ArchiveType]
	d := make(map[string]interface{})
	m := make(map[string]interface{})
	json.Unmarshal(a.Data, &m)
	for k, v := range dep {
		d[k] = map[string]string{v: m[v].(string)}
	}
	a.Dependency = d

	return validate.NewErrors(), nil
}
