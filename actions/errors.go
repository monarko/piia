package actions

import (
	"github.com/gobuffalo/buffalo"
	"github.com/monarko/piia/models"
)

// ErrorsDefault returns the error page
func ErrorsDefault(c buffalo.Context) error {
	user := c.Value("current_user").(*models.User)
	InsertLog("error", "Error", c.Err().Error(), "", "", user.ID, c)
	tmpl := "/errors/" + c.Param("status") + ".html"
	return c.Render(200, r.HTML(tmpl, "application-non-logged-in.html"))
}
