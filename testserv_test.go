package testserv

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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
		server, client := startServer(
			Instruction{HeaderDelay: 5 * time.Millisecond, StatusCode: 200, Body: []byte("hello")},
			Instruction{
				StatusCode: 400,
				Header: http.Header{
					"Foo":            []string{"Bar"},
					"Content-Length": []string{"1111"},
				},
				BodyServiceTime: 5 * time.Millisecond,
				Body:            []byte("baz"),
			},
			Instruction{StatusCode: 500, Body: []byte{}},
		)
		defer server.Close()

		start1 := time.Now()
		resp1, err1, body1 := get(server, client)
		duration1 := time.Now().Sub(start1)

		require.NoError(t, err1)
		require.NotNil(t, resp1)
		assert.Equal(t, 200, resp1.StatusCode)
		assert.Equal(t, "hello", body1)
		assert.GreaterOrEqual(t, duration1, 5*time.Millisecond)

		start2 := time.Now()
		resp2, err2, _ := get(server, client)
		duration2 := time.Now().Sub(start2)

		assert.EqualError(t, err2, "unexpected EOF")
		require.NotNil(t, resp2)
		assert.Equal(t, 400, resp2.StatusCode)
		assert.Equal(t, "Bar", resp2.Header.Get("Foo"))
		assert.Equal(t, "1111", resp2.Header.Get("Content-Length"))
		assert.GreaterOrEqual(t, duration2, 5*time.Millisecond)

		resp3, err3, body3 := get(server, client)

		require.NoError(t, err3)
		require.NotNil(t, resp3)
		assert.Equal(t, 500, resp3.StatusCode)
		assert.Equal(t, "0", resp3.Header.Get("Content-Length"))
		assert.Equal(t, "", body3)
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
