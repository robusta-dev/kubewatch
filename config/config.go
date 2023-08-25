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

//go:generate bash -c "go install ../tools/yannotated && yannotated -o sample.go -format go -package config -type Config"

package config

import (
	"io"
	"os"
	"path/filepath"
	"runtime"

	"gopkg.in/yaml.v3"
)

var (
	// ConfigFileName stores file of config
	ConfigFileName = ".kubewatch.yaml"

	// ConfigSample is a sample configuration file.
	ConfigSample = yannotated
)

// Handler contains handler configuration
type Handler struct {
	Slack        Slack        `yaml:"slack"`
	SlackWebhook SlackWebhook `yaml:"slackwebhook"`
	Hipchat      Hipchat      `yaml:"hipchat"`
	Mattermost   Mattermost   `yaml:"mattermost"`
	Flock        Flock        `yaml:"flock"`
	Webhook      Webhook      `yaml:"webhook"`
	CloudEvent   CloudEvent   `yaml:"cloudevent"`
	MSTeams      MSTeams      `yaml:"msteams"`
	SMTP         SMTP         `yaml:"smtp"`
	Lark         Lark         `yaml:"lark"`
}

// Resource contains resource configuration
type Resource struct {
	Deployment            bool `yaml:"deployment"`
	ReplicationController bool `yaml:"rc"`
	ReplicaSet            bool `yaml:"rs"`
	DaemonSet             bool `yaml:"ds"`
	StatefulSet           bool `yaml:"statefulset"`
	Services              bool `yaml:"svc"`
	Pod                   bool `yaml:"po"`
	Job                   bool `yaml:"job"`
	Node                  bool `yaml:"node"`
	ClusterRole           bool `yaml:"clusterrole"`
	ClusterRoleBinding    bool `yaml:"clusterrolebinding"`
	ServiceAccount        bool `yaml:"sa"`
	PersistentVolume      bool `yaml:"pv"`
	Namespace             bool `yaml:"ns"`
	Secret                bool `yaml:"secret"`
	ConfigMap             bool `yaml:"configmap"`
	Ingress               bool `yaml:"ing"`
	HPA                   bool `yaml:"hpa"`
	Event                 bool `yaml:"event"`
	CoreEvent             bool `yaml:"coreevent"`
}

// Config struct contains kubewatch configuration
type Config struct {
	// Handlers know how to send notifications to specific services.
	Handler Handler `yaml:"handler"`

	//Reason   []string `yaml:"reason"`

	// Resources to watch.
	Resource Resource `yaml:"resource"`

	// For watching specific namespace, leave it empty for watching all.
	// this config is ignored when watching namespaces
	Namespace string `yaml:"namespace,omitempty"`

	// Message properties .
	Message Message `yaml:"message"`
	// Diff properties .
	Diff Diff `yaml:"diff"`
}

type Diff struct {
	Enabled    bool     `yaml:"enabled"`
	IgnorePath []string `yaml:"ignore"`
}

// Message contains message configuration.
type Message struct {
	// Message title.
	Title string `yaml:"title"`
}

// Slack contains slack configuration
type Slack struct {
	// Slack "legacy" API token.
	Token string `yaml:"token"`
	// Slack channel.
	Channel string `yaml:"channel"`
	// Title of the message.
	//Title string `yaml:"title"` // moved to Message
}

// SlackWebhook contains slack configuration
type SlackWebhook struct {
	// Slack channel.
	Channel string `yaml:"channel"`
	// Slack Username.
	Username string `yaml:"username"`
	// Slack Emoji.
	Emoji string `yaml:"emoji"`
	// Slack Webhook Url.
	Slackwebhookurl string `yaml:"slackwebhookurl"`
}

// Hipchat contains hipchat configuration
type Hipchat struct {
	// Hipchat token.
	Token string `yaml:"token"`
	// Room name.
	Room string `yaml:"room"`
	// URL of the hipchat server.
	Url string `yaml:"url"`
}

// Mattermost contains mattermost configuration
type Mattermost struct {
	Channel  string `yaml:"room"`
	Url      string `yaml:"url"`
	Username string `yaml:"username"`
}

// Flock contains flock configuration
type Flock struct {
	// URL of the flock API.
	Url string `yaml:"url"`
}

// Webhook contains webhook configuration
type Webhook struct {
	// Webhook URL.
	Url     string `yaml:"url"`
	Cert    string `yaml:"cert"`
	TlsSkip bool   `yaml:"tlsskip"`
}

// Lark contains lark configuration
type Lark struct {
	// Webhook URL.
	WebhookURL string `yaml:"webhookurl"`
}

// CloudEvent contains CloudEvent configuration
type CloudEvent struct {
	Url string `yaml:"url"`
}

// MSTeams contains MSTeams configuration
type MSTeams struct {
	// MSTeams API Webhook URL.
	WebhookURL string `yaml:"webhookurl"`
}

// SMTP contains SMTP configuration.
type SMTP struct {
	// Destination e-mail address.
	To string `yaml:"to" yaml:"to,omitempty"`
	// Sender e-mail address .
	From string `yaml:"from" yaml:"from,omitempty"`
	// Smarthost, aka "SMTP server"; address of server used to send email.
	Smarthost string `yaml:"smarthost" yaml:"smarthost,omitempty"`
	// Subject of the outgoing emails.
	Subject string `yaml:"subject" yaml:"subject,omitempty"`
	// Extra e-mail headers to be added to all outgoing messages.
	Headers map[string]string `yaml:"headers" yaml:"headers,omitempty"`
	// Authentication parameters.
	Auth SMTPAuth `yaml:"auth" yaml:"auth,omitempty"`
	// If "true" forces secure SMTP protocol (AKA StartTLS).
	RequireTLS bool `yaml:"requireTLS" yaml:"requireTLS"`
	// SMTP hello field (optional)
	Hello string `yaml:"hello" yaml:"hello,omitempty"`
}

