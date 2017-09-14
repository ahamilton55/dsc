# DSC
This is a basic log parser for Nginx logs that outputs a basic set of statsd metrics to a log file.

## Running tests and benchmarks
Tests don't take very long so they are part of the build process for the Go binary. To run the tests, just build with the Dockerfile at the root of the repo.

## Docker images
Docker images have been pushed to Docker Hub as public images so that you can easily pull down and create the kubernetes service.

Only a single repo was used for both images and they are separated by tags:

| tag            | image              |
| -------------- | ------------------ |
| latest         | nginx stats parser |
| nginx-test-app | nginx test app     |

The Nginx stats parser container is uses a multi-stage build so that we don't carry all of the Go compiler code which isn't required to run the code. This way we have a small, single layer container but this did require me to update the code to deal with properly redirecting output to `STDOUT` if `/dev/stdout` is passed in as the output location.

## Nginx test app
The nginx test app just simply adds return values for different routes in an Nginx vhost declaration. You can then hit the following endpoints to return return the corresponding status codes:

| route      | status code |
| ---------- | ----------- |
| /          | 200         |
| /redirect  | 301         |
| /not-found | 404         |
| /error     | 500         |

## Nginx test app Kubernetes Service
Can be deployed using:

```
kubectl create -f nginx-test-app.yaml
```

Viewing logs could be done after looking up the pod name and running the following:
```
kubectl logs -f -c nginx-stats <pod_name>
```
