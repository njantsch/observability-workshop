# STACKIT Observability Workshop: The "Link Shortener" App
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


### 2. Prerequisites

Before you begin, please ensure you have the following tools installed on your machine:

- **Project**: Access to the observability-workshop project in the STACKIT portal.
- **Container Registry**: Access to the obs-workshop project in the STACKIT Container Registry.
- **Git**: To clone this repository.
- **Go**: (Version 1.21+) To build and run the apps locally.
- **Docker**: To run the local Redis database and to build container images.
- **kubectl**: To deploy the application to Kubernetes (SKE).

### 3. Workshop Flow

We have two ways to run the application: locally for fast development and on Kubernetes for the kubernetes environment.

**A. Local Development**

During the labs, you'll need to change code often. To test your changes quickly without building and deploying a Docker image every time, you can run the entire application stack locally.

1. Start the Local Environment:
First, start the Redis database.

- `make run-local`


This command starts the redis container and will wait for your next instructions.

2. Start the Backend:

- `make run-backend-local`


This command will build the backend-app (if needed) and run it in the foreground.

3. Start the Frontend:
Open a new terminal window or tab.

- `make run-frontend-local`


This command will build the frontend-app (if needed) and run it in the foreground.

4. Test Your Local App:
Open a fourth terminal (or use the first one) to test the app:
Create a new short link
`curl -X POST -d 'https://stackit.de' http://localhost:8080/shorten`
Output: e.g., "id2861" (Your ID will be random)
Open the Browser and connect to `http://localhost:8080/id2861`
Or use curl via `curl -L http://localhost:8080/id2861`
This should redirect you to [https://stackit.de](https://stackit.de)


You should now see the log output for this request appear live in your backend and frontend terminals.

5. Stop the Local Environment:

- `make stop-local`

To clean up the compiled binaries, you can run make clean-local.

**B. "Production" Workflow (Kubernetes / SKE)**

This is the workflow for deploying your finished, instrumented application to your SKE cluster.

1. Log in to the STACKIT Container Registry:

- `make docker-login`


2. Build and Push the Images:
This command builds the Docker images for both frontend-app and backend-app, tags them, and pushes them to your registry.

- `make build-push`


3. Deploy to SKE:
This command applies the Kubernetes manifests from the deploy/ directory. It dynamically replaces the __REGISTRY_URL__ placeholder in the kubernetes.yaml with the value from your Makefile.
It also prompts you to enter your registry credentials which will be then used to create a docker-pull-secret in your kubernetes cluster.

- `make deploy`


After a minute, your app will be running in SKE and accessible via a LoadBalancer Service.

### 4. The Workshop Labs

Throughout this workshop, we will take this "un-instrumented" application and...

Implement Metrics (Prometheus): Add custom metrics (Counters and Histograms) to the Go code to measure request rates, errors, and latencies.

Implement Logging (Loki): Convert the simple text logs to structured (JSON) logs to make them searchable and more powerful.

Implement Tracing (OpenTelemetry): Add distributed tracing to follow a single request as it jumps from the frontend-app to the backend-app and the database, allowing us to pinpoint bottlenecks.

Build a Unified Dashboard: Combine all our new data sources (App Metrics, SKE Platform Metrics, Logs, Traces) into a single SRE dashboard in STACKIT Grafana.

### 5. Makefile Quick Reference

#### Local Development

- `make run-local`: Starts the local Redis container (Docker).
- `make run-backend-local`: (Run in new terminal) Builds and runs the backend app in the foreground.
- `make run-frontend-local`: (Run in new terminal) Builds and runs the frontend app in the foreground.
- `make stop-local`: Stops and removes the local Redis container. (Stop apps with Ctrl+C).
- `make build-local`: Builds the local Go binaries (places them in /bin).
- `make clean-local`: Deletes the local Go binaries and temp directories.
- `make clean`: Convenience target to stop redis and clean binaries.

#### Kubernetes (SKE) Deployment

- `make docker-login`: Logs you into the STACKIT container registry.
- `make docker-build`: Builds all container images.
- `make docker-push`: Pushes all container images to your registry.
- `make build-push`: Runs docker-build then docker-push.
- `make create-pull-secret`: Creates the registry secret in your cluster, prompting if needed.
- `make deploy`: Deploys the application to your currently configured Kubernetes cluster.
- `make undeploy`: Deletes the application from your cluster
- `make delete-pull-secret`: Removes the pull secret and patch from your cluster.