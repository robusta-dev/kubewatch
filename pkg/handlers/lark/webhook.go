/*
Copyright 2018 Bitnami

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

package lark

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"os"

	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/bitnami-labs/kubewatch/config"
	"github.com/bitnami-labs/kubewatch/pkg/event"
)

var webhookErrMsg = `
%s

You need to set Lark Webhook url
using "--url/-u", or using environment variables:

export KW_LARK_WEBHOOK_URL=webhook_url

Command line flags will override environment variables

`

// Webhook handler implements handler.Handler interface,
// Notify event to Webhook channel
type Webhook struct {
	Url string
}

// TextMessage for messages
type TextMessage struct {
	MsgType string       `json:"msg_type"`
	Content *TextContent `json:"content"`
}

type TextContent struct {
	Text string `json:"text"`
}

// Init prepares Webhook configuration
func (m *Webhook) Init(c *config.Config) error {
	url := c.Handler.Lark.WebhookURL
	if url == "" {
		url = os.Getenv("KW_LARK_WEBHOOK_URL")
	}
	m.Url = url
	return checkMissingWebhookVars(m)
}

// Handle handles an event.
func (m *Webhook) Handle(e event.Event) {
	webhookMessage := prepareWebhookMessage(e, m)

	err := postMessage(m.Url, webhookMessage)
	if err != nil {
		logrus.Printf("%s\n", err)
		return
	}
	logrus.Printf("Message successfully sent to lark webhook: %s at %s ", m.Url, time.Now())
}

func checkMissingWebhookVars(s *Webhook) error {
	if s.Url == "" {
		return fmt.Errorf(webhookErrMsg, "Missing Webhook url")
	}
	return nil
}

func prepareWebhookMessage(e event.Event, m *Webhook) *TextMessage {
	return &TextMessage{
		MsgType: "text",
		Content: &TextContent{Text: e.Message()},
	}
}

func postMessage(url string, textMessage *TextMessage) error {
	message, err := json.Marshal(textMessage)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(message))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	_, err = client.Do(req)
	if err != nil {
		return err
	}
	return nil
}
