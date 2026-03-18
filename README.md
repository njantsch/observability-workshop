# Lab 2: Structured Logging with slog and Fluent-Bit

## Objective
In this lab, we will convert our unformatted text logs (`log.Printf`) into structured JSON logs using Go's modern `log/slog` package. 

Once our apps output JSON to standard out, a Fluent-Bit agent (running as a DaemonSet in our cluster) will scrape those logs, parse the JSON, and ship them to our central STACKIT Observability Loki instance.

## Helpful Resources
* **[A Guide to the Go slog Package](https://go.dev/blog/slog)**: Learn how to initialize a JSON logger and pass contextual attributes.

## Prerequisites
1. Ensure your `devcontainer` and `kind` cluster are running (`make cluster-up`).
2. Make sure your Prometheus credentials set in `deploy/lab-1/prometheus.yaml`.
3. Configure your STACKIT Observability Log Push credentials in `deploy/lab-2/fluent-bit.yaml`. 
   Replace the placeholders (`<PUSH_URL_DOMAIN>`, `<YOUR_INSTANCE_ID>`, `<YOUR_OBSERVABILITY_API_USERNAME>`, `<YOUR_OBSERVABILITY_API_PASSWORD>`) in the `[OUTPUT]` block.

## Tasks

Open `frontend-app/main.go` and `backend-app/main.go` and look for the `TODO:` blocks.

### Task 1: Initialize the JSON Logger
* **Goal**: Instead of using the default global logger, create a new `slog.Logger` instance that writes JSON to `os.Stdout`. Add a default attribute indicating the service name (e.g., `service=frontend-app` or `service=backend-app`).

### Task 2: Implement Structured Logs
* **Goal**: Go through both files and replace all the old `log.Printf` and `log.Println` calls with the equivalent `Info`, `Warn`, and `Error` calls.

## Verification

1. **Deploy your changes:**
   Run `make deploy` to build your code, load it into the cluster, and apply all manifests (including Fluent-Bit).
2. **Verify Local Output:**
   Run `kubectl logs -l app=frontend-app` and `kubectl logs -l app=backend-app`. 
   You should see clean JSON strings instead of plain text lines.
3. **Generate Traffic:**
   Port-forward the frontend app:
   `kubectl port-forward svc/frontend-app-svc 8080:80`
   Generate some data: `curl -X POST -d 'https://stackit.de' http://localhost:8080/shorten`
4. **Verify Remote Logs:**
   Log into your STACKIT Grafana instance, go to Explore, and query `{job="fluentbit"}`. You should see your JSON logs arriving and fully parsed!

---
*Got hopelessly stuck? Ask the person on your left or ask the person on your right. But if everything fails, simply run `git checkout lab-2-solution` to see the completed implementation.*