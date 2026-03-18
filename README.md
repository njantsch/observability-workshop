# Lab 1: Metrics with Prometheus

## Objective
In this lab, we will take our "black box" frontend service and instrument it to expose **RED metrics** (Rate, Errors, Duration) using the Prometheus Go client. 

Once instrumented, our application will be automatically scraped by the Prometheus instance running in our cluster, which will push the metrics to our central STACKIT Observability instance.

## Helpful Resources
Before you begin, you may want to keep the following documentation open. You will need to search these pages to figure out how to initialize, register, and update metrics:
* **[Prometheus Go Client (pkg.go.dev)](https://pkg.go.dev/github.com/prometheus/client_golang/prometheus)**: To find certain functions.
* **[Prometheus Documentation - Metric Types](https://prometheus.io/docs/concepts/metric_types/)**: To understand the difference between Counters and Histograms.

## Prerequisites
1. Ensure you have your `devcontainer` running.
2. Start your local cluster: `make cluster-up`
3. Configure your STACKIT Observability Push credentials in `deploy/lab-1/prometheus.yaml`. Replace the placeholders (`<YOUR_REMOTE_WRITE_PUSH_URL>`, `<YOUR_OBSERVABILITY_API_USERNAME>`, `<YOUR_OBSERVABILITY_API_PASSWORD>`) with your credentials.

## Tasks

Open the `frontend-app/main.go` file and look for the `TODO:` blocks.

### Task 1: Initialize Prometheus Metrics
We need to define the metrics we want to track globally so our middleware can use them.
* **Goal**: Create a `Counter` for total requests and a `Histogram` for request duration. Register both with prometheus.

### Task 2: Implement the Middleware
Find the `prometheusMiddleware` function. This function intercepts every HTTP request.
* **Goal**: Measure the duration of the request (`time.Since`) and observe it in your Histogram. Increment your Counter. Pass the necessary label values that were defined in the metric.

### Task 3: Apply the Middleware
* **Goal**: Tell the Gorilla Mux router `r` to actually use your newly created `prometheusMiddleware` so it wraps around your routes.

## Verification

1. **Deploy your changes:**
   Run `make deploy` to build your code, load it into the cluster, and apply the manifests.
2. **Generate Traffic:**
   Port-forward the frontend app:
   `kubectl port-forward svc/frontend-app-svc 8080:80`
   Open a new terminal and run a few `curl` commands to generate data:
   `curl -X POST -d 'https://stackit.de' http://localhost:8080/shorten`
3. **Verify Local Metrics:**
   Port-forward the metrics port:
   `kubectl port-forward svc/frontend-app-svc 9090:9090`
   Open `http://localhost:9090/metrics` in your browser. You should see `http_requests_total` and `http_request_duration_seconds` populated!
4. **Verify Remote Metrics:**
   Log into your STACKIT Grafana instance, go to Explore, and query `http_requests_total`. You should see your metrics arriving from your local cluster!

---
*Got hopelessly stuck? Ask the person on your left or ask the person on your right. But if everything fails, simply run `git checkout lab-1-solution` to see the completed implementation.*
