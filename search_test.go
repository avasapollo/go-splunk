package splunk

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetURL(t *testing.T) {
	s := Search{
		SearchID: "TestID",
		client: &Client{
			config: &Config{
				BaseURL: "http://localhost:8090",
			},
		},
	}

	if url := s.URL(); url != "http://localhost/en-US/app/search/search?sid=TestID" {
		t.Errorf("Bad URL: %s", url)
	}
	if url := s.URL("http://localhost:81"); url != "http://localhost:81/en-US/app/search/search?sid=TestID" {
		t.Errorf("Bad URL: %s", url)
	}
}

func TestClient_DeleteSearchJob(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			switch {
			case req.Method != http.MethodDelete:
				rw.WriteHeader(http.StatusMethodNotAllowed)
				return
			case req.RequestURI != "/services/search/jobs/job_id_1?output_mode=json":
				rw.WriteHeader(http.StatusBadRequest)
				return
			}
			rw.WriteHeader(http.StatusOK)
		}))
		defer server.Close()
		client := &Client{
			config: &Config{
				BaseURL:    server.URL,
				HTTPClient: http.DefaultClient,
			},
		}
		err := client.DeleteSearchJob(context.Background(), "job_id_1")
		require.NoError(t, err)
	})
}