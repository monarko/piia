package actions

import "github.com/gobuffalo/buffalo"

// ParticipantHandler is a default handler to serve up
// list of participants.
func ParticipantHandler(c buffalo.Context) error {
	// then you'll have
	// GET /participant - will show CREATE
	// POST /participant/1 - will save new participant
	return c.Render(200, r.HTML("participants.html"))
}

