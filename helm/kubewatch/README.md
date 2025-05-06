<!--- app-name: Kubewatch -->

# Kubewatch packaged by Bitnami

Kubewatch is a Kubernetes watcher that currently publishes notification to Slack. Run it in your k8s cluster, and you will get event notifications in a slack channel.

[Overview of Kubewatch](https://github.com/bitnami-labs/kubewatch)

## TL;DR

```console
$ helm repo add robusta https://robusta-charts.storage.googleapis.com && helm repo update
$ helm install my-release robusta/kubewatch
```

## Introduction

This chart bootstraps a kubewatch deployment on a [Kubernetes](https://kubernetes.io) cluster using the [Helm](https://helm.sh) package manager.

## Prerequisites

- Kubernetes 1.19+
- Helm 3.2.0+

## Installing the Chart

To install the chart with the release name `my-release`:

```console
$ helm install my-release bitnami/kubewatch
```

The command deploys kubewatch on the Kubernetes cluster in the default configuration. The [Parameters](#parameters) section lists the parameters that can be configured during installation.

## Uninstalling the Chart

To uninstall/delete the `my-release` deployment:

```console
$ helm delete my-release
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Parameters

### Global parameters

| Name                      | Description                                     | Value |
| ------------------------- | ----------------------------------------------- | ----- |
| `global.imageRegistry`    | Global Docker image registry                    | `""`  |
| `global.imagePullSecrets` | Global Docker registry secret names as an array | `[]`  |


### Common parameters

| Name                     | Description                                                                             | Value          |
| ------------------------ | --------------------------------------------------------------------------------------- | -------------- |
| `kubeVersion`            | Force target Kubernetes version (using Helm capabilities if not set)                    | `""`           |
| `nameOverride`           | String to partially override common.names.fullname template                             | `""`           |
| `fullnameOverride`       | String to fully override common.names.fullname template                                 | `""`           |
| `commonLabels`           | Labels to add to all deployed objects                                                   | `{}`           |
| `commonAnnotations`      | Annotations to add to all deployed objects                                              | `{}`           |
| `diagnosticMode.enabled` | Enable diagnostic mode (all probes will be disabled and the command will be overridden) | `false`        |
| `diagnosticMode.command` | Command to override all containers in the the deployment(s)/statefulset(s)              | `["sleep"]`    |
| `diagnosticMode.args`    | Args to override all containers in the the deployment(s)/statefulset(s)                 | `["infinity"]` |
| `extraDeploy`            | Array of extra objects to deploy with the release                                       | `[]`           |


### Kubewatch parameters

| Name                                     | Description                                                                      | Value                  |
| ---------------------------------------- | -------------------------------------------------------------------------------- | ---------------------- |
| `image.registry`                         | Kubewatch image registry                                                         | `docker.io`            |
| `image.repository`                       | Kubewatch image repository                                                       | `bitnami/kubewatch`    |
| `image.tag`                              | Kubewatch image tag (immutable tags are recommended)                             | `0.1.0-debian-10-r513` |
| `image.pullPolicy`                       | Kubewatch image pull policy                                                      | `IfNotPresent`         |
| `image.pullSecrets`                      | Specify docker-registry secret names as an array                                 | `[]`                   |
| `hostAliases`                            | Add deployment host aliases                                                      | `[]`                   |
| `slack.enabled`                          | Enable Slack notifications                                                       | `true`                 |
| `slack.channel`                          | Slack channel to notify                                                          | `XXXX`                 |
| `slack.token`                            | Slack API token                                                                  | `XXXX`                 |
| `hipchat.enabled`                        | Enable HipChat notifications                                                     | `false`                |
| `hipchat.room`                           | HipChat room to notify                                                           | `""`                   |
| `hipchat.token`                          | HipChat token                                                                    | `""`                   |
| `hipchat.url`                            | HipChat URL                                                                      | `""`                   |
| `webex.enabled`                          | Enable Webex notifications                                                       | `false`                |
| `webex.room`                             | Webex room to notify                                                             | `""`                   |
| `webex.token`                            | Webex token                                                                      | `""`                   |
| `webex.url`                              | Webex URL                                                                        | `""`                   |
| `mattermost.enabled`                     | Enable Mattermost notifications                                                  | `false`                |
| `mattermost.channel`                     | Mattermost channel to notify                                                     | `""`                   |
| `mattermost.url`                         | Mattermost URL                                                                   | `""`                   |
| `mattermost.username`                    | Mattermost user to notify                                                        | `""`                   |
| `flock.enabled`                          | Enable Flock notifications                                                       | `false`                |
| `flock.url`                              | Flock URL                                                                        | `""`                   |
| `msteams.enabled`                        | Enable Microsoft Teams notifications                                             | `false`                |
| `msteams.webhookurl`                     | Microsoft Teams webhook URL                                                      | `""`                   |
| `webhook.enabled`                        | Enable Webhook notifications                                                     | `false`                |
| `webhook.url`                            | Webhook URL                                                                      | `""`                   |
| `smtp.enabled`                           | Enable SMTP (email) notifications                                                | `false`                |
| `smtp.to`                                | Destination email address (required)                                             | `""`                   |
| `smtp.from`                              | Source email address (required)                                                  | `""`                   |
| `smtp.hello`                             | SMTP hello field (optional)                                                      | `""`                   |
| `smtp.smarthost`                         | SMTP server address (name:port) (required)                                       | `""`                   |
| `smtp.subject`                           | Source email subject                                                             | `""`                   |
| `smtp.auth.username`                     | Username for LOGIN and PLAIN auth mech                                           | `""`                   |
| `smtp.auth.password`                     | Password for LOGIN and PLAIN auth mech                                           | `""`                   |
| `smtp.auth.secret`                       | Secret for CRAM-MD5 auth mech                                                    | `""`                   |
| `smtp.auth.identity`                     | Identity for PLAIN auth mech                                                     | `""`                   |
| `smtp.requireTLS`                        | Force STARTTLS. Set to `true` or `false`                                         | `""`                   |
| `namespaceToWatch`                       | Namespace to watch, leave it empty for watching all                              | `""`                   |
| `resourcesToWatch.deployment`            | Watch changes to Deployments                                                     | `true`                 |
| `resourcesToWatch.replicationcontroller` | Watch changes to ReplicationControllers                                          | `false`                |
| `resourcesToWatch.replicaset`            | Watch changes to ReplicaSets                                                     | `false`                |
| `resourcesToWatch.daemonset`             | Watch changes to DaemonSets                                                      | `false`                |
| `resourcesToWatch.services`              | Watch changes to Services                                                        | `false`                |
| `resourcesToWatch.pod`                   | Watch changes to Pods                                                            | `true`                 |
| `resourcesToWatch.job`                   | Watch changes to Jobs                                                            | `false`                |
| `resourcesToWatch.persistentvolume`      | Watch changes to PersistentVolumes                                               | `false`                |
| `command`                                | Override default container command (useful when using custom images)             | `[]`                   |
| `args`                                   | Override default container args (useful when using custom images)                | `[]`                   |
| `lifecycleHooks`                         | for the Kubewatch container(s) to automate configuration before or after startup | `{}`                   |
| `extraEnvVars`                           | Extra environment variables to be set on Kubewatch container                     | `[]`                   |
| `extraEnvVarsCM`                         | Name of existing ConfigMap containing extra env vars                             | `""`                   |
| `extraEnvVarsSecret`                     | Name of existing Secret containing extra env vars                                | `""`                   |


### Kubewatch deployment parameters

| Name                                    | Description                                                                               | Value           |
| --------------------------------------- | ----------------------------------------------------------------------------------------- | --------------- |
| `replicaCount`                          | Number of Kubewatch replicas to deploy                                                    | `1`             |
| `podSecurityContext.enabled`            | Enable Kubewatch containers' SecurityContext                                              | `false`         |
| `podSecurityContext.fsGroup`            | Set Kubewatch containers' SecurityContext fsGroup                                         | `""`            |
| `containerSecurityContext.enabled`      | Enable Kubewatch pods' Security Context                                                   | `false`         |
| `containerSecurityContext.runAsUser`    | Set Kubewatch pods' SecurityContext runAsUser                                             | `""`            |
| `containerSecurityContext.runAsNonRoot` | Set Kubewatch pods' SecurityContext runAsNonRoot                                          | `""`            |
| `resources.limits`                      | The resources limits for the Kubewatch container                                          | `{}`            |
| `resources.requests`                    | The requested resources for the Kubewatch container                                       | `{}`            |
| `startupProbe.enabled`                  | Enable startupProbe                                                                       | `false`         |
| `startupProbe.initialDelaySeconds`      | Initial delay seconds for startupProbe                                                    | `10`            |
| `startupProbe.periodSeconds`            | Period seconds for startupProbe                                                           | `10`            |
| `startupProbe.timeoutSeconds`           | Timeout seconds for startupProbe                                                          | `1`             |
| `startupProbe.failureThreshold`         | Failure threshold for startupProbe                                                        | `3`             |
| `startupProbe.successThreshold`         | Success threshold for startupProbe                                                        | `1`             |
| `livenessProbe.enabled`                 | Enable livenessProbe                                                                      | `false`         |
| `livenessProbe.initialDelaySeconds`     | Initial delay seconds for livenessProbe                                                   | `10`            |
| `livenessProbe.periodSeconds`           | Period seconds for livenessProbe                                                          | `10`            |
| `livenessProbe.timeoutSeconds`          | Timeout seconds for livenessProbe                                                         | `1`             |
| `livenessProbe.failureThreshold`        | Failure threshold for livenessProbe                                                       | `3`             |
| `livenessProbe.successThreshold`        | Success threshold for livenessProbe                                                       | `1`             |
| `readinessProbe.enabled`                | Enable readinessProbe                                                                     | `false`         |
| `readinessProbe.initialDelaySeconds`    | Initial delay seconds for readinessProbe                                                  | `10`            |
| `readinessProbe.periodSeconds`          | Period seconds for readinessProbe                                                         | `10`            |
| `readinessProbe.timeoutSeconds`         | Timeout seconds for readinessProbe                                                        | `1`             |
| `readinessProbe.failureThreshold`       | Failure threshold for readinessProbe                                                      | `3`             |
| `readinessProbe.successThreshold`       | Success threshold for readinessProbe                                                      | `1`             |
| `customStartupProbe`                    | Override default startup probe                                                            | `{}`            |
| `customLivenessProbe`                   | Override default liveness probe                                                           | `{}`            |
| `customReadinessProbe`                  | Override default readiness probe                                                          | `{}`            |
| `podAffinityPreset`                     | Pod affinity preset. Ignored if `affinity` is set. Allowed values: `soft` or `hard`       | `""`            |
| `podAntiAffinityPreset`                 | Pod anti-affinity preset. Ignored if `affinity` is set. Allowed values: `soft` or `hard`  | `soft`          |
| `nodeAffinityPreset.type`               | Node affinity preset type. Ignored if `affinity` is set. Allowed values: `soft` or `hard` | `""`            |
| `nodeAffinityPreset.key`                | Node label key to match. Ignored if `affinity` is set.                                    | `""`            |
| `nodeAffinityPreset.values`             | Node label values to match. Ignored if `affinity` is set.                                 | `[]`            |
| `affinity`                              | Affinity for pod assignment                                                               | `{}`            |
| `nodeSelector`                          | Node labels for pod assignment                                                            | `{}`            |
| `tolerations`                           | Tolerations for pod assignment                                                            | `[]`            |
| `priorityClassName`                     | Controller priorityClassName                                                              | `""`            |
| `schedulerName`                         | Name of the k8s scheduler (other than default)                                            | `""`            |
| `topologySpreadConstraints`             | Topology Spread Constraints for pod assignment                                            | `[]`            |
| `podLabels`                             | Extra labels for Kubewatch pods                                                           | `{}`            |
| `podAnnotations`                        | Annotations for Kubewatch pods                                                            | `{}`            |
| `extraVolumes`                          | Optionally specify extra list of additional volumes for Kubewatch pods                    | `[]`            |
| `extraVolumeMounts`                     | Optionally specify extra list of additional volumeMounts for Kubewatch container(s)       | `[]`            |
| `updateStrategy.type`                   | Deployment strategy type.                                                                 | `RollingUpdate` |
| `initContainers`                        | Add additional init containers to the Kubewatch pods                                      | `[]`            |
| `sidecars`                              | Add additional sidecar containers to the Kubewatch pods                                   | `[]`            |


### RBAC parameters

| Name                                          | Description                                                                                                         | Value   |
| --------------------------------------------- | ------------------------------------------------------------------------------------------------------------------- | ------- |
| `rbac.create`                                 | Whether to create & use RBAC resources or not                                                                       | `false` |
| `serviceAccount.create`                       | Specifies whether a ServiceAccount should be created                                                                | `true`  |
| `serviceAccount.name`                         | Name of the service account to use. If not set and create is true, a name is generated using the fullname template. | `""`    |
| `serviceAccount.automountServiceAccountToken` | Automount service account token for the server service account                                                      | `true`  |
| `serviceAccount.annotations`                  | Annotations for service account. Evaluated as a template. Only used if `create` is `true`.                          | `{}`    |


Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`. For example,

```console
$ helm install my-release bitnami/kubewatch \
  --set=slack.channel="#bots",slack.token="XXXX-XXXX-XXXX"
```

Alternatively, a YAML file that specifies the values for the above parameters can be provided while installing the chart. For example,

```console
$ helm install my-release -f values.yaml bitnami/kubewatch
```

> **Tip**: You can use the default [values.yaml](values.yaml)

## Configuration and installation details

### [Rolling VS Immutable tags](https://docs.bitnami.com/containers/how-to/understand-rolling-tags-containers/)

It is strongly recommended to use immutable tags in a production environment. This ensures your deployment does not change automatically if the same tag is updated with a different image.

Bitnami will release a new chart updating its containers if a new version of the main container, significant changes, or critical vulnerabilities exist.

### Create a Slack bot

Open [https://my.slack.com/services/new/bot](https://my.slack.com/services/new/bot) to create a new Slack bot.
The API token can be found on the edit page (it starts with `xoxb-`).

Invite the Bot to your channel by typing `/join @name_of_your_bot` in the Slack message area.

### Adding extra environment variables

In case you want to add extra environment variables (useful for advanced operations like custom init scripts), you can use the `extraEnvVars` property.

```yaml
extraEnvVars:
  - name: LOG_LEVEL
    value: debug
```

Alternatively, you can use a ConfigMap or a Secret with the environment variables. To do so, use the `extraEnvVarsCM` or the `extraEnvVarsSecret` values.

### Sidecars and Init Containers

If you have a need for additional containers to run within the same pod as the Kubewatch app (e.g. an additional metrics or logging exporter), you can do so via the `sidecars` config parameter. Simply define your container according to the Kubernetes container spec.

```yaml
sidecars:
  - name: your-image-name
    image: your-image
    imagePullPolicy: Always
    ports:
      - name: portname
       containerPort: 1234
```

Similarly, you can add extra init containers using the `initContainers` parameter.

```yaml
initContainers:
  - name: your-image-name
    image: your-image
    imagePullPolicy: Always
    ports:
      - name: portname
        containerPort: 1234
```

### Deploying extra resources

There are cases where you may want to deploy extra objects, such a ConfigMap containing your app's configuration or some extra deployment with a micro service used by your app. For covering this case, the chart allows adding the full specification of other objects using the `extraDeploy` parameter.

### Setting Pod's affinity

This chart allows you to set your custom affinity using the `affinity` parameter. Find more information about Pod's affinity in the [kubernetes documentation](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#affinity-and-anti-affinity).

As an alternative, you can use of the preset configurations for pod affinity, pod anti-affinity, and node affinity available at the [bitnami/common](https://github.com/bitnami/charts/tree/master/bitnami/common#affinities) chart. To do so, set the `podAffinityPreset`, `podAntiAffinityPreset`, or `nodeAffinityPreset` parameters.

## Troubleshooting

Find more information about how to deal with common errors related to Bitnami's Helm charts in [this troubleshooting guide](https://docs.bitnami.com/general/how-to/troubleshoot-helm-chart-issues).

## Upgrading

### To 3.0.0

- Chart labels were adapted to follow the [Helm charts standard labels](https://helm.sh/docs/chart_best_practices/labels/#standard-labels).
- This version also introduces `bitnami/common`, a [library chart](https://helm.sh/docs/topics/library_charts/#helm) as a dependency. More documentation about this new utility could be found [here](https://github.com/bitnami/charts/tree/master/bitnami/common#bitnami-common-library-chart). Please, make sure that you have updated the chart dependencies before executing any upgrade.

Consequences:

- Backwards compatibility is not guaranteed. To upgrade to `3.0.0`, install a new release of the Kubewatch chart.

### To 2.0.0

[On November 13, 2020, Helm v2 support was formally finished](https://github.com/helm/charts#status-of-the-project), this major version is the result of the required changes applied to the Helm Chart to be able to incorporate the different features added in Helm v3 and to be consistent with the Helm project itself regarding the Helm v2 EOL.

**What changes were introduced in this major version?**

- Previous versions of this Helm Chart use `apiVersion: v1` (installable by both Helm 2 and 3), this Helm Chart was updated to `apiVersion: v2` (installable by Helm 3 only). [Here](https://helm.sh/docs/topics/charts/#the-apiversion-field) you can find more information about the `apiVersion` field.
- The different fields present in the *Chart.yaml* file has been ordered alphabetically in a homogeneous way for all the Bitnami Helm Charts

**Considerations when upgrading to this version**

- If you want to upgrade to this version from a previous one installed with Helm v3, you shouldn't face any issues
- If you want to upgrade to this version using Helm v2, this scenario is not supported as this version doesn't support Helm v2 anymore
- If you installed the previous version with Helm v2 and wants to upgrade to this version with Helm v3, please refer to the [official Helm documentation](https://helm.sh/docs/topics/v2_v3_migration/#migration-use-cases) about migrating from Helm v2 to v3

**Useful links**

- https://docs.bitnami.com/tutorials/resolve-helm2-helm3-post-migration-issues/
- https://helm.sh/docs/topics/v2_v3_migration/
- https://helm.sh/blog/migrate-from-helm-v2-to-helm-v3/

### To 1.0.0

Helm performs a lookup for the object based on its group (apps), version (v1), and kind (Deployment). Also known as its GroupVersionKind, or GVK. Changing the GVK is considered a compatibility breaker from Kubernetes' point of view, so you cannot "upgrade" those objects to the new GVK in-place. Earlier versions of Helm 3 did not perform the lookup correctly which has since been fixed to match the spec.

In https://github.com/helm/charts/pull/17285 the `apiVersion` of the deployment resources was updated to `apps/v1` in tune with the api's deprecated, resulting in compatibility breakage.

This major version signifies this change.

## License

Copyright &copy; 2022 Bitnami

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
