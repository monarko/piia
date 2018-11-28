package actions

import (
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/validate"
	"github.com/monarko/piia/mailers"
	"github.com/monarko/piia/models"
	"github.com/pkg/errors"
)

// UsersIndex default implementation.
func UsersIndex(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	users := &models.Users{}
	// Paginate results. Params "page" and "per_page" control pagination.
	// Default values are "page=1" and "per_page=20".
	q := tx.PaginateFromParams(c.Params())
	// Retrieve all Posts from the DB
	if err := q.All(users); err != nil {
		return errors.WithStack(err)
	}
	// Make posts available inside the html template
	c.Set("users", users)
	// Add the paginator to the context so it can be used in the template.
	c.Set("pagination", q.Paginator)
	breadcrumbMap := make(map[string]interface{})
	breadcrumbMap["Users"] = "/users/index"
	c.Set("breadcrumbMap", breadcrumbMap)
	return c.Render(200, r.HTML("users/index.html"))
}

// UsersRegisterGet displays a register form
func UsersRegisterGet(c buffalo.Context) error {
	// Make user available inside the html template
	c.Set("user", &models.User{})
	return c.Render(200, r.HTML("users/register.html", "application-non-logged-in.html"))
}

// UsersRegisterPost adds a User to the DB. This function is mapped to the
// path POST /accounts/register
func UsersRegisterPost(c buffalo.Context) error {
	// Allocate an empty User
	user := &models.User{}
	// Bind user to the html form elements
	if err := c.Bind(user); err != nil {
		return errors.WithStack(err)
	}
	// Get the DB connection from the context
	tx := c.Value("tx").(*pop.Connection)
	// Validate the data from the html form
	verrs, err := user.Create(tx)
	if err != nil {
		return errors.WithStack(err)
	}
	if verrs.HasAny() {
		// Make user available inside the html template
		c.Set("user", user)
		// Make the errors available inside the html template
		c.Set("errors", verrs.Errors)
		// Render again the register.html template that the user can
		// correct the input.
		return c.Render(422, r.HTML("users/register.html", "application-non-logged-in.html"))
	}
	// If there are no errors set a success message
	c.Flash().Add("success", "Registered successfully. You can login now.")
	// and redirect to the home page
	return c.Redirect(302, "/")
}

// UsersLoginGet displays a login form
func UsersLoginGet(c buffalo.Context) error {
	return c.Render(200, r.HTML("users/login.html", "application-non-logged-in.html"))
}

// UsersLoginPost logs in a user.
func UsersLoginPost(c buffalo.Context) error {
	user := &models.User{}
	// Bind the user to the html form elements
	if err := c.Bind(user); err != nil {
		return errors.WithStack(err)
	}
	tx := c.Value("tx").(*pop.Connection)
	err := user.Authorize(tx)
	if err != nil {
		c.Set("user", user)
		verrs := validate.NewErrors()
		verrs.Add("Login", "Invalid email or password.")
		c.Set("errors", verrs.Errors)
		return c.Render(422, r.HTML("users/login.html", "application-non-logged-in.html"))
	}
	logErr := InsertLog("login", "User logged in", "", "", "", user.ID, c)
	if logErr != nil {
		return errors.WithStack(logErr)
	}
	c.Session().Set("current_user_id", user.ID)
	c.Flash().Add("success", "Welcome back!")

	redirectPath := "/"

	if user.Permission.OverRead && !user.Permission.Screening && !user.Permission.StudyCoordinator {
		redirectPath = "/cases/index"
	} else if !user.Admin {
		redirectPath = "/participants/index"
	}

	return c.Redirect(302, redirectPath)
}

// UsersLogout clears the session and logs out the user.
func UsersLogout(c buffalo.Context) error {
	user := c.Value("current_user").(*models.User)
	logErr := InsertLog("logout", "User logged out", "", "", "", user.ID, c)
	if logErr != nil {
		return errors.WithStack(logErr)
	}
	c.Session().Clear()
	c.Flash().Add("success", "Goodbye!")
	return c.Redirect(302, "/")
}

// UsersCreateGet returns the form
func UsersCreateGet(c buffalo.Context) error {
	c.Set("user", &models.User{})

	breadcrumbMap := make(map[string]interface{})
	breadcrumbMap["Users"] = "/users/index"
	breadcrumbMap["New User"] = "/users/create"
	c.Set("breadcrumbMap", breadcrumbMap)
	return c.Render(200, r.HTML("users/create.html"))
}

// UsersCreatePost returns the form
func UsersCreatePost(c buffalo.Context) error {
	// Allocate an empty User
	user := &models.User{}
	// Bind user to the html form elements
	if err := c.Bind(user); err != nil {
		return errors.WithStack(err)
	}
	// Get the DB connection from the context
	tx := c.Value("tx").(*pop.Connection)
	// Validate the data from the html form
	verrs, err := user.Create(tx)
	if err != nil {
		return errors.WithStack(err)
	}
	if verrs.HasAny() {
		// Make user available inside the html template
		c.Set("user", user)
		// Make the errors available inside the html template
		c.Set("errors", verrs.Errors)
		// Render again the register.html template that the user can
		// correct the input.
		return c.Render(422, r.HTML("users/create.html"))
	}

	currentUser := c.Value("current_user").(*models.User)
	logErr := InsertLog("create", "User created the user: "+user.Name, "", user.ID.String(), "user", currentUser.ID, c)
	if logErr != nil {
		return errors.WithStack(logErr)
	}

	newUserEmail := &mailers.EmailDetails{}
	newUserEmail.To = []string{user.Email}
	newUserEmail.Subject = "Welcome to PIIA (peer)"
	newUserEmail.Data = map[string]interface{}{
		"name":  user.Name,
		"email": user.Email,
		"root":  envy.Get("APP_HOST", "http://127.0.0.1"),
		"link":  envy.Get("APP_HOST", "http://127.0.0.1") + "/auth/google",
	}

	// return c.Render(200, r.JSON(newUserEmail))

	// mailers.SendWelcomeEmail(newUserEmail)
	err = newUserEmail.SendMessage(c)
	if err != nil {
		return errors.WithStack(err)
		// c.Flash().Add("danger", err.Error())
	}

	c.Flash().Add("success", "User is created.")
	return c.Redirect(302, "/users/index")
}

