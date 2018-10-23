package mailers

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"
	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/packr"
	"github.com/mailgun/mailgun-go"
	"github.com/monarko/piia/helpers"
	"github.com/pkg/errors"
)

var r *render.Engine

var (
	mailgunDomain = envy.Get("MAILGUN_DOMAIN", "mg.monarko.com")
	privateAPIKey = envy.Get("MAILGUN_API_KEY", "")

	emailName    = envy.Get("EMAIL_NAME", "")
	emailAddress = envy.Get("EMAIL_SENDER", "no-reply@example.com")
	mg           *mailgun.MailgunImpl

	mailgunURL string
)

func init() {
	mailgunURL = fmt.Sprintf("https://api:%s@api.mailgun.net/v3/%s/messages", privateAPIKey, mailgunDomain)

	mg = mailgun.NewMailgun(mailgunDomain, privateAPIKey)

	// // Pulling config from the env.
	// port := envy.Get("SMTP_PORT", "1025")
	// host := envy.Get("SMTP_HOST", "localhost")
	// user := envy.Get("SMTP_USER", "")
	// password := envy.Get("SMTP_PASSWORD", "")

	// var err error
	// smtp, err = mail.NewSMTPSender(host, port, user, password)

	// if err != nil {
	// 	log.Fatal(err)
	// }

	r = render.New(render.Options{
		HTMLLayout:   "layout.html",
		TemplatesBox: packr.NewBox("../templates/mail"),
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

	c.Set("data", em.Data)
	rend := r.HTML("welcome_email.html")
	var buf bytes.Buffer
	err := rend.Render(&buf, em.Data)
	if err != nil {
		return errors.WithStack(err)
	}
	m := mg.NewMessage(from, em.Subject, buf.String(), em.To...)
	resp, id, err := mg.Send(m)

	if err != nil {
		return errors.WithStack(err)
	}

	return errors.Errorf("ID: %s Resp: %s\n", id, resp)

	// return nil
}

// Send an email
func (em EmailDetails) Send(c buffalo.Context) error {
	code := helpers.RandomStringWithLength(15)
	codeString := "WebKitFormBoundary" + code

	from := fmt.Sprintf("Content-Disposition: form-data; name=\"from\"\r\n\r\n\"%s\" <%s>\r\n", emailName, emailAddress)
	to := fmt.Sprintf("Content-Disposition: form-data; name=\"to\"\r\n\r\n%s\r\n", strings.Join(em.To, ","))
	subject := fmt.Sprintf("Content-Disposition: form-data; name=\"subject\"\r\n\r\n%s\r\n", em.Subject)

	c.Set("data", em.Data)
	rend := r.HTML("welcome_email.html")
	var buf bytes.Buffer
	err := rend.Render(&buf, em.Data)
	if err != nil {
		return errors.WithStack(err)
	}

	html := fmt.Sprintf("Content-Disposition: form-data; name=\"html\"\r\n\r\n%s", buf.String())

	payloadString := fmt.Sprintf(
		"------%s\r\n%s------%s\r\n%s------%s\r\n%s------%s\r\n%s\r\n------%s--",
		codeString,
		from,
		codeString,
		to,
		codeString,
		subject,
		codeString,
		html,
		codeString,
	)
	payload := strings.NewReader(payloadString)

	req, err := http.NewRequest("POST", mailgunURL, payload)
	if err != nil {
		// log.Fatal(err)
		return errors.WithStack(err)
	}

	contentType := fmt.Sprintf("multipart/form-data; boundary=----%s", codeString)
	req.Header.Add("content-type", contentType)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("cache-control", "no-cache")

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	res, err := client.Do(req)
	if err != nil {
		return errors.WithStack(err)
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return errors.WithStack(err)
	}

	log.Println(string(body))

	return nil
}
