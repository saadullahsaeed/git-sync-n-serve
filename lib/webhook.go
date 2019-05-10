package syncserve

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	defaultWebhookTimeout = time.Second * 5
	defaultContentType    = "application/json"
)

// Event ...
type Event struct {
	Repository string
	Branch     string
	Action     string
}

// String ...
func (e *Event) String() string {
	return fmt.Sprintf("%s %s", e.Action, e.Repository)
}

// Webhook ...
type Webhook struct {
	URL             string
	Method          string
	PayloadTemplate string
	ContentType     string
	Logger          *log.Entry
}

// Send ...
func (w *Webhook) Send(e *Event) error {
	tmpl, err := template.New("webhook").Parse(w.PayloadTemplate)
	if err != nil {
		w.Logger.Error(err)
		return err
	}

	var buff bytes.Buffer
	err = tmpl.Execute(&buff, e)
	if err != nil {
		w.Logger.Error(err)
		return err
	}

	req, err := http.NewRequest(w.Method, w.URL, &buff)
	if err != nil {
		w.Logger.Error(err)
		return err
	}

	client := &http.Client{
		Timeout: defaultWebhookTimeout,
	}
	_, err = client.Do(req)
	if err != nil {
		w.Logger.Error(err)
	}
	return err
}

// Start the loop to send notifications
func (w *Webhook) Start(ch chan Event) {
	for {
		event := <-ch
		w.Logger.WithField("action", event.Action).Info("event received")

		if err := w.Send(&event); err != nil {
			w.Logger.Error(err)
		}
	}
}