// UsersEditGet returns the form
func UsersEditGet(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	user := &models.User{}
	if err := tx.Find(user, c.Param("uid")); err != nil {
		return c.Error(404, err)
	}
	c.Set("user", user)

	breadcrumbMap := make(map[string]interface{})
	breadcrumbMap["Users"] = "/users/index"
	breadcrumbMap["Edit User"] = "/users/edit"
	c.Set("breadcrumbMap", breadcrumbMap)
	return c.Render(200, r.HTML("users/edit.html"))
}

// UsersEditPost returns the form
func UsersEditPost(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	user := &models.User{}
	if err := tx.Find(user, c.Param("uid")); err != nil {
		return c.Error(404, err)
	}
	user.Admin = false
	user.Permission.StudyCoordinator = false
	user.Permission.Screening = false
	user.Permission.OverRead = false
	user.Permission.ReferralTracker = false
	if err := c.Bind(user); err != nil {
		return errors.WithStack(err)
	}
	verrs, err := user.Update(tx)
	if err != nil {
		return errors.WithStack(err)
	}

	if verrs.HasAny() {
		c.Set("user", user)
		c.Set("errors", verrs.Errors)
		breadcrumbMap := make(map[string]interface{})
		breadcrumbMap["Users"] = "/users/index"
		breadcrumbMap["Edit User"] = "/users/edit"
		c.Set("breadcrumbMap", breadcrumbMap)
		return c.Render(422, r.HTML("users/edit.html"))
	}

	currentUser := c.Value("current_user").(*models.User)
	logErr := InsertLog("update", "User updated the user: "+user.Name, "", user.ID.String(), "user", currentUser.ID, c)
	if logErr != nil {
		return errors.WithStack(logErr)
	}

	// If there are no errors set a success message
	c.Flash().Add("success", "User is updated.")
	// and redirect to the home page
	return c.Redirect(302, "/users/index")
}

// SetCurrentUser attempts to find a user based on the current_user_id
// in the session. If one is found it is set on the context.
func SetCurrentUser(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		if uid := c.Session().Get("current_user_id"); uid != nil {
			u := &models.User{}
			tx := c.Value("tx").(*pop.Connection)
			err := tx.Find(u, uid)
			if err == nil {
				c.Set("current_user", u)
				if u.Admin {
					c.Set("admin_user", u.Admin)
				}
			}
		}
		return next(c)
	}
}

// AdminRequired requires a user to be logged in and to be an admin before accessing a route.
func AdminRequired(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		user, ok := c.Value("current_user").(*models.User)
		if ok && user.Admin {
			return next(c)
		}
		c.Flash().Add("danger", "You are not authorized to view that page.")
		return c.Redirect(302, "/")
	}
}

// LoginRequired requires a user to be logged in and to be an admin before accessing a route.
func LoginRequired(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		_, ok := c.Value("current_user").(*models.User)
		if ok {
			return next(c)
		}
		c.Flash().Add("danger", "You are not authorized to view that page.")
		return c.Redirect(302, "/")
	}
}

// ScreeningPermissionRequired requires a user to be logged in and to be an admin before accessing a route.
func ScreeningPermissionRequired(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		user, ok := c.Value("current_user").(*models.User)
		if ok {
			if user.Admin || user.Permission.Screening || user.Permission.StudyCoordinator {
				return next(c)
			}
		}
		c.Flash().Add("danger", "You are not authorized to view that page.")
		return c.Redirect(302, "/")
	}
}

// OverReadingPermissionRequired requires a user to be logged in and to be an admin before accessing a route.
func OverReadingPermissionRequired(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		user, ok := c.Value("current_user").(*models.User)
		if ok {
			if user.Admin || user.Permission.OverRead || user.Permission.StudyCoordinator {
				return next(c)
			}
		}
		c.Flash().Add("danger", "You are not authorized to view that page.")
		return c.Redirect(302, "/")
	}
}

// ReferralTrackerPermissionRequired requires a user to be logged in and to be an admin before accessing a route.
func ReferralTrackerPermissionRequired(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		user, ok := c.Value("current_user").(*models.User)
		if ok {
			if user.Admin || user.Permission.ReferralTracker || user.Permission.StudyCoordinator {
				return next(c)
			}
		}
		c.Flash().Add("danger", "You are not authorized to view that page.")
		return c.Redirect(302, "/")
	}
}

// StudyCoordinatorPermissionRequired requires a user to be logged in and to be an admin before accessing a route.
func StudyCoordinatorPermissionRequired(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		user, ok := c.Value("current_user").(*models.User)
		if ok {
			if user.Admin || user.Permission.StudyCoordinator {
				return next(c)
			}
		}
		c.Flash().Add("danger", "You are not authorized to view that page.")
		return c.Redirect(302, "/")
	}
}
