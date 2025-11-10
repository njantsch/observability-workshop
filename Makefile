# =============================================================================
# Workshop Makefile
#
# Local usage:
# 1. Run 'make run-local' to start Redis.
# 2. Run in two different terminals:
#    - Terminal 1: 'make run-backend-local'
#    - Terminal 2: 'make run-frontend-local'
# 3. Run 'make clean' to stop Redis and remove binaries.
#
# SKE usage:
# 1. Run 'make docker-login' to log into the STACKIT Container Registry.
# 2. Run 'make build-push' to build and push all images.
# 3. Run 'make deploy' to deploy to the k8s cluster configured in K8S_CONTEXT.
#
# =============================================================================

# Configuration
REGISTRY_URL = registry.onstackit.cloud/obs-workshop/$(shell whoami)
K8S_CONTEXT= kind-kind

# List of all applications to build (must match directory names)
APPS = frontend-app backend-app

# Image tag
TAG = latest

# Docker & Local Config
REDIS_CONTAINER_NAME = workshop-redis
BIN_DIR = bin
FRONTEND_BIN = $(BIN_DIR)/frontend-bin
BACKEND_BIN = $(BIN_DIR)/backend-bin
PULL_SECRET_NAME = stackit-registry-secret

.PHONY: all docker-build docker-push build-push clean \
        run-local stop-local run-redis-local stop-redis-local \
        build-local run-frontend-local run-backend-local clean-local\
		deploy create-pull-secret

# Default target
all: build-local

# Builds Application binaries
build-local:
	@mkdir -p $(BIN_DIR)
	@echo "--- Building local Go binaries ---"
	@echo "Building backend-app..."
	@(cd ./backend-app && go build -o ../$(BACKEND_BIN) .)
	@echo "Building frontend-app..."
	@(cd ./frontend-app && go build -o ../$(FRONTEND_BIN) .)

# Starts a local Redis container
run-redis-local:
	@echo "--- Starting local Redis container ($(REDIS_CONTAINER_NAME)) ---"
	@docker run -d --name $(REDIS_CONTAINER_NAME) -p 6379:6379 redis:6-alpine || echo "Redis container already running or failed."
	@echo "Waiting for Redis to start..."
	@sleep 3

# Stops and removes the local Redis container
stop-redis-local:
	@echo "--- Stopping local Redis container ($(REDIS_CONTAINER_NAME)) ---"
	@docker stop $(REDIS_CONTAINER_NAME) || true
	@docker rm $(REDIS_CONTAINER_NAME) || true

# Starts ONLY the Redis container and tells user what to do next.
run-local: run-redis-local
	@echo "\n--- Redis is running ---"
	@echo "Please open TWO new terminal tabs:"
	@echo "In Terminal 1, run: make run-backend-local"
	@echo "In Terminal 2, run: make run-frontend-local"
	@echo "\nPress Ctrl+C in each terminal to stop."
	@echo "Run 'make stop-local' to stop Redis when finished."

# Runs the backend app in the foreground
run-backend-local: build-local
	@echo "--- Starting backend-app (Port 8081) ---"
	@echo "Logs will stream below. Press Ctrl+C to stop."
	@REDIS_ADDR=localhost:6379 ./$(BACKEND_BIN)

# Runs the frontend app in the foreground
run-frontend-local: build-local
	@echo "--- Starting frontend-app (http://localhost:8080) ---"
	@echo "Logs will stream below. Press Ctrl+C to stop."
	@BACKEND_SVC_URL=http://localhost:8081 ./$(FRONTEND_BIN)

# Stops ONLY the Redis container.
stop-local: stop-redis-local
	@echo "--- Local development stopped ---"
	@echo "If you want to clean up binaries, run 'make clean-local'"

# Cleans up local binaries.
clean-local:
	@echo "Cleaning up binaries..."
	@rm -rf $(BIN_DIR)
	@echo "--- Local cleanup complete ---"

# Convenience target to stop redis and clean binaries.
clean: stop-local clean-local

# Builds all Docker images listed in $(APPS).
docker-build:
	@echo "--- Building Docker images ---"
	@for app in $(APPS); do \
		echo "--> Building $$app (./$$app)"; \
		docker build -t $(REGISTRY_URL)/$$app:$(TAG) ./$$app; \
	done
	@echo "--- Build complete ---"

