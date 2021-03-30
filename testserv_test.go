package testserv

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler_ServeHTTP(t *testing.T) {
	t.Run("Out of Instructions", func(t *testing.T) {
		server, client := startServer()
		defer server.Close()

		resp, err, body := get(server, client)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, 400, resp.StatusCode)
		assert.Equal(t, "Out of instructions: N[0] >= len(Inst)[0]", body)
		assert.Equal(t, "41", resp.Header.Get("Content-Length"))
	})
	t.Run("Nil Body", func(t *testing.T) {
		server, client := startServer(Instruction{StatusCode: 200})
		defer server.Close()

		resp, err, body := get(server, client)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, 200, resp.StatusCode)
		assert.NotContains(t, resp.Header, "Content-Length")
		assert.Empty(t, body)
	})
	t.Run("Multiple Instructions", func(t *testing.T) {
		server, client := startServer()
		defer server.Close()

		resp, err, body := get(server, client)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, 400, resp.StatusCode)
		assert.Equal(t, "Out of instructions: N[0] >= len(Inst)[0]", body)
		assert.Equal(t, "41", resp.Header.Get("Content-Length"))
	})
}

func startServer(inst ...Instruction) (*httptest.Server, *http.Client) {
	server := httptest.NewServer(&Handler{Inst: inst})
	return server, server.Client()
}

func get(server *httptest.Server, client *http.Client) (*http.Response, error, string) {
	resp, err := client.Get(server.URL)
	if err != nil {
		return resp, err, ""
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp, err, ""
	}

	return resp, nil, string(body)
}
