/*
Copyright 2016 Skippbox, Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package config

import (
//"io/ioutil"
"os"
"testing"
)

var configStr = `
{
    "handler": {
        "slack": {
	  "channel": "slack_channel",
	  "token": "slack_token"
	},
        "webhook": {
            "url": "http://localhost:8080"
        }
    },
    "reason": ["created", "deleted", "updated"],
    "resource": {
    	"deployment": "false",
    	"replicationcontroller": "false",
    	"replicaset": "false",
    	"daemonset": "false",
    	"services": "false",
    	"pod": "false",
    	"secret": "true",
    	"configmap": "true",
        "ingress": "false",
    },
}
`

func TestCheckMissingResourceEnvvars_Webhook(t *testing.T) {
	expectedURL := "http://example.com/webhook"
	os.Setenv("KW_WEBHOOK_URL", expectedURL)
	defer os.Unsetenv("KW_WEBHOOK_URL")

	c := &Config{}
	c.CheckMissingResourceEnvvars()

	if c.Handler.Webhook.Url != expectedURL {
		t.Errorf("Expected Webhook URL to be %q, but got %q", expectedURL, c.Handler.Webhook.Url)
	}
}
