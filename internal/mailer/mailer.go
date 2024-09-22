package mailer

import (
	"bytes"
	"embed"
	"html/template"
	"time"

	"github.com/go-mail/mail/v2"
)

//go:embed "templates"
var templateFs embed.FS

type Mailer struct {
	dailer *mail.Dialer
	sender string
}

func New(host string, port int, username, password, sender string) Mailer {
	dialer := mail.NewDialer(host, port, username, password)
	dialer.Timeout = 10 * time.Second
	return Mailer{
		dailer: dialer,
		sender: sender,
	}
}

func (m Mailer) Send(recipient string, templateFile string, data interface{}) error {
	tmpl, err := template.New("email").ParseFS(templateFs, "templates/"+templateFile)
	if err != nil {
		return err
	}
	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return err
	}

	plainBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(plainBody, "plainBody", data)
	if err != nil {
		return err
	}

	htmlBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(htmlBody, "htmlBody", data)
	if err != nil {
		return err
	}

	msg := mail.NewMessage()
	msg.SetHeader("To", recipient)
	msg.SetHeader("From", m.sender)
	msg.SetHeader("Subject", subject.String())
	msg.SetHeader("text/plain", plainBody.String())
	msg.AddAlternative("text/html", htmlBody.String())
	for i := 0; i < 3; i++ {

		err = m.dailer.DialAndSend(msg)
		if nil == err {
			return nil
		}
		time.Sleep(300 * time.Millisecond)
	}
	return err
}