# Pushes all built images to the registry.
docker-push:
	@echo "--- Pushing images to $(REGISTRY_URL) ---"
	@for app in $(APPS); do \
		echo "--> Pushing $(REGISTRY_URL)/$$app:$(TAG)"; \
		docker push $(REGISTRY_URL)/$$app:$(TAG); \
	done
	@echo "--- Push complete ---"

# Convenience target to build and then push.
build-push: docker-build docker-push

# Log into STACKIT Container Registry.
docker-login:
	docker login registry.onstackit.cloud

# Creates a pull-secret and adds it to the default service account
create-pull-secret:
	@echo "--- Checking for pull secret '$(PULL_SECRET_NAME)' in context '$(K8S_CONTEXT)'... ---"
	@if ! kubectl --context $(K8S_CONTEXT) get secret $(PULL_SECRET_NAME) > /dev/null 2>&1; then \
		echo "Secret not found. Starting creation process..."; \
		echo "Please provide your STACKIT registry credentials for the docker-pull-secret."; \
		read -p "Enter Registry Username: " DOCKER_USER; \
		read -s -p "Enter Registry Password or Token: " DOCKER_PASS; \
		echo ""; \
		kubectl --context $(K8S_CONTEXT) create secret docker-registry $(PULL_SECRET_NAME) \
			--docker-server=registry.onstackit.cloud \
			--docker-username=$$DOCKER_USER \
			--docker-password=$$DOCKER_PASS \
			--dry-run=client -o yaml | kubectl --context $(K8S_CONTEXT) apply -f -; \
	else \
		echo "Secret '$(PULL_SECRET_NAME)' already exists. Skipping creation."; \
	fi
	@echo "--- Patching 'default' service account to use the pull secret automatically ---"
	@kubectl --context $(K8S_CONTEXT) patch serviceaccount default -p '{"imagePullSecrets": [{"name": "$(PULL_SECRET_NAME)"}]}'

# Deletes the pull-secret and unpatches the service account
delete-pull-secret:
	@echo "--- Removing patch from 'default' service account in context '$(K8S_CONTEXT)' ---"
	@kubectl --context $(K8S_CONTEXT) patch serviceaccount default --type=json -p='[{"op": "remove", "path": "/imagePullSecrets"}]' || echo "Service account already unpatched or not found."
	@echo "--- Deleting pull secret '$(PULL_SECRET_NAME)' in context '$(K8S_CONTEXT)' ---"
	@kubectl --context $(K8S_CONTEXT) delete secret $(PULL_SECRET_NAME) --ignore-not-found=true

# Deploys the applications to Kubernetes using the REGISTRY_URL.
deploy: create-pull-secret deploy-prometheus deploy-fluent-bit
	@echo "--- Deploying to Kubernetes context: $(K8S_CONTEXT) ---"
	@sed -e "s|__REGISTRY_URL__|$(REGISTRY_URL)|g" deploy/kubernetes.yaml | kubectl --context $(K8S_CONTEXT) apply -f -

deploy-prometheus:
	@echo "--- Deploying Prometheus ---"
	@kubectl --context $(K8S_CONTEXT) apply -f deploy/lab-1/

undeploy-prometheus:
	@echo "--- Deleting Prometheus ---"
	@kubectl --context $(K8S_CONTEXT) delete -f deploy/lab-1/ || true

deploy-fluent-bit:
	@echo "--- Deploying Fluent-Bit ---"
	@kubectl --context $(K8S_CONTEXT) apply -f deploy/lab-2/

undeploy-fluent-bit:
	@echo "--- Deleting Fluent-Bit ---"
	@kubectl --context $(K8S_CONTEXT) delete -f deploy/lab-2/ || true

# Deletes the applications from Kubernetes
undeploy: delete-pull-secret undeploy-prometheus undeploy-fluent-bit
	@echo "--- Deleting setup from Kubernetes context: $(K8S_CONTEXT) ---"
	@sed -e "s|__REGISTRY_URL__|$(REGISTRY_URL)|g" deploy/kubernetes.yaml | kubectl --context $(K8S_CONTEXT) delete -f -