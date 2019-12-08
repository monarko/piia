package grifts

import (
	"fmt"
	"strings"

	"github.com/markbates/grift/grift"
	"github.com/monarko/piia/models"
	"golang.org/x/crypto/bcrypt"
)

var _ = grift.Namespace("user", func() {

	grift.Desc("superadmin:make", "Make an user superadmin")
	grift.Add("superadmin:make", func(c *grift.Context) error {
		messages := make([]string, 0)
		if len(c.Args) >= 1 {
			for _, e := range c.Args {
				email := strings.TrimSpace(e)
				q := tx.Where("email = ?", email)
				exists, err := q.Exists("users")

				if !exists {
					messages = append(messages, "=> "+email+" email not found")
				} else {
					user := &models.User{}
					err = q.First(user)
					if err != nil {
						messages = append(messages, "=> "+email+" updating failed")
					} else {
						user.Admin = true
						err = tx.Update(user)
						if err != nil {
							messages = append(messages, "=> "+email+" updating failed")
						} else {
							messages = append(messages, "=> "+email+" has been given superadmin previleges")
						}
					}
				}
			}
		}

		if len(messages) == 0 {
			messages = append(messages, "--- No valid user found to make them superadmin ---")
		}

		fmt.Println(strings.Join(messages, "\n"))

		return nil
	})

})

var _ = grift.Namespace("user", func() {

	grift.Desc("superadmin:revoke", "Revoke an user's superadmin previlege")
	grift.Add("superadmin:revoke", func(c *grift.Context) error {
		messages := make([]string, 0)
		if len(c.Args) >= 1 {
			for _, e := range c.Args {
				email := strings.TrimSpace(e)
				q := tx.Where("email = ?", email)
				exists, err := q.Exists("users")

				if !exists {
					messages = append(messages, "=> "+email+" email not found")
				} else {
					user := &models.User{}
					err = q.First(user)
					if err != nil {
						messages = append(messages, "=> "+email+" updating failed")
					} else {
						user.Admin = true
						err = tx.Update(user)
						if err != nil {
							messages = append(messages, "=> "+email+" updating failed")
						} else {
							messages = append(messages, "=> "+email+" has been revoked of superadmin previleges")
						}
					}
				}
			}
		}

		if len(messages) == 0 {
			messages = append(messages, "No valid user found to make them superadmin")
		}

		fmt.Println(strings.Join(messages, "\n"))

		return nil
	})

})

var _ = grift.Namespace("user", func() {

	grift.Desc("superadmin:create", "Create an user with superadmin previleges")
	grift.Add("superadmin:create", func(c *grift.Context) error {
		messages := make([]string, 0)
		if len(c.Args) >= 1 {
			email := strings.TrimSpace(c.Args[0])
			name := strings.TrimSpace(c.Args[1])
			password := strings.TrimSpace(c.Args[2])

			user := &models.User{}
			user.Admin = true
			user.Name = name
			user.Email = email
			user.Permission.OverRead = false
			user.Permission.Screening = false
			user.Permission.StudyCoordinator = false
			user.Permission.ReferralTracker = false
			if len(password) > 0 {
				pwdHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
				if err != nil {
					messages = append(messages, "=> \""+name+"\" creation failed")
					messages = append(messages, "=> "+err.Error())
					fmt.Println(strings.Join(messages, "\n"))
				}
				user.PasswordHash = string(pwdHash)
			}

			err := tx.Create(user)
			if err != nil {
				messages = append(messages, "=> \""+name+"\" creation failed")
				messages = append(messages, "=> "+err.Error())
			} else {
				messages = append(messages, "=> \""+name+"\" is created and has been given superadmin previleges")
			}
		}

		if len(messages) == 0 {
			messages = append(messages, "--- No valid input given ---")
		}

		fmt.Println(strings.Join(messages, "\n"))

		return nil
	})

})
