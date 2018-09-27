package grifts

import (
	"github.com/gobuffalo/buffalo"
	"github.com/monarko/piia/actions"
)

func init() {
	buffalo.Grifts(actions.App())
}
