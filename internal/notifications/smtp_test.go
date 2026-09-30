package notifications

import (
	"bufio"
	"context"
	"errors"
	"net"
	"net/textproto"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

type smtpFixture struct {
	mu         sync.Mutex
	messages   []string
	recipients []string
	reject     bool
	listener   net.Listener
	wg         sync.WaitGroup
}

func smtpServer(t *testing.T) (*smtpFixture, Config) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	f := &smtpFixture{listener: ln}
	f.wg.Add(1)
	go func() {
		defer f.wg.Done()
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			f.wg.Add(1)
			go func() {
				defer f.wg.Done()
				defer c.Close()
				_ = c.SetDeadline(time.Now().Add(5 * time.Second))
				r := textproto.NewReader(bufio.NewReader(c))
				w := bufio.NewWriter(c)
				write := func(s string) { _, _ = w.WriteString(s + "\r\n"); _ = w.Flush() }
				write("220 local fixture")
				for {
					line, err := r.ReadLine()
					if err != nil {
						return
					}
					switch {
					case strings.HasPrefix(line, "EHLO"):
						write("250 local fixture")
					case strings.HasPrefix(line, "MAIL FROM:"):
						write("250 sender accepted")
					case strings.HasPrefix(line, "RCPT TO:"):
						f.mu.Lock()
						f.recipients = append(f.recipients, line)
						f.mu.Unlock()
						write("250 recipient accepted")
					case line == "DATA":
						write("354 data")
						b, err := r.ReadDotBytes()
						if err != nil {
							return
						}
						f.mu.Lock()
						reject := f.reject
						if !reject {
							f.messages = append(f.messages, string(b))
						}
						f.mu.Unlock()
						if reject {
							write("451 fixture private response do-not-persist")
						} else {
							write("250 accepted")
						}
					case line == "QUIT":
						write("221 bye")
						return
					default:
						write("500 unsupported")
					}
				}
			}()
		}
	}()
	t.Cleanup(func() { _ = ln.Close(); f.wg.Wait() })
	host, port, _ := net.SplitHostPort(ln.Addr().String())
	n, _ := strconv.Atoi(port)
	return f, Config{Enabled: true, Host: host, Port: n, TLS: "loopback", From: "iwa@example.test", Recipients: []string{"operator@example.test", "second@example.test"}, ReminderSeconds: 3600, PollSeconds: 1, TimeoutSeconds: 2}
}
func (f *smtpFixture) count() int { f.mu.Lock(); defer f.mu.Unlock(); return len(f.messages) }

func TestSMTPTransportAndNoDowngrade(t *testing.T) {
	f, c := smtpServer(t)
	sender, err := NewSMTP(c)
	if err != nil {
		t.Fatal(err)
	}
	m := Message{ID: "stable@iwa.invalid", Category: "source_delay", Kind: "opened", IncidentID: 1, Scope: "fixture", CreatedAt: time.Now(), OpenedAt: time.Now()}
	if err = sender.Send(context.Background(), m); err != nil {
		t.Fatal(err)
	}
	f.mu.Lock()
	body := f.messages[0]
	recipients := len(f.recipients)
	f.mu.Unlock()
	if f.count() != 1 || recipients != 2 || !strings.Contains(body, "Message-ID: <stable@iwa.invalid>") || !strings.Contains(body, "source_delay") || strings.Contains(body, "operator@example.test") {
		t.Fatal("SMTP envelope or content mismatch")
	}
	c.TLS = "starttls"
	c.Username = "private-user"
	c.Password = "private-password"
	sender, _ = NewSMTP(c)
	if err = sender.Send(context.Background(), m); !errors.Is(err, ErrDelivery) || strings.Contains(err.Error(), c.Password) {
		t.Fatal("STARTTLS downgrade or error leak")
	}
	c.TLS = "implicit"
	sender, _ = NewSMTP(c)
	if err = sender.Send(context.Background(), m); !errors.Is(err, ErrDelivery) {
		t.Fatal("implicit TLS accepted plaintext")
	}
	if f.count() != 1 {
		t.Fatal("TLS failure sent plaintext message")
	}
}

func TestPrivateConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "smtp.json")
	write := func(body string) {
		t.Helper()
		if err := os.WriteFile(path, []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
	}
	valid := `{"enabled":true,"host":"smtp.example.test","from":"iwa@example.test","recipients":["operator@example.test"],"username":"secret-user","password":"secret-password"}`
	write(valid)
	c, err := LoadConfig(path)
	if err != nil || c.ReminderSeconds != 21600 || c.TLS != "starttls" {
		t.Fatalf("defaults: %v", err)
	}
	for _, body := range []string{
		strings.Replace(valid, `"host":"smtp.example.test"`, `"host":"smtp.example.test","tls":"loopback"`, 1),
		strings.Replace(valid, `"from":"iwa@example.test"`, `"from":"iwa@example.test\r\nBcc: theft@example.test"`, 1),
		strings.Replace(valid, `"recipients":["operator@example.test"]`, `"recipients":[]`, 1),
		strings.Replace(valid, `"enabled":true`, `"enabled":true,"reminder_seconds":-1`, 1),
		strings.Replace(valid, `"enabled":true`, `"enabled":true,"unknown_secret":"secret"`, 1), valid + ` {}`, `{"enabled":true,`,
	} {
		write(body)
		_, err = LoadConfig(path)
		if !errors.Is(err, ErrConfig) || strings.Contains(err.Error(), "secret") {
			t.Fatal("invalid private config or leak")
		}
	}
	write(`{"enabled":false}`)
	if c, err = LoadConfig(path); err != nil || c.Enabled {
		t.Fatal("disabled configuration")
	}
	if c, err = LoadConfig(""); err != nil || c.Enabled {
		t.Fatal("default must not send email")
	}
	if _, err = LoadConfig(path + "missing"); !errors.Is(err, ErrConfig) {
		t.Fatal("missing explicit file must fail")
	}
}

func TestReportTransport(t *testing.T) {
	f, c := smtpServer(t)
	s, _ := NewSMTP(c)
	at := time.Now()
	if err := s.SendReport(context.Background(), "daily@iwa.invalid", "[IWA] Report allerte", "Toscana\nStato non disponibile", at); err != nil {
		t.Fatal(err)
	}
	f.mu.Lock()
	body := f.messages[0]
	f.mu.Unlock()
	if !strings.Contains(body, "Message-ID: <daily@iwa.invalid>") || !strings.Contains(body, "Stato non disponibile") || strings.Contains(body, "operator@example.test") {
		t.Fatal("report transport")
	}
	for _, item := range []struct{ id, subject string }{{"daily@iwa.invalid", "hello\r\nBcc: attacker"}, {"bad\n@iwa.invalid", "Report"}} {
		if s.SendReport(context.Background(), item.id, item.subject, "body", at) != ErrDelivery {
			t.Fatal("header injection")
		}
	}
	c.TLS = "starttls"
	s, _ = NewSMTP(c)
	if s.SendReport(context.Background(), "daily@iwa.invalid", "Report", "body", at) != ErrDelivery {
		t.Fatal("report TLS downgrade")
	}
	if f.count() != 1 {
		t.Fatal("invalid report delivered")
	}
}
