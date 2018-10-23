package models

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/uuid"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

// User object
type User struct {
	ID                         uuid.UUID     `json:"id" db:"id"`
	CreatedAt                  time.Time     `json:"-" db:"created_at"`
	UpdatedAt                  time.Time     `json:"-" db:"updated_at"`
	Username                   string        `json:"username" db:"username"`
	Email                      string        `json:"email" db:"email"`
	Name                       string        `json:"name" db:"name"`
	Admin                      bool          `json:"-" db:"admin"`
	PasswordHash               string        `json:"-" db:"password_hash"`
	Password                   string        `json:"-" db:"-"`
	PasswordConfirm            string        `json:"-" db:"-"`
	Participants               Participants  `has_many:"participants" json:"-"`
	Screenings                 Screenings    `has_many:"screenings" fk_id:"screener_id" json:"-"`
	OverReadings               OverReadings  `has_many:"over_readings" fk_id:"over_reader_id" json:"-"`
	PermissionScreening        bool          `json:"-" db:"permission_screening"`
	PermissionOverRead         bool          `json:"-" db:"permission_overread"`
	PermissionStudyCoordinator bool          `json:"-" db:"permission_study_coordinator"`
	SystemLogs                 SystemLogs    `has_many:"system_logs" json:"-"`
	Mobile                     string        `json:"mobile" db:"mobile"`
	Site                       string        `json:"site" db:"site"`
	Notifications              Notifications `has_many:"notifications" json:"-"`
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
	pwdHash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return validate.NewErrors(), errors.WithStack(err)
	}
	u.PasswordHash = string(pwdHash)
	return tx.ValidateAndCreate(u)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (u *User) Validate(tx *pop.Connection) (*validate.Errors, error) {
	valids := make([]validate.Validator, 0)
	valids = append(valids, &validators.StringIsPresent{Field: u.Username, Name: "Username"})
	valids = append(valids, &validators.StringIsPresent{Field: u.Email, Name: "Email"})
	valids = append(valids, &validators.StringIsPresent{Field: u.Name, Name: "Name"})
	valids = append(valids, &validators.EmailIsPresent{Name: "Email", Field: u.Email})
	valids = append(valids, &validators.StringsMatch{Name: "Password", Field: u.Password, Field2: u.PasswordConfirm, Message: "Passwords do not match."})
	valids = append(valids, &UsernameNotTaken{Name: "Username", Field: u.Username, tx: tx})
	valids = append(valids, &EmailNotTaken{Name: "Email", Field: u.Email, tx: tx})

	if u.PermissionScreening {
		valids = append(valids, &validators.StringIsPresent{Field: u.Site, Name: "Site"})
	}

	return validate.Validate(valids...), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (u *User) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (u *User) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// UsernameNotTaken Check
type UsernameNotTaken struct {
	Name  string
	Field string
	tx    *pop.Connection
}

// IsValid checks if username is valid or not
func (v *UsernameNotTaken) IsValid(errors *validate.Errors) {
	query := v.tx.Where("username = ?", v.Field)
	queryUser := User{}
	err := query.First(&queryUser)
	if err == nil {
		// found a user with same username
		errors.Add(validators.GenerateKey(v.Name), fmt.Sprintf("The username %s is not available.", v.Field))
	}
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
