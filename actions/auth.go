package actions

import (
	"fmt"
	"os"
	"strings"

	"github.com/monarko/piia/models"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
	"github.com/pkg/errors"
)

func init() {
	gothic.Store = App().SessionStore

	goth.UseProviders(
		google.New(os.Getenv("GOOGLE_KEY"), os.Getenv("GOOGLE_SECRET"), fmt.Sprintf("%s%s", App().Host, "/auth/google/callback")),
	)
}

// AuthCallback functio
func AuthCallback(c buffalo.Context) error {
	guser, err := gothic.CompleteUserAuth(c.Response(), c.Request())
	if err != nil {
		return errors.WithStack(err)
		// return c.Error(401, err)
	}
	tx := c.Value("tx").(*pop.Connection)
	q := tx.Where("provider = ? and provider_id = ?", "google", guser.UserID)
	exists, err := q.Exists("users")
	if err != nil {
		return errors.WithStack(err)
	}

	u := &models.User{}
	if exists {
		err = q.First(u)
		if err != nil {
			return errors.WithStack(err)
		}
		// User found and validated
	} else {
		p := tx.Where("email = ?", guser.Email)
		pExists, err := p.Exists("users")
		if err != nil {
			return errors.WithStack(err)
		}

		if pExists {
			err = p.First(u)
			if err != nil {
				return errors.WithStack(err)
			}
			if len(strings.TrimSpace(u.Name)) == 0 {
				u.Name = guser.Name
			}
			u.Provider = guser.Provider
			u.ProviderID = guser.UserID
			u.Avatar = guser.AvatarURL
			err = tx.Update(u)
			if err != nil {
				return errors.WithStack(err)
			}
			// User found but not activated
		} else {
			// If there are no errors set a success message
			c.Flash().Add("danger", "You are not authorized.")
			// and redirect to the index page
			return c.Redirect(302, "/")
			// User not found
		}
	}

	c.Session().Set("current_user_id", u.ID)
	err = c.Session().Save()
	if err != nil {
		return errors.WithStack(err)
	}

	logErr := InsertLog("login", "User logged in", "", "", "", u.ID, c)
	if logErr != nil {
		return errors.WithStack(logErr)
	}

	redirectPath := "/"

	if u.PermissionOverRead && !u.PermissionScreening && !u.PermissionStudyCoordinator {
		redirectPath = "/cases/index"
	} else if !u.Admin {
		redirectPath = "/participants/index"
	}

	return c.Redirect(302, redirectPath)
}
