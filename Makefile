# =============================================================================
# Observability Workshop Makefile
# =============================================================================

CLUSTER_NAME = obs-workshop
APPS = frontend-app backend-app
TAG = latest

.PHONY: cluster-up cluster-down build-apps load-apps deploy undeploy local-run local-stop clean

# Local Development
local-run:
	@echo "Starting Local Redis..."
	@docker run -d --name workshop-redis -p 6379:6379 redis:6-alpine || true
	@echo "\nRedis is running. Please open TWO new terminal tabs and run:"
	@echo "Terminal 1: REDIS_ADDR=localhost:6379 go run ./backend-app"
	@echo "Terminal 2: BACKEND_SVC_URL=http://localhost:8081 go run ./frontend-app"

local-stop:
	@echo "Stopping Local Redis..."
	@docker rm -f workshop-redis || true

# Kubernetes (Kind)
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
		echo "--> Building $$app"; \
		docker build -t obs-workshop/$$app:$(TAG) ./$$app; \
	done

load-apps: build-apps
	@echo "Loading Images into Kind Cluster..."
	@for app in $(APPS); do \
		echo "--> Loading $$app"; \
		kind load docker-image obs-workshop/$$app:$(TAG) --name $(CLUSTER_NAME); \
	done

deploy: load-apps
	@echo "Deploying to Kubernetes..."
	@kubectl apply -f deploy/kubernetes.yaml
	@echo "Applications deployed. Run 'kubectl get pods' to check status."

undeploy:
	@echo "Removing from Kubernetes..."
	@kubectl delete -f deploy/kubernetes.yaml --ignore-not-found=true

clean: local-stop cluster-down
