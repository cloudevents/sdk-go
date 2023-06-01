Gin Receiver for CloudEvents
----------------------------------

An example of a Gin webframework CloudEvents receiver with a [TektonEvent](https://tekton.dev/docs/pipelines/events/)

# Steps

Get dependencies
```shell
cd samples/
 go get github.com/gin-gonic/gin
 go get github.com/rs/zerolog/log

```

Run the app
```shell
 go run main.go
```

Test a CloudEvent
```shell
curl -v \
 -H "Ce-Id: e7d95c20-6eb4-4614-946d-27b0ce41c7ff" \
 -H "Ce-Source: /apis///namespaces/dimitar//clone-build-n4qhgl" \
 -H "Ce-Subject: clone-build-n4qhgl"  \
 -H "Ce-Specversion: 1.0" \
 -H "Ce-Type: dev.tekton.event.pipelinerun.started.v1" \
 -H "Content-Type: application/json"  \
 -d @event.json http://localhost:8080

...
< HTTP/1.1 200 OK
... 
```

Logs output
```shell
Got an Event: Context Attributes,
  specversion: 1.0
  type: dev.tekton.event.pipelinerun.started.v1
  source: /apis///namespaces/dimitar//clone-build-n4qhgl
  subject: clone-build-n4qhgl
  id: e7d95c20-6eb4-4614-946d-27b0ce41c7ff
  datacontenttype: application/json
Data,
  {
    "pipelineRun": {
      "metadata": {
        "name": "clone-build-n4qhgl",
        "namespace": "dimitar",
        "uid": "44ef2940-b2d9-4ecb-ad12-808a69972f02",

        .....
  }
  [GIN] 2023/02/13 - 14:16:51 | 200 |       639.6µs |       127.0.0.1 | POST     "/"
```
