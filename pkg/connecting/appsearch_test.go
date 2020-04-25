package connecting

import (
	"fmt"
	"testing"

	"github.com/gookit/color"
)

var (
	red   = color.FgRed.Render
	green = color.FgGreen.Render
)

func TestCreateAppsearchClient(t *testing.T) {
	endpoint := "http://localhost:3002"
	api := "/test/api/v1/"
	token := "private-pq75aoza89cab89ozgd46aib"
	client := CreateAppsearchClient(endpoint, token, api)

	if fmt.Sprintf("%T", client.Client) != "*http.Client" {
		t.Errorf("\n%s:\n\n%s\n\n%s:\n\n%s", green("[expected]"), "*http.Client", red("[actual]"), fmt.Sprintf("%T", client))
	}

	if client.Token != token {
		t.Errorf("\n%s:\n\n%s\n\n%s:\n\n%s", green("[expected]"), token, red("[actual]"), client.Token)
	}

	if client.API != api {
		t.Errorf("\n%s:\n\n%s\n\n%s:\n\n%s", green("[expected]"), api, red("[actual]"), client.API)
	}

	if client.Endpoint != endpoint {
		t.Errorf("\n%s:\n\n%s\n\n%s:\n\n%s", green("[expected]"), endpoint, red("[actual]"), client.Endpoint)
	}
}
