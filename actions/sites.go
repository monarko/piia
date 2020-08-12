package actions

import (
	"github.com/gobuffalo/buffalo"
)

// Site object
type Site map[string]string

var (
	sites = Site{
		"A": "Sansai",
		"D": "Doi Saket",
		"J": "Jomthong",
		"K": "Khlong Luang",
		"L": "Lamlukka",
		"N": "Nongseau",
		"O": "Phrao",
		"R": "Rajavithi",
		"S": "San Patong",
		"T": "Thanyaburi",
	}
)

// GetSystemSites returns current system sites
func GetSystemSites() Site {
	return sites
}

// ChangeSite changes the site
func ChangeSite(c buffalo.Context) error {
	selectedSite := c.Param("site")
	site := ""

	if _, ok := sites[selectedSite]; ok {
		site = selectedSite
	}

	c.Session().Set("site", site)

	referrer := c.Request().Referer()

	return c.Redirect(302, referrer)
}
