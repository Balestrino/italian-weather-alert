package notifications

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"mime"
	"net"
	"net/smtp"
	"strings"
	"time"
	"unicode/utf8"
)

var ErrDelivery = errors.New("SMTP delivery failed")

type SMTP struct{ config Config }

func NewSMTP(c Config) (*SMTP, error) {
	if !c.Enabled || c.Validate() != nil {
		return nil, ErrConfig
	}
	return &SMTP{config: c}, nil
}

// Send suppresses server responses, which can contain credentials or addresses.
// STARTTLS is mandatory in starttls mode; no downgrade or insecure certificates.
func (s *SMTP) Send(ctx context.Context, m Message) error {
	if err := s.send(ctx, m); err != nil {
		return ErrDelivery
	}
	return nil
}
func (s *SMTP) send(ctx context.Context, m Message) error {
	return s.sendMail(ctx, m.ID, fmt.Sprintf("[IWA] %s / %s / incident %d", m.Category, m.Kind, m.IncidentID), m.Body(), m.CreatedAt)
}

// SendReport shares the secure envelope with incident delivery. Arbitrary
// source content belongs only in the body, never in mail headers.
func (s *SMTP) SendReport(ctx context.Context, id, subject, body string, at time.Time) error {
	if at.IsZero() || id == "" || len(id) > 200 || strings.ContainsAny(id, "\r\n<> \x00") || !strings.Contains(id, "@") || subject == "" || len(subject) > 300 || strings.ContainsAny(subject, "\r\n\x00") || len(body) > 256*1024 || !utf8.ValidString(body) || strings.ContainsRune(body, 0) {
		return ErrDelivery
	}
	if s.sendMail(ctx, id, mime.QEncoding.Encode("utf-8", subject), body, at) != nil {
		return ErrDelivery
	}
	return nil
}

func (s *SMTP) sendMail(ctx context.Context, id, subject, body string, at time.Time) error {
	c := s.config
	ctx, cancel := context.WithTimeout(ctx, time.Duration(c.TimeoutSeconds)*time.Second)
	defer cancel()
	conn, err := (&net.Dialer{}).DialContext(ctx, "tcp", c.address())
	if err != nil {
		return err
	}
	defer conn.Close()
	stop := context.AfterFunc(ctx, func() { _ = conn.Close() })
	defer stop()
	deadline, _ := ctx.Deadline()
	if err = conn.SetDeadline(deadline); err != nil {
		return err
	}
	tlsConfig := &tls.Config{ServerName: c.Host, MinVersion: tls.VersionTLS12}
	var transport net.Conn = conn
	if c.TLS == "implicit" {
		tlsConn := tls.Client(conn, tlsConfig)
		if err = tlsConn.HandshakeContext(ctx); err != nil {
			return err
		}
		transport = tlsConn
	}
	client, err := smtp.NewClient(transport, c.Host)
	if err != nil {
		return err
	}
	defer client.Close()
	if c.TLS == "starttls" {
		if ok, _ := client.Extension("STARTTLS"); !ok {
			return ErrDelivery
		}
		if err = client.StartTLS(tlsConfig); err != nil {
			return err
		}
	}
	if c.Username != "" {
		if err = client.Auth(smtp.PlainAuth("", c.Username, c.Password, c.Host)); err != nil {
			return err
		}
	}
	if err = client.Mail(c.From); err != nil {
		return err
	}
	for _, recipient := range c.Recipients {
		if err = client.Rcpt(recipient); err != nil {
			return err
		}
	}
	w, err := client.Data()
	if err != nil {
		return err
	}
	// No recipient list or arbitrary upstream text appears in the headers.
	header := fmt.Sprintf("From: %s\r\nTo: undisclosed-recipients:;\r\nSubject: %s\r\nDate: %s\r\nMessage-ID: <%s>\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\nContent-Transfer-Encoding: 8bit\r\n\r\n", c.From, subject, at.UTC().Format(time.RFC1123Z), id)
	if _, err = w.Write([]byte(header + strings.ReplaceAll(strings.ReplaceAll(body, "\r\n", "\n"), "\n", "\r\n"))); err != nil {
		return err
	}
	if err = w.Close(); err != nil {
		return err
	}
	// DATA acceptance is the delivery boundary; a subsequent QUIT failure does
	// not mean that the server rejected the message.
	_ = client.Quit()
	return nil
}
