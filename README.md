# Lab 3: Distributed Tracing with OpenTelemetry

## Objective
In this final lab, we implement **Distributed Tracing**. Traces allow us to track a single user request as it travels across the network from the frontend, into the backend, and finally to the database. We will use the OpenTelemetry (OTel) SDK.

Our apps will push their spans to an OpenTelemetry Collector running in our cluster. The collector will then forward all trace data to our STACKIT Observability Tempo instance.

## Helpful Resources
* **[OpenTelemetry Go Documentation](https://opentelemetry.io/docs/languages/go/)**: Review how to initialize the SDK and instrument HTTP clients/routers.
* **[OTel Go Contrib Libraries](https://github.com/open-telemetry/opentelemetry-go-contrib)**: We will use the `net/http/otelhttp` and `gorilla/mux/otelmux` packages from here.

## Prerequisites
1. Ensure your `devcontainer` and `kind` cluster are running (`make cluster-up`).
2. Make sure your credentials are set in `deploy/lab-1` and `deploy/lab-2`.
3. Configure your STACKIT Observability Trace Push credentials in `deploy/lab-3/otel-collector.yaml`. Replace the placeholders (`<YOUR_OBSERVABILITY_OTEL_HTTP_URL>`, `<YOUR_BASE64_ENCODED_OBSERVABILITY_USER_AND_PASSWORD>`).

## Tasks

### Task 1: Initialize the SDK and Middleware (Frontend & Backend)
* **Goal**: Call `initTracerProvider(logger)` (which is provided for you in `tracing.go`) in the `init()` functions of both apps.
* **Goal**: Wrap the Gorilla Mux routers in both applications with `otelmux.Middleware`. This automatically extracts incoming trace headers and creates the parent span.

### Task 2: Instrument the HTTP Client (Frontend)
Go's default `http.Get` and `http.Post` calls do not propagate trace contexts.
* **Goal**: Create a custom `*http.Client` that uses `otelhttp.NewTransport`. 
* **Goal**: Rewrite the backend calls in `frontend-app` to use http requests with context and execute them via your custom client. **Passing the request context is critical for linking the traces!**

### Task 3: Create Manual Spans (Backend)
The middleware tracks the HTTP request, but what if the database is slow? 
* **Goal**: Create manual child spans specifically around your `rdb.Set` and `rdb.Get` Redis calls. Add the Redis key as a span attribute. Don't forget to end them!

## Verification

1. **Deploy your changes:**
   Run `make deploy` to build your code, load it into the cluster, and deploy all manifests including the OTel Collector.
2. **Generate Traffic:**
   Port-forward the frontend app:
   `kubectl port-forward svc/frontend-app-svc 8080:80`
   Generate some data: `curl -X POST -d 'https://stackit.de' http://localhost:8080/shorten`
3. **Verify Local Traces (Optional):**
   If you want to test traces locally without Kubernetes, you can run `make local-run`, which starts a local Jaeger instance at `http://localhost:16686`.
4. **Verify Remote Traces:**
   Log into your STACKIT Grafana instance, go to Explore, select your Tempo data source, and query for your traces. You should see a beautiful waterfall chart visualizing the hop from Frontend -> Backend -> Redis!

---
*Got hopelessly stuck? Ask the person on your left or ask the person on your right. But if everything fails, simply run `git checkout lab-3-solution` to see the completed implementation.*
