# STACKIT Observability Workshop
Welcome to the STACKIT Observability Workshop! This repository contains the sample application we will be instrumenting today.

The goal of this workshop is to take a "black box" application and make it fully observable by implementing the "Three Pillars of Observability": Metrics, Logs, and Traces.

### 1. Application Architecture

The application is a simple "Link Shortener" built as two microservices and a database:
- **frontend-app (Go)**: The public-facing API that handles user requests.
    - `POST` /shorten: Takes a long URL and asks the backend to generate a short link.
    - `GET` /{shortlink}: Asks the backend for the long URL and performs a redirect.
- **backend-app (Go)**: The internal service that handles the "business logic".
    - `POST` /generate: Generates a random short ID.
    - `GET` /resolve/{id}: Looks up the ID.
- **redis**: A simple Redis database used to store the mapping between short links and long URLs.


### 2. Prerequisites & Environment Setup

1. Ensure you have Docker and your favorite IDE installed.
2. Install `go` with version `>=1.23` via `brew install go`
3. Install `kind` via `brew install kind` or `go install sigs.k8s.io/kind@v0.33.0`


### 3. Running the Application

We use a local Kubernetes cluster (`kind`) to simulate a production-like environment for our telemetry data.

#### Start the Cluster
To provision your local cluster, run:
```bash
make cluster-up
```

#### Build and Deploy
To build the Docker images, load them into the local cluster, and deploy the application manifests, run:
```bash
make deploy
```
Wait a few moments for the pods to start (kubectl get pods).

#### Test the App
Since the application is running inside the `kind` cluster, we need to port-forward the frontend service to access it from our browser:
```bash
kubectl port-forward svc/frontend-app-svc 8080:80
```
Open a new terminal to test the app:
1. **Create a short link:**
`curl -X POST -d 'https://stackit.com' http://localhost:8080/shorten`
*(Output example: "id2861")*
2. **Test the redirect:**
Open your browser and navigate to `http://localhost:8080/id2861` (replace with your ID). You should be redirected to stackit.com!

#### Cleanup
To remove the application from Kubernetes: `make undeploy`
To shut down the entire local cluster: `make cluster-down`

### 4. The Workshop Labs

Throughout this workshop, we will take this "un-instrumented" application and...

Implement Metrics (Prometheus): Add custom metrics (Counters and Histograms) to the Go code to measure request rates, errors, and latencies.

Implement Logging (Loki): Convert the simple text logs to structured (JSON) logs to make them searchable and more powerful.

Implement Tracing (OpenTelemetry): Add distributed tracing to follow a single request as it jumps from the frontend-app to the backend-app and the database, allowing us to pinpoint bottlenecks.

Build a Unified Dashboard: Combine all our new data sources (App Metrics, Logs, Traces) into a single SRE dashboard in STACKIT Grafana.
