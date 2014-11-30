package controllers

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "../autogen"
	ez "github.com/medvednikov/ezweb"
)

var (
	ts *httptest.Server
)

func TestQuickStart(t *testing.T) {
	// Create a test server serving one of the controllers
	ts = httptest.NewServer(http.HandlerFunc(ez.GetHandler(&Home{})))
	defer ts.Close()

	// Test the welcome message
	response := GET(t, ts.URL)
	if response != "Hello, stranger! :)" {
		t.Fatal("Wrong output: ", response)
	}

	// Test one of the arguments
	response = GET(t, ts.URL+"?name=Bobby")
	if response != "Hello, Bobby! :)" {
		t.Fatal("Wrong output: ", response)
	}

}

func GET(t *testing.T, url string) string {
	res, err := http.Get(url)
	handle(t, err)

	body, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()

	return string(body)

}

func handle(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}
