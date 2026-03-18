# =============================================================================
# Observability Workshop Makefile
# =============================================================================

CLUSTER_NAME = obs-workshop
APPS = frontend-app backend-app
TAG = latest

.PHONY: cluster-up cluster-down build-apps load-apps deploy undeploy local-run local-stop clean

# Local Development (No Kubernetes)
local-run:
	@echo "Starting Local Redis & Jaeger..."
	@docker run -d --name workshop-redis -p 6379:6379 redis:6-alpine || true
	@docker run -d --name workshop-jaeger -e COLLECTOR_OTLP_ENABLED=true -p 16686:16686 -p 4318:4318 jaegertracing/all-in-one:1.53 || true
	@echo "\nLocal infrastructure running. Jaeger UI available at: http://localhost:16686"
	@echo "Please open TWO new terminal tabs and run:"
	@echo "Terminal 1: REDIS_ADDR=localhost:6379 go run ./backend-app"
	@echo "Terminal 2: BACKEND_SVC_URL=http://localhost:8081 go run ./frontend-app"

local-stop:
	@echo "Stopping Local Infrastructure..."
	@docker rm -f workshop-redis workshop-jaeger || true

# Kubernetes (Kind) Workflow
cluster-up:
	@echo "Creating Kind Cluster..."
	@kind create cluster --name $(CLUSTER_NAME) || true
	@echo "Kind cluster '$(CLUSTER_NAME)' is up and running."

cluster-down:
	@echo "Deleting Kind Cluster..."
	@kind delete cluster --name $(CLUSTER_NAME) || true

build-apps:
	@echo "Building Docker Images..."
	@for app in $(APPS); do \
		echo "--> Building $$app..."; \
		docker build -t obs-workshop/$$app:$(TAG) ./$$app; \
	done

load-apps: build-apps
	@echo "Loading Images into Kind Cluster..."
	@for app in $(APPS); do \
		echo "--> Loading $$app..."; \
		kind load docker-image obs-workshop/$$app:$(TAG) --name $(CLUSTER_NAME); \
	done

deploy: load-apps
	@echo "Deploying to Kubernetes..."
	@kubectl apply -f deploy/kubernetes.yaml
	@echo "Deploying Prometheus & Fluent-Bit..."
	@kubectl apply -f deploy/lab-1/
	@kubectl apply -f deploy/lab-2/
	@echo "Deploying OTel Collector..."
	@kubectl apply -f deploy/lab-3/
	@echo "Applications deployed. Run 'kubectl get pods' to check status."

undeploy:
	@echo "Removing from Kubernetes..."
	@kubectl delete -f deploy/kubernetes.yaml --ignore-not-found=true
	@kubectl delete -f deploy/lab-1/ --ignore-not-found=true
	@kubectl delete -f deploy/lab-2/ --ignore-not-found=true
	@kubectl delete -f deploy/lab-3/ --ignore-not-found=true

clean: local-stop cluster-down
