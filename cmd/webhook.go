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

package cmd

import (
	"strconv"

	"github.com/bitnami-labs/kubewatch/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// webhookConfigCmd represents the webhook subcommand
var webhookConfigCmd = &cobra.Command{
	Use:   "webhook",
	Short: "specific webhook configuration",
	Long:  `specific webhook configuration`,
	Run: func(cmd *cobra.Command, args []string) {
		conf, err := config.New()
		if err != nil {
			logrus.Fatal(err)
		}

		url, err := cmd.Flags().GetString("url")
		if err == nil {
			if len(url) > 0 {
				conf.Handler.Webhook.Url = url
			}
		} else {
			logrus.Fatal(err)
		}

		cert, err := cmd.Flags().GetString("cert")
		if err == nil {
			if len(cert) > 0 {
				conf.Handler.Webhook.Cert = cert
			}
		} else {
			logrus.Fatal(err)
		}

		tlsSkip, err := cmd.Flags().GetString("tlsskip")
		if err == nil {
			if len(tlsSkip) > 0 {
				skip, err := strconv.ParseBool(tlsSkip)
				if err != nil {
					logrus.Fatal(err)
				}
				conf.Handler.Webhook.TlsSkip = skip
			} else {
				conf.Handler.Webhook.TlsSkip = false
			}
		} else {
			logrus.Fatal(err)
		}

		if err = conf.Write(); err != nil {
			logrus.Fatal(err)
		}
	},
}

func init() {
	webhookConfigCmd.Flags().StringP("url", "u", "", "Specify Webhook url")
	webhookConfigCmd.Flags().StringP("cert", "", "", "Specify Webhook cert path")
	webhookConfigCmd.Flags().StringP("tlsskip", "", "", "Specify whether Webhook skips tls verify; TRUE or FALSE")
}
