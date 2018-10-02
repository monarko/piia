package actions

import (
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/middleware"
	"github.com/gobuffalo/buffalo/middleware/ssl"
	"github.com/gobuffalo/envy"
	"github.com/unrolled/secure"

	"github.com/gobuffalo/buffalo/middleware/csrf"
	"github.com/gobuffalo/buffalo/middleware/i18n"
	"github.com/gobuffalo/packr"
	"github.com/monarko/piia/models"
)

// ENV is used to help switch settings based on where the
// application is being run. Default is "development".
var ENV = envy.Get("GO_ENV", "development")
var app *buffalo.App

// T is for translation
var T *i18n.Translator

// App is where all routes and middleware for buffalo
// should be defined. This is the nerve center of your
// application.
func App() *buffalo.App {
	if app == nil {
		app = buffalo.New(buffalo.Options{
			Env:         ENV,
			SessionName: "_piia_session",
		})
		// Automatically redirect to SSL
		app.Use(forceSSL())

		if ENV == "development" {
			app.Use(middleware.ParameterLogger)
		}

		// Protect against CSRF attacks. https://www.owasp.org/index.php/Cross-Site_Request_Forgery_(CSRF)
		// Remove to disable this.
		app.Use(csrf.New)

		// Wraps each request in a transaction.
		//  c.Value("tx").(*pop.PopTransaction)
		// Remove to disable this.
		app.Use(middleware.PopTransaction(models.DB))

		// Setup and use translations:
		app.Use(translations())
		app.Use(SetCurrentUser)

		app.GET("/", HomeHandler)

		// app.Resource("/users", UsersResource{})
		auth := app.Group("/users")
		auth.GET("/register", UsersRegisterGet)
		auth.POST("/register", UsersRegisterPost)
		auth.GET("/login", UsersLoginGet)
		auth.POST("/login", UsersLoginPost)
		auth.GET("/logout", UsersLogout)

		auth.GET("/index", AdminRequired(UsersIndex))

		participants := app.Group("/participants")
		participants.Use(LoginRequired)
		participants.GET("/index", ParticipantsIndex)
		participants.GET("/create", ParticipantsCreateGet)
		participants.POST("/create", ParticipantsCreatePost)
		participants.GET("/edit/{pid}", ParticipantsEditGet).Name("participantsEditPath")
		participants.POST("/edit/{pid}", ParticipantsEditPost).Name("participantsEditPath")
		// participants.GET("/delete", ParticipantsDelete)
		// participants.GET("/detail", ParticipantsDetail)

		screenings := participants.Group("/{pid}/screenings")
		screenings.GET("/index", ScreeningsIndex)
		screenings.GET("/create", ScreeningsCreateGet)
		screenings.POST("/create", ScreeningsCreatePost)

		overReadings := participants.Group("/{pid}/overreadings")
		overReadings.GET("/index", OverReadingsIndex)
		overReadings.GET("/create", OverReadingsCreateGet)
		overReadings.POST("/create", OverReadingsCreatePost)

		app.ServeFiles("/", assetsBox) // serve files from the public directory
	}

	return app
}

// translations will load locale files, set up the translator `actions.T`,
// and will return a middleware to use to load the correct locale for each
// request.
// for more information: https://gobuffalo.io/en/docs/localization
func translations() buffalo.MiddlewareFunc {
	var err error
	if T, err = i18n.New(packr.NewBox("../locales"), "en-US"); err != nil {
		app.Stop(err)
	}
	return T.Middleware()
}

// forceSSL will return a middleware that will redirect an incoming request
// if it is not HTTPS. "http://example.com" => "https://example.com".
// This middleware does **not** enable SSL. for your application. To do that
// we recommend using a proxy: https://gobuffalo.io/en/docs/proxy
// for more information: https://github.com/unrolled/secure/
func forceSSL() buffalo.MiddlewareFunc {
	return ssl.ForceSSL(secure.Options{
		SSLRedirect:     ENV == "production",
		SSLProxyHeaders: map[string]string{"X-Forwarded-Proto": "https"},
	})
}
