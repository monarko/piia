package actions

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo-pop/v2/pop/popmw"
	"github.com/gobuffalo/envy"
	csrf "github.com/gobuffalo/mw-csrf"
	forcessl "github.com/gobuffalo/mw-forcessl"
	i18n "github.com/gobuffalo/mw-i18n"
	paramlogger "github.com/gobuffalo/mw-paramlogger"
	"github.com/gobuffalo/packr/v2"
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
			app.Use(paramlogger.ParameterLogger)
		}

		// Protect against CSRF attacks. https://www.owasp.org/index.php/Cross-Site_Request_Forgery_(CSRF)
		// Remove to disable this.
		app.Use(csrf.New)

		// Wraps each request in a transaction.
		//  c.Value("tx").(*pop.PopTransaction)
		// Remove to disable this.
		app.Use(popmw.Transaction(models.DB))

		// Setup and use translations:
		app.Use(translations())
		app.Use(SetSiteStatus)
		app.Use(SetCurrentUser)
		app.Use(SetCurrentLang)
		app.Use(SetCurrentSite)
		app.Use(SetBreadcrumb)

		sitePerm, _, _ := SiteStatus()

		app.GET("/", HomeHandler)

		auth := app.Group("/users")
		auth.GET("/login", UsersLoginGet)
		auth.POST("/login", UsersLoginPost)
		auth.GET("/logout", UsersLogout)

		auth.GET("/", AdminRequired(UsersIndex))
		auth.GET("/index", AdminRequired(UsersIndex))
		auth.GET("/create", SitePermissionRequired(AdminRequired(UsersCreateGet), sitePerm.User.Create)).Name("usersCreatePath")
		auth.POST("/create", SitePermissionRequired(AdminRequired(UsersCreatePost), sitePerm.User.Create)).Name("usersCreatePath")
		auth.GET("/edit/{uid}", SitePermissionRequired(AdminRequired(UsersEditGet), sitePerm.User.Edit)).Name("usersEditPath")
		auth.POST("/edit/{uid}", SitePermissionRequired(AdminRequired(UsersEditPost), sitePerm.User.Edit)).Name("usersEditPath")

		participants := app.Group("/participants")
		participants.Use(LoginRequired)
		participants.Use(ScreeningPermissionRequired)
		participants.Middleware.Skip(ScreeningPermissionRequired, ParticipantsIndex, ParticipantsDetail)
		participants.GET("/", ParticipantsIndex)
		participants.GET("/index", ParticipantsIndex)
		participants.GET("/create", SitePermissionRequired(ParticipantsCreateGet, sitePerm.Participant.Create))
		participants.POST("/create", SitePermissionRequired(ParticipantsCreatePost, sitePerm.Participant.Create))
		participants.GET("/edit/{pid}", SitePermissionRequired(ParticipantsEditGet, sitePerm.Participant.Edit)).Name("participantsEditPath")
		participants.POST("/edit/{pid}", SitePermissionRequired(ParticipantsEditPost, sitePerm.Participant.Edit)).Name("participantsEditPath")
		participants.DELETE("/delete/{pid}", SitePermissionRequired(AdminRequired(ParticipantsDestroy), sitePerm.Participant.Archive)).Name("participantsDeletePath")
		participants.GET("/{pid}", ParticipantsDetail).Name("participantPath")

		cases := app.Group("/cases")
		cases.Use(LoginRequired)
		cases.GET("/", CasesIndex)
		cases.GET("/index", CasesIndex)

		screenings := participants.Group("/{pid}/screenings")
		screenings.Use(ScreeningPermissionRequired)
		screenings.GET("/create", SitePermissionRequired(ScreeningsCreateGet, sitePerm.Screening.Create)).Name("participantScreeningsCreatePath")
		screenings.POST("/create", SitePermissionRequired(ScreeningsCreatePost, sitePerm.Screening.Create)).Name("participantScreeningsCreatePath")
		screenings.GET("/edit/{sid}", SitePermissionRequired(ScreeningsEditGet, sitePerm.Screening.Edit)).Name("participantScreeningsEditPath")
		screenings.POST("/edit/{sid}", SitePermissionRequired(ScreeningsEditPost, sitePerm.Screening.Edit)).Name("participantScreeningsEditPath")
		screenings.DELETE("/delete/{sid}", SitePermissionRequired(AdminRequired(ScreeningsDestroy), sitePerm.Screening.Archive)).Name("participantScreeningsDeletePath")

		overReadings := screenings.Group("/{sid}/overreadings")
		overReadings.Middleware.Skip(ScreeningPermissionRequired, OverReadingsCreateGet, OverReadingsCreatePost, OverReadingsDetails, OverReadingsEditGet, OverReadingsEditPost)
		overReadings.Use(OverReadingPermissionRequired)
		overReadings.GET("/create", SitePermissionRequired(OverReadingsCreateGet, sitePerm.OverReading.Create)).Name("participantScreeningOverreadingsCreatePath")
		overReadings.POST("/create", SitePermissionRequired(OverReadingsCreatePost, sitePerm.OverReading.Create)).Name("participantScreeningOverreadingsCreatePath")
		overReadings.GET("/edit/{oid}", SitePermissionRequired(OverReadingsEditGet, sitePerm.OverReading.Edit)).Name("participantScreeningOverreadingsEditPath")
		overReadings.POST("/edit/{oid}", SitePermissionRequired(OverReadingsEditPost, sitePerm.OverReading.Edit)).Name("participantScreeningOverreadingsEditPath")
		overReadings.DELETE("/delete/{oid}", SitePermissionRequired(AdminRequired(OverReadingDestroy), sitePerm.OverReading.Archive)).Name("participantScreeningOverreadingsDeletePath")
		overReadings.Middleware.Skip(OverReadingPermissionRequired, OverReadingsDetails)
		overReadings.GET("/{oid}", OverReadingsDetails).Name("participantScreeningOverreadingPath")

		screenings.POST("/{sid}/appointmentdone", SitePermissionRequired(UpdateReferredMessage, sitePerm.Referrals.Create)).Name("participantsAppointmentPath")

		analytics := app.Group("/analytics")
		analytics.Use(LoginRequired)
		analytics.Use(StudyCoordinatorPermissionRequired)
		analytics.Middleware.Skip(StudyCoordinatorPermissionRequired, ReportsIndex)
		analytics.GET("/", ReportsIndex)
		analytics.GET("/index", ReportsIndex)
		analytics.POST("/api/list", ReportsIndexAPI)
		analytics.GET("/full-download", AdminRequired(DownloadFull))
		analytics.GET("/full-download-csv", AdminRequired(DownloadFullCSV))

		referrals := app.Group("/referrals")
		referrals.Use(LoginRequired)
		referrals.Use(ReferralTrackerPermissionRequired)
		referrals.Middleware.Skip(ReferralTrackerPermissionRequired, ReferralsIndex, ReferralsParticipantsView)
		referrals.GET("/", ReferralsIndex)
		referrals.GET("/index", ReferralsIndex)
		referrals.GET("/participants/{pid}", SitePermissionRequired(ReferralsParticipantsGet, sitePerm.Referrals.Create)).Name("referralsParticipantPath")
		referrals.GET("/participants/{pid}/view", ReferralsParticipantsView).Name("referralsParticipantViewPath")

		referrals.DELETE("/participants/{pid}/delete/{rid}", SitePermissionRequired(AdminRequired(ReferralsDestroy), sitePerm.Referrals.Archive)).Name("referralsParticipantDeletePath")

		notifications := app.Group("/notifications")
		notifications.Use(LoginRequired)
		notifications.GET("/", NotificationsIndex)
		notifications.GET("/index", NotificationsIndex)

		notifications.DELETE("/delete/{nid}", SitePermissionRequired(AdminRequired(NotificationsDestroy), sitePerm.Screening.Archive)).Name("notificationsDeletePath")

		logs := app.Group("/logs")
		logs.Use(AdminRequired)
		logs.GET("/", SystemLogsIndex)
		logs.GET("/index", SystemLogsIndex)

		archive := app.Group("/archives")
		archive.Use(AdminRequired)
		archive.GET("/", ArchiveIndex)
		archive.GET("/index", ArchiveIndex)
		archive.GET("/restore/{aid}", SitePermissionRequired(ArchiveRestore, sitePerm.OverReading.Restore || sitePerm.Screening.Restore || sitePerm.Participant.Restore || sitePerm.Referrals.Restore)).Name("archivesRestorePath")
		archive.DELETE("/delete/{aid}", SitePermissionRequired(ArchiveDestroy, sitePerm.OverReading.Delete || sitePerm.Screening.Delete || sitePerm.Participant.Delete || sitePerm.Referrals.Delete)).Name("archivesDeletePath")
		archive.GET("/{aid}", ArchiveShow).Name("archivePath")

		app.GET("/errors/{status}", ErrorsDefault)

		app.GET("/switch", ChangeLanguage)
		app.GET("/switch-site", ChangeSite)

		app.POST("/notifications", SitePermissionRequired(ScreeningPermissionRequired(ChangeNotificationStatus), sitePerm.Screening.Create || sitePerm.Screening.Edit))

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
	if T, err = i18n.New(packr.New("../locales", "../locales"), "en-US"); err != nil {
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
	return forcessl.Middleware(secure.Options{
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
		log.Println("Error handler:", err, c.Request().RequestURI)
		ct := c.Request().Header.Get("Content-Type")
		if c.Value("current_user") != nil {
			user, ok := c.Value("current_user").(*models.User)
			if ok {
				InsertLog("error", "Error", err.Error(), "", "", user.ID, c)
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
