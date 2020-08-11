package actions

import (
	"strings"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/pop"
	"github.com/monarko/piia/helpers"
	"github.com/monarko/piia/models"
)

// SetCurrentUser attempts to find a user based on the current_user_id
// in the session. If one is found it is set on the context.
func SetCurrentUser(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		if uid := c.Session().Get("current_user_id"); uid != nil {
			u := &models.User{}
			tx := c.Value("tx").(*pop.Connection)
			err := tx.Find(u, uid)
			if err == nil {
				u.Sites = strings.Split(u.Site, "")
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

// StudyTeamMemberPermissionRequired requires a user to be logged in and to be an admin before accessing a route.
func StudyTeamMemberPermissionRequired(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		user, ok := c.Value("current_user").(*models.User)
		if ok {
			if user.Admin || user.Permission.StudyTeamMember {
				return next(c)
			}
		}
		c.Flash().Add("danger", "You are not authorized to view that page.")
		return c.Redirect(302, "/")
	}
}

// SetCurrentLang attempts to find a user based on the current_user_id
// in the session. If one is found it is set on the context.
func SetCurrentLang(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		if lang := c.Session().Get("lang"); lang != nil {
			c.Set("current_lang", lang)
		} else {
			c.Set("current_lang", "en")
		}
		return next(c)
	}
}

// SetCurrentSite attempts to find a user based on the current_user_id
// in the session. If one is found it is set on the context.
func SetCurrentSite(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		if site := c.Session().Get("site"); site != nil {
			c.Set("current_site", site)
		} else {
			user, ok := c.Value("current_user").(*models.User)
			if ok {
				if user.Admin || user.Permission.StudyCoordinator {
					c.Set("current_site", "")
				} else {
					c.Set("current_site", user.Site)
					if len(user.Site) > 1 {
						c.Set("current_site", user.Sites[0])
					}
				}
			} else {
				c.Set("current_site", "")
			}
		}

		return next(c)
	}
}

// SetBreadcrumb sets a systemwide breadcrumb object
func SetBreadcrumb(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		if breadcrumb := c.Session().Get("breadcrumb"); breadcrumb != nil {
			c.Set("breadcrumb", breadcrumb)
		} else {
			bread := make(helpers.Breadcrumbs, 0)
			home := helpers.Breadcrumb{Title: "nav_home", Path: "/"}
			bread = append(bread, home)
			c.Set("breadcrumb", bread)
		}
		return next(c)
	}
}

// Perm object
type Perm struct {
	Create  bool
	Edit    bool
	Archive bool
	Delete  bool
	Restore bool
}

// SitePermission object
type SitePermission struct {
	User        Perm
	Participant Perm
	Screening   Perm
	OverReading Perm
	Referrals   Perm
}

