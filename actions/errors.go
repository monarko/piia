package actions

import "github.com/gobuffalo/buffalo"

// ErrorsDefault returns the error page
func ErrorsDefault(c buffalo.Context) error {
	tmpl := "/errors/" + c.Param("status") + ".html"
	return c.Render(200, r.HTML(tmpl, "application-non-logged-in.html"))
}
