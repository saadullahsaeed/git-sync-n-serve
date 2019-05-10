package syncserve_test

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	syncserve "github.com/saadullahsaeed/git-sync-static-server/lib"
	log "github.com/sirupsen/logrus"
)

func TestWebhook_Send(t *testing.T) {
	logger := log.WithContext(context.Background())
	tests := []struct {
		method          string
		tpl             string
		expectedPayload string
		err             error
	}{
		{
			method:          "GET",
			tpl:             "{{ .String }}",
			expectedPayload: "updated test",
			err:             nil,
		},
		{
			method:          "POST",
			tpl:             "Message: {{ .String }}",
			expectedPayload: "Message: updated test",
			err:             nil,
		},
		{
			method:          "POST",
			tpl:             "{{ .String }",
			expectedPayload: "",
			err:             errors.New("template: webhook:1: unexpected \"}\" in operand"),
		},
		{
			method:          "POST",
			tpl:             "{{ .X }}",
			expectedPayload: "",
			err:             errors.New("template: webhook:1:3: executing \"webhook\" at <.X>: can't evaluate field X in type *syncserve.Event"),
		},
	}

	for _, tt := range tests {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			if r.Method != tt.method {
				t.Errorf("expected method %s but got %s", tt.method, r.Method)
				return
			}

			payload, err := ioutil.ReadAll(r.Body)
			if err != nil {
				t.Error(err)
				return
			}

			if string(payload) != tt.expectedPayload {
				t.FailNow()
			}
		}))

		e := &syncserve.Event{Action: "updated", Repository: "test"}
		w := &syncserve.Webhook{
			URL:             ts.URL,
			Method:          tt.method,
			PayloadTemplate: tt.tpl,
			Logger:          logger,
		}
		err := w.Send(e)
		if err != nil {
			if err.Error() != tt.err.Error() {
				t.FailNow()
			}
		}

		if err == nil && tt.err != nil {
			t.FailNow()
		}

		ts.Close()
	}
}