func defaultPermissions() map[string]SitePermission {
	planning := SitePermission{
		User: Perm{
			Create:  true,
			Edit:    true,
			Archive: true,
			Delete:  true,
			Restore: true,
		},
		Participant: Perm{
			Create:  false,
			Edit:    false,
			Archive: false,
			Delete:  false,
			Restore: false,
		},
		Screening: Perm{
			Create:  false,
			Edit:    false,
			Archive: false,
			Delete:  false,
			Restore: false,
		},
		OverReading: Perm{
			Create:  false,
			Edit:    false,
			Archive: false,
			Delete:  false,
			Restore: false,
		},
		Referrals: Perm{
			Create:  false,
			Edit:    false,
			Archive: false,
			Delete:  false,
			Restore: false,
		},
	}
	recruiting := SitePermission{
		User: Perm{
			Create:  true,
			Edit:    true,
			Archive: true,
			Delete:  true,
			Restore: true,
		},
		Participant: Perm{
			Create:  true,
			Edit:    true,
			Archive: true,
			Delete:  true,
			Restore: true,
		},
		Screening: Perm{
			Create:  true,
			Edit:    true,
			Archive: true,
			Delete:  true,
			Restore: true,
		},
		OverReading: Perm{
			Create:  true,
			Edit:    true,
			Archive: true,
			Delete:  true,
			Restore: true,
		},
		Referrals: Perm{
			Create:  true,
			Edit:    true,
			Archive: true,
			Delete:  true,
			Restore: true,
		},
	}
	closed := SitePermission{
		User: Perm{
			Create:  true,
			Edit:    true,
			Archive: true,
			Delete:  true,
			Restore: true,
		},
		Participant: Perm{
			Create:  false,
			Edit:    true,
			Archive: true,
			Delete:  false,
			Restore: true,
		},
		Screening: Perm{
			Create:  false,
			Edit:    true,
			Archive: true,
			Delete:  false,
			Restore: true,
		},
		OverReading: Perm{
			Create:  true,
			Edit:    true,
			Archive: true,
			Delete:  true,
			Restore: true,
		},
		Referrals: Perm{
			Create:  true,
			Edit:    true,
			Archive: true,
			Delete:  true,
			Restore: true,
		},
	}
	locked := SitePermission{
		User: Perm{
			Create:  true,
			Edit:    true,
			Archive: true,
			Delete:  true,
			Restore: true,
		},
		Participant: Perm{
			Create:  false,
			Edit:    false,
			Archive: false,
			Delete:  false,
			Restore: false,
		},
		Screening: Perm{
			Create:  false,
			Edit:    false,
			Archive: false,
			Delete:  false,
			Restore: false,
		},
		OverReading: Perm{
			Create:  false,
			Edit:    false,
			Archive: false,
			Delete:  false,
			Restore: false,
		},
		Referrals: Perm{
			Create:  false,
			Edit:    false,
			Archive: false,
			Delete:  false,
			Restore: false,
		},
	}
	completed := SitePermission{
		User: Perm{
			Create:  false,
			Edit:    false,
			Archive: false,
			Delete:  false,
			Restore: false,
		},
		Participant: Perm{
			Create:  false,
			Edit:    false,
			Archive: false,
			Delete:  false,
			Restore: false,
		},
		Screening: Perm{
			Create:  false,
			Edit:    false,
			Archive: false,
			Delete:  false,
			Restore: false,
		},
		OverReading: Perm{
			Create:  false,
			Edit:    false,
			Archive: false,
			Delete:  false,
			Restore: false,
		},
		Referrals: Perm{
			Create:  false,
			Edit:    false,
			Archive: false,
			Delete:  false,
			Restore: false,
		},
	}
	perms := map[string]SitePermission{
		"planning":   planning,
		"recruiting": recruiting,
		"closed":     closed,
		"locked":     locked,
		"completed":  completed,
	}

	return perms
}

// SiteStatus returns the site details
func SiteStatus() (SitePermission, string, map[string]string) {
	desc := map[string]map[string]string{
		"planning": {
			"desc": "Site is currently in planning state",
			"type": "btn-outline-success",
		},
		"recruiting": {
			"desc": "Site is actively recruiting people to participate",
			"type": "btn-outline-primary",
		},
		"closed": {
			"desc": "Site is currently in closed state",
			"type": "btn-outline-info",
		},
		"locked": {
			"desc": "Site is currently in locked state",
			"type": "btn-outline-warning",
		},
		"completed": {
			"desc": "Study is completed",
			"type": "btn-outline-danger",
		},
	}
	currentStatus := strings.TrimSpace(envy.Get("SITE_STATUS", "recruiting"))
	perms := defaultPermissions()

	return perms[currentStatus], currentStatus, desc[currentStatus]
}

// SetSiteStatus sets the permission for specific site
func SetSiteStatus(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		perm, status, desc := SiteStatus()
		c.Set("sitePerm", perm)
		c.Set("siteStatus", status)
		c.Set("siteStatusDesc", desc)
		return next(c)
	}
}

// SitePermissionRequired requires a user to be logged in and to be an admin before accessing a route.
func SitePermissionRequired(next buffalo.Handler, perm bool) buffalo.Handler {
	return func(c buffalo.Context) error {
		if perm {
			return next(c)
		}
		p := c.Request().Referer()
		c.Flash().Add("danger", "Currently no support for this action.")
		return c.Redirect(302, p)
	}
}
