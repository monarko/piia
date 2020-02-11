package actions

import (
	"net/url"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/validate"
	"github.com/pkg/errors"

	"github.com/monarko/piia/helpers"
	"github.com/monarko/piia/mailers"
	"github.com/monarko/piia/models"
)

// UsersIndex default implementation.
func UsersIndex(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	users := &models.Users{}
	// Paginate results. Params "page" and "per_page" control pagination.
	// Default values are "page=1" and "per_page=20".
	q := tx.PaginateFromParams(c.Value("paginateParam").(url.Values)).Order("created_at DESC")
	// Retrieve all Posts from the DB
	if err := q.All(users); err != nil {
		return errors.WithStack(err)
	}
	// Make posts available inside the html template
	c.Set("users", users)
	// Add the paginator to the context so it can be used in the template.
	c.Set("pagination", q.Paginator)

	b := c.Value("breadcrumb").(helpers.Breadcrumbs)
	b = append(b, helpers.Breadcrumb{Title: "Users", Path: "/users/index"})
	c.Set("breadcrumb", b)

	return c.Render(200, r.HTML("users/index.html"))
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
	return c.Redirect(302, "/")
}

// UsersCreateGet returns the form
func UsersCreateGet(c buffalo.Context) error {
	c.Set("user", &models.User{})

	b := c.Value("breadcrumb").(helpers.Breadcrumbs)
	b = append(b, helpers.Breadcrumb{Title: "Users", Path: "/users/index"})
	b = append(b, helpers.Breadcrumb{Title: "New User", Path: "/users/create"})
	c.Set("breadcrumb", b)

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

	b := c.Value("breadcrumb").(helpers.Breadcrumbs)
	b = append(b, helpers.Breadcrumb{Title: "Users", Path: "/users/index"})
	b = append(b, helpers.Breadcrumb{Title: "New User", Path: "/users/create"})
	c.Set("breadcrumb", b)

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
	user.Sites = user.UserSites()
	c.Set("user", user)

	b := c.Value("breadcrumb").(helpers.Breadcrumbs)
	b = append(b, helpers.Breadcrumb{Title: "Users", Path: "/users/index"})
	b = append(b, helpers.Breadcrumb{Title: "Edit User", Path: "/users/edit"})
	c.Set("breadcrumb", b)

	return c.Render(200, r.HTML("users/edit.html"))
}

// UsersEditPost returns the form
func UsersEditPost(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	user := &models.User{}
	if err := tx.Find(user, c.Param("uid")); err != nil {
		return c.Error(404, err)
	}

	b := c.Value("breadcrumb").(helpers.Breadcrumbs)
	b = append(b, helpers.Breadcrumb{Title: "Users", Path: "/users/index"})
	b = append(b, helpers.Breadcrumb{Title: "Edit User", Path: "/users/edit"})
	c.Set("breadcrumb", b)

	user.Admin = false
	user.Permission.StudyCoordinator = false
	user.Permission.StudyTeamMember = false
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
