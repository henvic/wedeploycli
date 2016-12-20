# Collector
Collector is a simple middleware micro service for collecting metrics for the WeDeploy CLI.

It receives the metrics in batch and stores them on a ElasticSearch backend using its bulk API.

## Running
No configuration is needed by default.
It runs on port 4884 and tries to connect to a regular local ElasticSearch instance via HTTP and save metrics to `/we-cli/metrics`.


## Endpoints

### POST /
It expects a body request formed by pure text where each line is a JSON entry. There must be no line breaking for any given JSON.

The structure data is available on the collector.Event type.

Some details:

* ID is generated on the middleware if not received
* Duplicate messages (same ID) are discarded immediately
* RequestID is generated for each request
* IP and XForwardedFor are set on the middleware and related to the request
* Text should be a user-friendly message
* Type is the operation being logged
* Tags and extra can be anything useful for better describing or adding context to a given type
* Raw is the line "as is" received. It is saved even in case of JSON processing failure. Theoretically it is a vector for inconsistencies or abuse, so it should be considered unreliable and used for debugging only