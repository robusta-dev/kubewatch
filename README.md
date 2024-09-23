<div align="center">

**This is the official Kubewatch project, [originally by Bitnami](https://github.com/bitnami-labs/kubewatch/), now maintained by [Robusta.dev](https://home.robusta.dev/).**

**Feel free to open issues, raise PRs or talk with us on [Slack](https://bit.ly/robusta-slack)!**

**kubewatch** is a Kubernetes watcher that publishes notification to available collaboration hubs/notification channels. Run it in your k8s cluster, and you will get event notifications through webhooks.

[See the blog post on KubeWatch 2.0 to learn more about how KubeWatch is used.](https://home.robusta.dev/blog/kubewatch-2-0-released)

<img src="./docs/kubewatch-logo.jpeg">

[![GoDoc](https://godoc.org/github.com/bitnami-labs/kubewatch?status.svg)](https://godoc.org/github.com/bitnami-labs/kubewatch) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/bitnami-labs/kubewatch/blob/master/LICENSE)
[![slack robusta](https://img.shields.io/badge/Slack-Join-4A154B?style=flat-square&logo=slack&logoColor=white)](https://bit.ly/robusta-slack)

</div>

# Latest image

```
robustadev/kubewatch:v2.9.0
```

# Usage
```
$ kubewatch -h

Kubewatch: A watcher for Kubernetes

kubewatch is a Kubernetes watcher that publishes notifications
to Slack/hipchat/mattermost/flock channels. It watches the cluster
for resource changes and notifies them through webhooks.

supported webhooks:
 - slack
 - slackwebhook
 - msteams
 - hipchat
 - mattermost
 - flock
 - webhook
 - cloudevent
 - smtp
 - webex

Usage:
  kubewatch [flags]
  kubewatch [command]

Available Commands:
  config      modify kubewatch configuration
  resource    manage resources to be watched
  version     print version

Flags:
  -h, --help   help for kubewatch

Use "kubewatch [command] --help" for more information about a command.

```

# Install

### Cluster Installation
#### Using helm:

When you have helm installed in your cluster, use the following setup:

```console
helm repo add robusta https://robusta-charts.storage.googleapis.com && helm repo update
helm install kubewatch robusta/kubewatch --set='rbac.create=true,slack.channel=#YOUR_CHANNEL,slack.token=xoxb-YOUR_TOKEN,resourcesToWatch.pod=true,resourcesToWatch.daemonset=true'
```

You may also provide a values file instead:

```yaml
rbac:
  create: true
  customRoles:
    - apiGroups: ["monitoring.coreos.com"]
      resources: ["prometheusrules"]
      verbs: ["get", "list", "watch"]
resourcesToWatch:
  deployment: false
  replicationcontroller: false
  replicaset: false
  daemonset: false
  services: true
  pod: true
  job: false
  node: false
  clusterrole: true
  clusterrolebinding: true
  serviceaccount: true
  persistentvolume: false
  namespace: false
  secret: false
  configmap: false
  ingress: false
  coreevent: false
  event: true
customresources:
  - group: monitoring.coreos.com
    version: v1
    resource: prometheusrules
slack:
  channel: '#YOUR_CHANNEL'
  token: 'xoxb-YOUR_TOKEN'
```

And use that:

```console
$ helm upgrade --install kubewatch robusta/kubewatch --values=values-file.yml
```

#### Using kubectl:

In order to run kubewatch in a Kubernetes cluster quickly, the easiest way is for you to create a [ConfigMap](https://github.com/robusta-dev/kubewatch/blob/master/kubewatch-configmap.yaml) to hold kubewatch configuration.

An example is provided at [`kubewatch-configmap.yaml`](https://github.com/robusta-dev/kubewatch/blob/master/kubewatch-configmap.yaml), do not forget to update your own slack channel and token parameters. Alternatively, you could use secrets.

Create k8s configmap:

```console
$ kubectl create -f kubewatch-configmap.yaml
```

Create the [Pod](https://github.com/robusta-dev/kubewatch/blob/master/kubewatch.yaml) directly, or create your own deployment:

```console
$ kubectl create -f kubewatch.yaml
```

A `kubewatch` container will be created along with `kubectl` sidecar container in order to reach the API server.

Once the Pod is running, you will start seeing Kubernetes events in your configured Slack channel. Here is a screenshot:

![slack](./docs/slack.png)

To modify what notifications you get, update the `kubewatch` ConfigMap and turn on and off (true/false) resources or configure any resource of your choosing with customresources (CRDs):

```
resource:
  deployment: false
  replicationcontroller: false
  replicaset: false
  daemonset: false
  services: true
  pod: true
  job: false
  node: false
  clusterrole: false
  clusterrolebinding: false
  serviceaccount: false
  persistentvolume: false
  namespace: false
  secret: false
  configmap: false
  ingress: false
  coreevent: false
  event: true
customresources:
  - group: monitoring.coreos.com
    version: v1
    resource: prometheusrules
```

#### Working with RBAC

Kubernetes Engine clusters running versions 1.6 or higher introduced Role-Based Access Control (RBAC). We can create `ServiceAccount` for it to work with RBAC.

```console
$ kubectl create -f kubewatch-service-account.yaml
```

If you do not have permission to create it, you need to become an admin first. For example, in GKE you would run:

```
$ kubectl create clusterrolebinding cluster-admin-binding --clusterrole=cluster-admin --user=REPLACE_EMAIL_HERE
```

Edit `kubewatch.yaml`, and create a new field under `spec` with `serviceAccountName: kubewatch`, you can achieve this by running:

```console
$ sed -i '/spec:/a\ \ serviceAccountName: kubewatch' kubewatch.yaml
```

Then just create `pod` as usual with:

```console
$ kubectl create -f kubewatch.yaml
```

#### Working with CRDs
`kubewatch` can be configured to monitor Kubernetes Custom Resource Definitions (CRDs), allowing you to receive notifications when changes occur.
To configure kubewatch to watch custom resources, you need to define the `customresources` section either in your values file or by using the `--set` flag with Helm commands. 

Include the custom resource configuration in your values file:

```yaml
customresources:
  - group: monitoring.coreos.com
    version: v1
    resource: prometheusrules
```

Then deploy or upgrade `kubwatch` with `helm upgrade` or `helm install`


Alternatively, you can pass this configuration directly using the `--set` flag:

```console
helm install kubewatch robusta/kubewatch --set='rbac.create=true,slack.channel=#YOUR_CHANNEL,slack.token=xoxb-YOUR_TOKEN,resourcesToWatch.pod=true,resourcesToWatch.daemonset=true,customresources[0].group=monitoring.coreos.com,customresources[0].version=v1,customresources[0].resource=prometheusrules'
```
#### Custom RBAC roles
After defining custom resources, make sure that kubewatch has the necessary RBAC permissions to access the custom resources you've configured. Without the appropriate permissions, `kubewatch` will not be able to monitor your custom resources, and you won't receive notifications for changes.

To grant these permissions, you can define custom RBAC roles using `customRoles` within the `rbac` section of your values file or by using the `--set` flag with Helm commands. This allows you to specify exactly which API groups, resources, and actions kubewatch should have access to.

Here’s how you can configure the necessary permissions to monitor your resources:
```yaml
rbac:
  create: true 
  customRoles:
    - apiGroups: ["monitoring.coreos.com"]
      resources: ["prometheusrules"]
      verbs: ["get", "list", "watch"]
```

Then deploy or upgrade `kubwatch` with `helm upgrade` or `helm install`


Alternatively, you can pass this configuration directly using the `--set` flag:

```console
helm install kubewatch robusta/kubewatch --set='rbac.create=true,slack.channel=#YOUR_CHANNEL,slack.token=xoxb-YOUR_TOKEN,customRoles[0].apiGroups={monitoring.coreos.com},customRoles[0].resources={prometheusrules},customRoles[0].verbs={get,list,watch}'
```

#### Metrics
`kubewatch` runs a Prometheus metrics endpoint at `/metrics` on port `2112` by default. This endpoint can be used to monitor health and the performance of `kubewatch`. 

The `kubewatch_events_total` metric can help track the total number of Kubernetes events, categorized by resource type (e.g., `Pods`, `Deployments`) and event type (e.g., `Create`, `Delete`).

You can change the default port (`2112`) on which the metrics server listens by setting the `LISTEN_ADDRESS` environment variable. 
Format is `host:port`. `:5454` means any host, and port `5454`


```yaml
extraEnvVars:
  - name: LISTEN_ADDRESS
    value: ":5454"
```

### Local Installation
#### Using go package installer:

```console
# Download and install kubewatch
$ go get -u github.com/robusta-dev/kubewatch

# Configure the notification channel
$ kubewatch config add slack --channel <slack_channel> --token <slack_token>

# Add resources to be watched
$ kubewatch resource add --po --svc
INFO[0000] resource svc configured
INFO[0000] resource po configured

# start kubewatch server
$ kubewatch
INFO[0000] Starting kubewatch controller                 pkg=kubewatch-service
INFO[0000] Starting kubewatch controller                 pkg=kubewatch-pod
INFO[0000] Processing add to service: default/kubernetes  pkg=kubewatch-service
INFO[0000] Processing add to service: kube-system/tiller-deploy  pkg=kubewatch-service
INFO[0000] Processing add to pod: kube-system/tiller-deploy-69ffbf64bc-h8zxm  pkg=kubewatch-pod
INFO[0000] Kubewatch controller synced and ready         pkg=kubewatch-service
INFO[0000] Kubewatch controller synced and ready         pkg=kubewatch-pod

```
#### Using Docker:

To Run Kubewatch Container interactively, place the config file in `$HOME/.kubewatch.yaml` location and use the following command.

```
docker run --rm -it --network host -v $HOME/.kubewatch.yaml:/root/.kubewatch.yaml -v $HOME/.kube/config:/opt/bitnami/kubewatch/.kube/config --name <container-name> robustadev/kubewatch
```

Example:

```
$ docker run --rm -it --network host -v $HOME/.kubewatch.yaml:/root/.kubewatch.yaml -v $HOME/.kube/config:/opt/bitnami/kubewatch/.kube/config --name kubewatch-app robustadev/kubewatch

==> Writing config file...
INFO[0000] Starting kubewatch controller                 pkg=kubewatch-service
INFO[0000] Starting kubewatch controller                 pkg=kubewatch-pod
INFO[0000] Starting kubewatch controller                 pkg=kubewatch-deployment
INFO[0000] Starting kubewatch controller                 pkg=kubewatch-namespace
INFO[0000] Processing add to namespace: kube-node-lease  pkg=kubewatch-namespace
INFO[0000] Processing add to namespace: kube-public      pkg=kubewatch-namespace
INFO[0000] Processing add to namespace: kube-system      pkg=kubewatch-namespace
INFO[0000] Processing add to namespace: default          pkg=kubewatch-namespace
....
```

To Demonise Kubewatch container use

```
$ docker run --rm -d --network host -v $HOME/.kubewatch.yaml:/root/.kubewatch.yaml -v $HOME/.kube/config:/opt/bitnami/kubewatch/.kube/config --name kubewatch-app robustadev/kubewatch
```

# Configure

Kubewatch supports `config` command for configuration. Config file will be saved at `$HOME/.kubewatch.yaml`

```
$ kubewatch config -h

config command allows admin setup his own configuration for running kubewatch

Usage:
  kubewatch config [flags]
  kubewatch config [command]

Available Commands:
  add         add webhook config to .kubewatch.yaml
  test        test handler config present in .kubewatch.yaml
  view        view .kubewatch.yaml

Flags:
  -h, --help   help for config

Use "kubewatch config [command] --help" for more information about a command.
```
### Example:

### slack:

- Create a [slack Bot](https://my.slack.com/services/new/bot)

- Edit the Bot to customize its name, icon and retrieve the API token (it starts with `xoxb-`).

- Invite the Bot into your channel by typing: `/invite @name_of_your_bot` in the Slack message area.

- Add Api token to kubewatch config using the following steps

  ```console
  $ kubewatch config add slack --channel <slack_channel> --token <slack_token>
  ```
  You have an altenative choice to set your SLACK token, channel via environment variables:

  ```console
  $ export KW_SLACK_TOKEN='XXXXXXXXXXXXXXXX'
  $ export KW_SLACK_CHANNEL='#channel_name'
  ```

### slackwebhookurl:

- Create a [slack app](https://api.slack.com/apps/new)

- Enable [Incoming Webhooks](https://api.slack.com/messaging/webhooks#enable_webhooks). (On "Settings" page.)

- Create an incoming webhook URL (Add New Webhook to Workspace on "Settings" page.)

- Pick a channel that the app will post to, and then click to Authorize your app. You will get back your webhook URL.  
  The Slack Webhook URL will look like: https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX

- Add slack webhook url to kubewatch config using the following steps

  ```console
  $ kubewatch config add slackwebhookurl --username <slack_username> --emoji <slack_emoji> --channel <slack_channel> --slackwebhookurl <slack_webhook_url>
  ```
  Or, you have an altenative choice to set your SLACK channel, username, emoji and webhook URL via environment variables:

  ```console
  $ export KW_SLACK_CHANNEL=slack_channel
  $ export KW_SLACK_USERNAME=slack_username
  $ export KW_SLACK_EMOJI=slack_emoji
  $ export KW_SLACK_WEBHOOK_URL=slack_webhook_url
  ```
  
 - Example apply done in a bash script:  
  
 ```console
 $ cat kubewatch-configmap-slackwebhook.yaml | sed "s|<slackchannel>|"\"$SlackChannel"\"|g;s|<slackusername>|"\"$SlackUsesrName"\"|g;s|<slackemoji>|"\"$SlackEmoji"\"|g;s|<SlackWebhookUrl>|"\"$WebhookUrl"\"|g" | kubectl create -f -
 ```
 
 - An example kubewatch-configmap-slackwebhook.yaml YAML File:  
  
 ```yaml
 apiVersion: v1
kind: ConfigMap
metadata:
  name: kubewatch
data:
  .kubewatch.yaml: |
    namespace: ""
    handler:
      slackwebhook:
        enabled: true
        channel: <slackchannel>
        username: <slackusername>
        emoji: <slackemoji>
        slackwebhookurl: <SlackWebhookUrl>
    resource:
      clusterrole: false
      configmap: false
      daemonset: false
      deployment: true
      ingress: false
      job: false
      namespace: false
      node: false
      persistentvolume: false
      pod: true
      replicaset: false
      replicationcontroller: false
      secret: false
      serviceaccount: false
      services: true
      event: true
      coreevent: false
    ```

### flock:

- Create a [flock bot](https://docs.flock.com/display/flockos/Bots).

- Add flock webhook url to config using the following command.
  ```console
  $ kubewatch config add flock --url <flock_webhook_url>
  ```
  You have an altenative choice to set your FLOCK URL

  ```console
  $ export KW_FLOCK_URL='https://api.flock.com/hooks/sendMessage/XXXXXXXX'
  ```

## Testing Config

To test the handler config by send test messages use the following command.
```
$ kubewatch config test -h

Tests handler configs present in .kubewatch.yaml by sending test messages

Usage:
  kubewatch config test [flags]

Flags:
  -h, --help   help for test
```

#### Example:

```
$ kubewatch config test

Testing Handler configs from .kubewatch.yaml
2019/06/03 12:29:23 Message successfully sent to channel ABCD at 1559545162.000100
```

## Viewing config
To view the entire config file `$HOME/.kubewatch.yaml` use the following command.
```
$ kubewatch config view
Contents of .kubewatch.yaml

handler:
  slack:
    token: xoxb-xxxxx-yyyy-zzz
    channel: kube-watch
  hipchat:
    token: ""
    room: ""
    url: ""
  webex:
    token: ""
    room: ""
    url: ""
  mattermost:
    channel: ""
    url: ""
    username: ""
  flock:
    url: ""
  webhook:
    url: ""
  cloudevent:
    url: ""
resource:
  deployment: false
  replicationcontroller: false
  replicaset: false
  daemonset: false
  services: false
  pod: true
  job: false
  node: false
  clusterrole: false
  clusterrolebinding: false
  serviceaccount: false
  persistentvolume: false
  namespace: false
  secret: false
  configmap: false
  ingress: false
  event: true
  coreevent: false
namespace: ""

```


## Resources

To manage the resources being watched, use the following command, changes will be saved to `$HOME/.kubewatch.yaml`.

```
$ kubewatch resource -h

manage resources to be watched

Usage:
  kubewatch resource [flags]
  kubewatch resource [command]

Available Commands:
  add         adds specific resources to be watched
  remove      remove specific resources being watched

Flags:
      
      --clusterrolebinding      watch for cluster role bindings
      --clusterrole             watch for cluster roles
      --cm                      watch for plain configmaps
      --deploy                  watch for deployments
      --ds                      watch for daemonsets
  -h, --help                    help for resource
      --ing                     watch for ingresses
      --job                     watch for jobs
      --node                    watch for Nodes
      --ns                      watch for namespaces
      --po                      watch for pods
      --pv                      watch for persistent volumes
      --rc                      watch for replication controllers
      --rs                      watch for replicasets
      --sa                      watch for service accounts
      --secret                  watch for plain secrets
      --svc                     watch for services
      --coreevent               watch for events from the kubernetes core api. (Old events api, replaced in kubernetes 1.19)

Use "kubewatch resource [command] --help" for more information about a command.

```

### Add/Remove resource:
```
$ kubewatch resource add -h

adds specific resources to be watched

Usage:
  kubewatch resource add [flags]

Flags:
  -h, --help   help for add

Global Flags:
      --clusterrole             watch for cluster roles
      --clusterrolebinding      watch for cluster role bindings
      --cm                      watch for plain configmaps
      --deploy                  watch for deployments
      --ds                      watch for daemonsets
      --ing                     watch for ingresses
      --job                     watch for jobs
      --node                    watch for Nodes
      --ns                      watch for namespaces
      --po                      watch for pods
      --pv                      watch for persistent volumes
      --rc                      watch for replication controllers
      --rs                      watch for replicasets
      --sa                      watch for service accounts
      --secret                  watch for plain secrets
      --svc                     watch for services
      --coreevent               watch for events from the kubernetes core api. (Old events api, replaced in kubernetes 1.19)

```

### Example:

```console
# rc, po and svc will be watched
$ kubewatch resource add --rc --po --svc

# rc, po and svc will be stopped from being watched
$ kubewatch resource remove --rc --po --svc
```

### Changing log level

In case you want to change the default log level, add an environment variable named `LOG_LEVEL` with value from `trace/debug/info/warning/error` 

```yaml
env:
- name: LOG_LEVEL
  value: debug
```

### Changing log format

In case you want to change the log format to `json`, add an environment variable named `LOG_FORMATTER` with value `json`

```yaml
env:
- name: LOG_FORMATTER
  value: json
```

# Build

### Using go

Clone the repository anywhere:
```console
$ git clone https://github.com/bitnami-labs/kubewatch.git
$ cd kubewatch
$ go build
```
or

You can also use the Makefile directly:

```console
$ make build
```

#### Prerequisites

- You need to have [Go](http://golang.org) (v1.5 or later)  installed. Make sure to set `$GOPATH`


### Using Docker

```console
$ make docker-image
$ docker images
REPOSITORY          TAG                 IMAGE ID            CREATED              SIZE
kubewatch           latest              919896d3cd90        3 minutes ago       27.9MB
```
#### Prerequisites

- you need to have [docker](https://docs.docker.com/) installed.

# Contribution

Refer to the [contribution guidelines](docs/CONTRIBUTION.md) to get started.