type SMTPAuth struct {
	// Username for PLAN and LOGIN auth mechanisms.
	Username string `yaml:"username" yaml:"username,omitempty"`
	// Password for PLAIN and LOGIN auth mechanisms.
	Password string `yaml:"password" yaml:"password,omitempty"`
	// Identity for PLAIN auth mechanism
	Identity string `yaml:"identity" yaml:"identity,omitempty"`
	// Secret for CRAM-MD5 auth mechanism
	Secret string `yaml:"secret" yaml:"secret,omitempty"`
}

// New creates new config object
func New() (*Config, error) {
	c := &Config{}
	if err := c.Load(); err != nil {
		return c, err
	}

	return c, nil
}

func createIfNotExist() error {
	// create file if not exist
	configFile := filepath.Join(configDir(), ConfigFileName)
	_, err := os.Stat(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			file, err := os.Create(configFile)
			if err != nil {
				return err
			}
			file.Close()
		} else {
			return err
		}
	}
	return nil
}

// Load loads configuration from config file
func (c *Config) Load() error {
	err := createIfNotExist()
	if err != nil {
		return err
	}

	file, err := os.Open(getConfigFile())
	if err != nil {
		return err
	}

	b, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	if len(b) != 0 {
		return yaml.Unmarshal(b, c)
	}

	return nil
}

// CheckMissingResourceEnvvars will read the environment for equivalent config variables to set
func (c *Config) CheckMissingResourceEnvvars() {
	if !c.Resource.DaemonSet && os.Getenv("KW_DAEMONSET") == "true" {
		c.Resource.DaemonSet = true
	}
	if !c.Resource.ReplicaSet && os.Getenv("KW_REPLICASET") == "true" {
		c.Resource.ReplicaSet = true
	}
	if !c.Resource.Namespace && os.Getenv("KW_NAMESPACE") == "true" {
		c.Resource.Namespace = true
	}
	if !c.Resource.Deployment && os.Getenv("KW_DEPLOYMENT") == "true" {
		c.Resource.Deployment = true
	}
	if !c.Resource.Pod && os.Getenv("KW_POD") == "true" {
		c.Resource.Pod = true
	}
	if !c.Resource.ReplicationController && os.Getenv("KW_REPLICATION_CONTROLLER") == "true" {
		c.Resource.ReplicationController = true
	}
	if !c.Resource.Services && os.Getenv("KW_SERVICE") == "true" {
		c.Resource.Services = true
	}
	if !c.Resource.Job && os.Getenv("KW_JOB") == "true" {
		c.Resource.Job = true
	}
	if !c.Resource.PersistentVolume && os.Getenv("KW_PERSISTENT_VOLUME") == "true" {
		c.Resource.PersistentVolume = true
	}
	if !c.Resource.Secret && os.Getenv("KW_SECRET") == "true" {
		c.Resource.Secret = true
	}
	if !c.Resource.ConfigMap && os.Getenv("KW_CONFIGMAP") == "true" {
		c.Resource.ConfigMap = true
	}
	if !c.Resource.Ingress && os.Getenv("KW_INGRESS") == "true" {
		c.Resource.Ingress = true
	}
	if !c.Resource.Node && os.Getenv("KW_NODE") == "true" {
		c.Resource.Node = true
	}
	if !c.Resource.ServiceAccount && os.Getenv("KW_SERVICE_ACCOUNT") == "true" {
		c.Resource.ServiceAccount = true
	}
	if !c.Resource.ClusterRole && os.Getenv("KW_CLUSTER_ROLE") == "true" {
		c.Resource.ClusterRole = true
	}
	if !c.Resource.ClusterRoleBinding && os.Getenv("KW_CLUSTER_ROLE_BINDING") == "true" {
		c.Resource.ClusterRoleBinding = true
	}
	if (c.Handler.Slack.Channel == "") && (os.Getenv("SLACK_CHANNEL") != "") {
		c.Handler.Slack.Channel = os.Getenv("SLACK_CHANNEL")
	}
	if (c.Handler.Slack.Token == "") && (os.Getenv("SLACK_TOKEN") != "") {
		c.Handler.Slack.Token = os.Getenv("SLACK_TOKEN")
	}
	if (c.Handler.SlackWebhook.Slackwebhookurl == "") && (os.Getenv("KW_SLACK_WEBHOOK_URL") != "") {
		c.Handler.SlackWebhook.Slackwebhookurl = os.Getenv("KW_SLACK_WEBHOOK_URL")
	}
}

func (c *Config) Write() error {
	f, err := os.OpenFile(getConfigFile(), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := yaml.NewEncoder(f)
	enc.SetIndent(2) // compat with old versions of kubewatch
	return enc.Encode(c)
}

func getConfigFile() string {
	configFile := filepath.Join(configDir(), ConfigFileName)
	if _, err := os.Stat(configFile); err == nil {
		return configFile
	}

	return ""
}

func configDir() string {
	if configDir := os.Getenv("KW_CONFIG"); configDir != "" {
		return configDir
	}

	if runtime.GOOS == "windows" {
		home := os.Getenv("USERPROFILE")
		return home
	}
	return os.Getenv("HOME")
	//path := "/etc/kubewatch"
	//if _, err := os.Stat(path); os.IsNotExist(err) {
	//	os.Mkdir(path, 755)
	//}
	//return path
}
