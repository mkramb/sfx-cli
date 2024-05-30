# sfx-cli

WORK IN PROGRESS

### Usage

Set required environment variables:

```
export SPLUNK_REALM="eu0"
export SPLUNK_TOKEN="<access-token>"
```

Execute a query to calculate min, max, avg:

```
./sfx-cli "data('k8s.container.memory_request', filter=filter('sf_environment', 'dev') and filter('k8s.container.name', 'api-gateway'), extrapolation='zero').publish()"
```
