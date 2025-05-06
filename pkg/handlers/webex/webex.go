package webex

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"os"

	webex "github.com/jbogarin/go-cisco-webex-teams/sdk"

	"github.com/bitnami-labs/kubewatch/config"
	"github.com/bitnami-labs/kubewatch/pkg/event"
)

var webexErrMsg = `
%s

You need to set both webex token and room for webex notify,
using "--token/-t", "--room/-r", and "--url/-u" or using environment variables:

export WEBEX_ACCESS_TOKEN=webex_token
export WEBEX_ACCESS_ROOM=webex_room
export WEBEX_ACCESS_URL=webex_url (defaults to https://webexapis.com/v1/messages)

Command line flags will override environment variables

`

// Webex handler implements handler.Handler interface,
// Notify event to Webex room
type Webex struct {
	Token string
	Room  string
	Url   string
}

// Init prepares Webex configuration
func (s *Webex) Init(c *config.Config) error {
	url := c.Handler.Webex.Url
	room := c.Handler.Webex.Room
	token := c.Handler.Webex.Token

	if token == "" {
		token = os.Getenv("WEBEX_ACCESS_TOKEN")
	}

	if room == "" {
		room = os.Getenv("WEBEX_ROOM")
	}

	if url == "" {
		url = os.Getenv("WEBEX_URL")
	}

	s.Token = token
	s.Room = room
	s.Url = url

	return checkMissingWebexVars(s)
}

// Handle handles the notification.
func (s *Webex) Handle(e event.Event) {
	client := webex.NewClient()
	client.SetAuthToken(s.Token)
	message := &webex.MessageCreateRequest{
		RoomID: s.Room,
		Text:   e.Message(),
	}

	_, response, err := client.Messages.CreateMessage(message)
	if err != nil {
		fmt.Println("Error sending message:", err)
		return
	}
	logrus.Printf("Message sent: Return Code %d", response.StatusCode())
	logrus.Printf("Message successfully sent to room %s", s.Room)
}

func checkMissingWebexVars(s *Webex) error {
	if s.Token == "" || s.Room == "" {
		return fmt.Errorf(webexErrMsg, "Missing webex token or room")
	}

	return nil
}
