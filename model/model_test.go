package model_test

import (
	"testing"

	. "github.com/bborbe/assert"
	"github.com/bborbe/teamvault-utils/model"
)

func TestTeamvaultApiUrlString(t *testing.T) {
	apiUrl := model.TeamvaultApiUrl("foo")
	if err := AssertThat(apiUrl.String(), Is("foo")); err != nil {
		t.Fatal(err)
	}
}
func TestTeamvaultApiUrlKey(t *testing.T) {
	apiUrl := model.TeamvaultApiUrl("foo")
	if err := AssertThat(apiUrl.String(), Is("foo")); err != nil {
		t.Fatal(err)
	}

	var tests = []struct {
		name          string
		url           string
		expectedError bool
		expectedKey   model.TeamvaultKey
	}{
		{"empty", "", true, ""},
		{"slash", "/", true, ""},
		{"two slashes", "hello/my/world", false, "my"},
		{"valid url", "https://teamvault.example.com/api/secrets/key123/", false, "key123"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiUrl := model.TeamvaultApiUrl(tt.url)
			key, err := apiUrl.Key()
			if (err != nil) != tt.expectedError {
				t.Fatalf("expected error %v got %v", tt.expectedError, err)
			}
			if key != tt.expectedKey {
				t.Fatalf("expected %v got %v", tt.expectedKey.String(), key.String())
			}
		})
	}
}
