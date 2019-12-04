package models

import (
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/monarko/piia/helpers/types"

	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

// User object
type User struct {
	ID              uuid.UUID `json:"id" db:"id"`
	CreatedAt       time.Time `json:"-" db:"created_at"`
	UpdatedAt       time.Time `json:"-" db:"updated_at"`
	Email           string    `json:"email" db:"email"`
	Name            string    `json:"name" db:"name"`
	Admin           bool      `json:"-" db:"admin"`
	PasswordHash    string    `json:"-" db:"password_hash"`
	Password        string    `json:"-" db:"-"`
	PasswordConfirm string    `json:"-" db:"-"`
	Mobile          string    `json:"mobile" db:"mobile"`
	Site            string    `json:"site" db:"site"`
	Provider        string    `json:"provider" db:"provider"`
	ProviderID      string    `json:"provider_id" db:"provider_id"`
	Avatar          string    `json:"avatar" db:"avatar"`
	Sites           []string  `json:"sites" db:"-"`

	Permission types.Permission `json:"permissions" db:"permissions"`

	Participants  Participants  `has_many:"participants" json:"-"`
	Screenings    Screenings    `has_many:"screenings" fk_id:"screener_id" json:"-"`
	OverReadings  OverReadings  `has_many:"over_readings" fk_id:"over_reader_id" json:"-"`
	SystemLogs    SystemLogs    `has_many:"system_logs" json:"-"`
	Notifications Notifications `has_many:"notifications" json:"-"`
}

// UserSites returns user sites
func (u User) UserSites() []string {
	sites := strings.Split(u.Site, "")
	return sites
}

// String is not required by pop and may be deleted
func (u User) String() string {
	ju, _ := json.Marshal(u)
	return string(ju)
}

// Users is not required by pop and may be deleted
type Users []User

// String is not required by pop and may be deleted
func (u Users) String() string {
	ju, _ := json.Marshal(u)
	return string(ju)
}

// Create validates and creates a new User.
func (u *User) Create(tx *pop.Connection) (*validate.Errors, error) {
	u.Email = strings.ToLower(u.Email)
	u.Admin = false
	if len(u.Password) > 0 {
		pwdHash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		if err != nil {
			return validate.NewErrors(), errors.WithStack(err)
		}
		u.PasswordHash = string(pwdHash)
	} else {
		u.PasswordHash = ""
	}

	return tx.ValidateAndCreate(u)
}

// Update validates and updates an User.
func (u *User) Update(tx *pop.Connection) (*validate.Errors, error) {
	u.Email = strings.ToLower(u.Email)
	if len(u.Password) > 0 {
		pwdHash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		if err != nil {
			return validate.NewErrors(), errors.WithStack(err)
		}
		u.PasswordHash = string(pwdHash)
	}
	return tx.ValidateAndUpdate(u)
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (u *User) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	valids := make([]validate.Validator, 0)
	valids = append(valids, &validators.StringIsPresent{Field: u.Email, Name: "Email"})
	valids = append(valids, &validators.StringIsPresent{Field: u.Name, Name: "Name"})
	valids = append(valids, &validators.EmailIsPresent{Name: "Email", Field: u.Email})
	valids = append(valids, &validators.StringsMatch{Name: "Password", Field: u.Password, Field2: u.PasswordConfirm, Message: "Passwords do not match."})
	valids = append(valids, &EmailNotTaken{Name: "Email", Field: u.Email, tx: tx})

	u.Site = strings.Join(u.Sites, "")

	if u.Permission.Screening && !u.Permission.StudyCoordinator && !u.Admin && !u.Permission.ReferralTracker && !u.Permission.OverRead {
		valids = append(valids, &validators.StringIsPresent{Field: u.Site, Name: "Site"})
	}

	return validate.Validate(valids...), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (u *User) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	valids := make([]validate.Validator, 0)
	valids = append(valids, &validators.StringIsPresent{Field: u.Email, Name: "Email"})
	valids = append(valids, &validators.StringIsPresent{Field: u.Name, Name: "Name"})
	valids = append(valids, &validators.EmailIsPresent{Name: "Email", Field: u.Email})
	valids = append(valids, &validators.StringsMatch{Name: "Password", Field: u.Password, Field2: u.PasswordConfirm, Message: "Passwords do not match."})

	u.Site = strings.Join(u.Sites, "")

	if u.Permission.Screening && !u.Permission.StudyCoordinator && !u.Admin && !u.Permission.ReferralTracker && !u.Permission.OverRead {
		valids = append(valids, &validators.StringIsPresent{Field: u.Site, Name: "Site"})
	}

	return validate.Validate(valids...), nil
}

// EmailNotTaken check for email
type EmailNotTaken struct {
	Name  string
	Field string
	tx    *pop.Connection
}

// IsValid performs the validation check for unique emails
func (v *EmailNotTaken) IsValid(errors *validate.Errors) {
	query := v.tx.Where("email = ?", v.Field)
	queryUser := User{}
	err := query.First(&queryUser)
	if err == nil {
		// found a user with the same email
		errors.Add(validators.GenerateKey(v.Name), "An account with that email already exists.")
	}
}

// Authorize checks user's password for logging in
func (u *User) Authorize(tx *pop.Connection) error {
	err := tx.Where("email = ?", strings.ToLower(u.Email)).First(u)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			// couldn't find an user with that email address
			return errors.New("user not found")
		}
		return errors.WithStack(err)
	}
	// confirm that the given password matches the hashed password from the db
	err = bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(u.Password))
	if err != nil {
		return errors.New("invalid password")
	}
	return nil
}
