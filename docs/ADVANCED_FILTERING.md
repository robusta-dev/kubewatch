# Advanced Filtering for Kubewatch

## Overview

The advanced filtering feature allows Kubewatch to filter out irrelevant Kubernetes events before sending them to Robusta, significantly reducing the performance impact on the Robusta system. This feature is particularly useful when monitoring large Kubernetes clusters where many events are generated but only a subset are actually relevant.

## Configuration

The filtering mechanism is controlled via the environment variable `ADVANCED_FILTERS`:

```bash
export ADVANCED_FILTERS=true  # Enable advanced filtering
export ADVANCED_FILTERS=false # Disable advanced filtering (default)
```

When not set, the feature defaults to `false` (disabled), maintaining backward compatibility.

## Filtering Rules

When advanced filtering is enabled, the following rules are applied:

### Event Resources (api/v1/Event and events.k8s.io/v1/Event)

- **Sent**:
  - Warning events that are Created
  - Any event with Reason "Evicted" (regardless of Type - Normal or Warning)
- **Filtered**:
  - Normal events (unless Reason is "Evicted")
  - Warning events with Update or Delete operations

### Job Resources

- **Always Sent**:
  - Create events
  - Delete events

- **Conditionally Sent** (Update events):
  - When the Job spec changes
  - When the Job fails (status condition contains "Failed")

- **Filtered**: Update events without spec changes or failures

### Pod Resources

- **Always Sent**:
  - Create events
  - Delete events

- **Conditionally Sent** (Update events):
  - When the Pod spec changes
  - When any container (including init containers) has restartCount > 0
  - When any container is waiting with reason "ImagePullBackOff"
  - When the Pod is evicted
  - When any container is terminated with reason "OOMKilled"

- **Filtered**: Update events without any of the above conditions

### All Other Resources

All events for resources not explicitly mentioned above (e.g., Deployments, Services, ConfigMaps, etc.) are sent without filtering.

## Implementation Details

The filtering logic is implemented in the `pkg/filter` package and integrated into the CloudEvent handler. The filter evaluates each event before it's sent to Robusta, checking the resource type and event characteristics against the defined rules.

### Key Components:

1. **Filter Package** (`pkg/filter/filter.go`): Contains the core filtering logic
2. **CloudEvent Handler Integration**: The filter is initialized and used in the CloudEvent handler
3. **Environment Variable**: `ADVANCED_FILTERS` controls whether filtering is active

## Usage Example

### Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kubewatch
spec:
  template:
    spec:
      containers:
      - name: kubewatch
        image: kubewatch:latest
        env:
        - name: ADVANCED_FILTERS
          value: "true"
        - name: KW_CLOUDEVENT_URL
          value: "https://your-robusta-endpoint"
```

### Docker Run

```bash
docker run -d \
  -e ADVANCED_FILTERS=true \
  -e KW_CLOUDEVENT_URL=https://your-robusta-endpoint \
  kubewatch:latest
```

## Testing

To run the filter tests:

```bash
cd pkg/filter
go test -v
```

## Performance Impact

With advanced filtering enabled, you can expect:

- **Reduced Network Traffic**: Fewer events sent to Robusta
- **Lower CPU/Memory Usage**: Robusta processes fewer irrelevant events
- **Improved Signal-to-Noise Ratio**: Only meaningful events are forwarded

## Debugging

When debugging is enabled (log level set to debug), filtered events are logged with details about why they were filtered:

```
DEBU[0001] Filtering out Event resource - type: Normal (only Warning events are sent)
DEBU[0002] Filtering out Pod update event - no significant changes detected
```

## Migration Guide

1. **Test in Development**: Enable filtering in a development environment first
2. **Monitor Metrics**: Compare the number of events before and after enabling filtering
3. **Gradual Rollout**: Enable filtering on a subset of Kubewatch instances before full deployment
4. **Verify Coverage**: Ensure important events are still being captured

## Troubleshooting

### Events Not Being Received

If expected events are not reaching Robusta after enabling filtering:

1. Check the ADVANCED_FILTERS environment variable is set correctly
2. Review the filtering rules to ensure your use case is covered
3. Enable debug logging to see which events are being filtered
4. Temporarily disable filtering to confirm it's the cause

### All Events Still Being Sent

If filtering appears to have no effect:

1. Verify ADVANCED_FILTERS is set to "true" (string value)
2. Check Kubewatch logs for "Advanced filtering is ENABLED" message
3. Ensure you're using the CloudEvent handler (filtering only works with CloudEvent handler)

## Future Enhancements

Potential improvements to the filtering system:

- Configurable filtering rules via configuration file
- Per-resource-type filtering toggles
- Custom filtering expressions
- Filtering statistics and metrics