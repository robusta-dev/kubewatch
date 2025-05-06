package webex

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/bitnami-labs/kubewatch/config"
)

func TestHipchatInit(t *testing.T) {
	s := &Webex{}
	expectedError := fmt.Errorf(webexErrMsg, "Missing webex token or room")

	var Tests = []struct {
		webex config.Webex
		err   error
	}{
		{config.Webex{Token: "foo", Room: "bar"}, nil},
		{config.Webex{Token: "foo"}, expectedError},
		{config.Webex{Room: "bar"}, expectedError},
		{config.Webex{}, expectedError},
	}

	for _, tt := range Tests {
		c := &config.Config{}
		c.Handler.Webex = tt.webex
		if err := s.Init(c); !reflect.DeepEqual(err, tt.err) {
			t.Fatalf("Init(): %v", err)
		}
	}
}
