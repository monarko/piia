package actions

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/middleware"
	"github.com/gobuffalo/buffalo/middleware/csrf"
	"github.com/gobuffalo/buffalo/middleware/i18n"
	"github.com/gobuffalo/buffalo/middleware/ssl"
	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/packr"
	"github.com/monarko/piia/models"
	"github.com/unrolled/secure"

	"github.com/markbates/goth/gothic"
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

		if ENV == "production" {
			setErrorHandler(app)
		}

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
		app.Use(SetCurrentLang)

		app.GET("/", HomeHandler)

		// app.Resource("/users", UsersResource{})
		auth := app.Group("/users")
		// auth.GET("/register", UsersRegisterGet)
		// auth.POST("/register", UsersRegisterPost)
		auth.GET("/login", UsersLoginGet)
		auth.POST("/login", UsersLoginPost)
		auth.GET("/logout", UsersLogout)

		auth.GET("/", AdminRequired(UsersIndex))
		auth.GET("/index", AdminRequired(UsersIndex))
		auth.GET("/create", AdminRequired(UsersCreateGet))
		auth.POST("/create", AdminRequired(UsersCreatePost))
		auth.GET("/edit/{uid}", AdminRequired(UsersEditGet)).Name("usersEditPath")
		auth.POST("/edit/{uid}", AdminRequired(UsersEditPost)).Name("usersEditPath")

		participants := app.Group("/participants")
		participants.Use(LoginRequired)
		participants.Use(ScreeningPermissionRequired)
		participants.GET("/", ParticipantsIndex)
		participants.GET("/index", ParticipantsIndex)
		participants.GET("/create", ParticipantsCreateGet)
		participants.POST("/create", ParticipantsCreatePost)
		participants.GET("/edit/{pid}", ParticipantsEditGet).Name("participantsEditPath")
		participants.POST("/edit/{pid}", ParticipantsEditPost).Name("participantsEditPath")
		participants.GET("/{pid}", ParticipantsDetail)
		// participants.GET("/delete", ParticipantsDelete)
		// participants.GET("/detail", ParticipantsDetail)

		cases := app.Group("/cases")
		cases.Use(LoginRequired)
		// cases.Use(OverReadingPermissionRequired)
		cases.GET("/", CasesIndex)
		cases.GET("/index", CasesIndex)

		screenings := participants.Group("/{pid}/screenings")
		screenings.Use(ScreeningPermissionRequired)
		// screenings.GET("/index", ScreeningsIndex)
		screenings.GET("/create", ScreeningsCreateGet)
		screenings.POST("/create", ScreeningsCreatePost)
		screenings.GET("/edit/{sid}", ScreeningsEditGet).Name("participantScreeningsEditPath")
		screenings.POST("/edit/{sid}", ScreeningsEditPost).Name("participantScreeningsEditPath")

		overReadings := screenings.Group("/{sid}/overreadings")
		overReadings.Middleware.Skip(ScreeningPermissionRequired, OverReadingsCreateGet, OverReadingsCreatePost)
		overReadings.Use(OverReadingPermissionRequired)
		overReadings.GET("/create", OverReadingsCreateGet)
		overReadings.POST("/create", OverReadingsCreatePost)
		overReadings.GET("/{oid}", OverReadingsDetails)

		screenings.POST("/{sid}/appointmentdone", UpdateReferredMessage).Name("participantsAppointmentPath")

		reports := app.Group("/reports")
		reports.Use(LoginRequired)
		reports.Use(StudyCoordinatorPermissionRequired)
		reports.GET("/", ReportsIndex)
		reports.GET("/index", ReportsIndex)

		referrals := app.Group("/referrals")
		referrals.Use(LoginRequired)
		referrals.Use(ReferralTrackerPermissionRequired)
		referrals.GET("/", ReferralsIndex)
		referrals.GET("/index", ReferralsIndex)
		referrals.GET("/participants/{pid}", ReferralsParticipantsGet)

		notifications := app.Group("/notifications")
		notifications.Use(LoginRequired)
		notifications.GET("/", NotificationsIndex)
		notifications.GET("/index", NotificationsIndex)

		// app.Resource("/system_logs", SystemLogsResource{})
		logs := app.Group("/logs")
		logs.Use(AdminRequired)
		logs.GET("/", SystemLogsIndex)
		logs.GET("/index", SystemLogsIndex)

		app.GET("/errors/{status}", ErrorsDefault)

		app.GET("/switch", ChangeLanguage)
		app.POST("/notifications", ScreeningPermissionRequired(ChangeNotificationStatus))

		authGoth := app.Group("/auth")
		authGoth.GET("/{provider}", buffalo.WrapHandlerFunc(gothic.BeginAuthHandler))
		authGoth.GET("/{provider}/callback", AuthCallback)

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

func setErrorHandler(app *buffalo.App) {
	app.ErrorHandlers[400] = customErrorHandler()
	app.ErrorHandlers[401] = customErrorHandler()
	app.ErrorHandlers[403] = customErrorHandler()
	app.ErrorHandlers[404] = customErrorHandler()
	app.ErrorHandlers[405] = customErrorHandler()
	app.ErrorHandlers[408] = customErrorHandler()
	app.ErrorHandlers[422] = customErrorHandler()

	app.ErrorHandlers[500] = customErrorHandler()
	app.ErrorHandlers[501] = customErrorHandler()
	app.ErrorHandlers[502] = customErrorHandler()
	app.ErrorHandlers[503] = customErrorHandler()
	app.ErrorHandlers[504] = customErrorHandler()
	app.ErrorHandlers[505] = customErrorHandler()
}

func customErrorHandler() buffalo.ErrorHandler {
	return func(status int, err error, c buffalo.Context) error {
		ct := c.Request().Header.Get("Content-Type")

		switch strings.ToLower(ct) {
		case "application/json", "text/json", "json":
			c.Logger().Error(err)
			c.Response().WriteHeader(status)

			msg := fmt.Sprintf("%+v", err)
			return json.NewEncoder(c.Response()).Encode(map[string]interface{}{
				"error": msg,
				"code":  status,
			})
		default:
			tmpl := "default"
			switch status {
			case 401:
				tmpl = "401"
			case 403:
				tmpl = "403"
			case 404:
				tmpl = "404"
			}
			return c.Redirect(302, "/errors/"+tmpl)
		}
	}
}
