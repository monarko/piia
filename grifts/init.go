package grifts

import (
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop"
	"github.com/monarko/piia/actions"
	"github.com/monarko/piia/models"
)

var tx *pop.Connection

func init() {
	buffalo.Grifts(actions.App())
	tx = models.DB
}
