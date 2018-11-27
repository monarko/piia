package mailers

import (
	"bytes"
	"fmt"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"
	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/packr"
	mailgun "github.com/mailgun/mailgun-go"
	"github.com/pkg/errors"
)

var r *render.Engine

var (
	mailgunDomain = envy.Get("MAILGUN_DOMAIN", "mg.monarko.com")
	privateAPIKey = envy.Get("MAILGUN_API_KEY", "")

	emailName    = envy.Get("EMAIL_NAME", "")
	emailAddress = envy.Get("EMAIL_SENDER", "no-reply@example.com")
	mg           *mailgun.MailgunImpl
	mailgunURL   string
)

func init() {
	mailgunURL = fmt.Sprintf("https://api:%s@api.mailgun.net/v3/%s/messages", privateAPIKey, mailgunDomain)

	mg = mailgun.NewMailgun(mailgunDomain, privateAPIKey)

	r = render.New(render.Options{
		HTMLLayout:   "layout.html",
		TemplatesBox: packr.NewBox("../templates/mail"),
		AssetsBox:    packr.NewBox("../public"),
		Helpers:      render.Helpers{},
	})
}

// EmailDetails model
type EmailDetails struct {
	To      []string
	Subject string
	Data    map[string]interface{}
}

// SendMessage method
func (em EmailDetails) SendMessage(c buffalo.Context) error {
	from := envy.Get("EMAIL_SENDER", "no-reply@example.com")

	link := em.Data["link"].(string)
	body := "An admin created an account for you. Click " + link + " to activate your account."

	html := ""
	buf := &bytes.Buffer{}
	rnd := r.HTML("welcome_email.html")
	err := rnd.Render(buf, em.Data)
	if err == nil {
		html = buf.String()
	}

	sendMessage(mg, from, em.Subject, body, html, em.To)

	return nil
}

func sendMessage(mg mailgun.Mailgun, sender, subject, body, html string, recipient []string) error {
	message := mg.NewMessage(sender, subject, body, recipient...)
	message.SetHtml(html)
	_, _, err := mg.Send(message)

	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func constructMessage(data map[string]interface{}, c buffalo.Context) string {
	body := ""
	c.Set("data", data)
	rend := r.HTML("welcome_email.html")
	var buf bytes.Buffer
	err := rend.Render(&buf, data)
	if err != nil {
		body = err.Error()
	} else {
		body = buf.String()
	}

	return body
}
