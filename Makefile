# =======================================================================================
# Testing running system

# Database Access
# dblab --host 0.0.0.0 --user postgres --db postgres --pass postgres --ssl disable --port 5433 --driver postgres

tidy:
	go mod tidy
	go mod vendor

run/service:
	go run ./app/services/yify/

run/desktop:
	go run ./app/desktop/

# =========================================================================================
# Building containers
VERSION := 1.0

all: yify-api
# Give the dockerfile, args and tag.
yify-api:
	docker build \
		-f zarf/docker/dockerfile.yify-api \
		-t yify-api:$(VERSION) \
		--build-arg BUILD_REF=$(VERSION) \
		--build-arg BUILD_DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"` \
		.

# ===================================================
# kind
KIND_CLUSTER := yify-cluster

# start the kind cluster
kind-up:
	kind create cluster \
		--image kindest/node:v1.21.1@sha256:69860bda5563ac81e3c0057d654b5253219618a22ec3a346306239bba8cfa1a6 \
		--name $(KIND_CLUSTER) \
		--config zarf/k8s/kind/kind-config.yaml

# delete the kind cluster
kind-down:
	kind delete cluster --name $(KIND_CLUSTER)

# Load the image to the kind env
# edit the tag of the image to be the current version.
kind-load:
	cd zarf/k8s/kind/yify-pod; kustomize edit set image yify-api-image=yify-api:$(VERSION)
	kind load docker-image yify-api:$(VERSION) --name $(KIND_CLUSTER)

kind-apply:
	kustomize build zarf/k8s/kind/database-pod | kubectl apply -f -
	kubectl wait --namespace=database-system --timeout=120s --for=condition=Available deployment/database-pod
	kustomize build zarf/k8s/kind/yify-pod | kubectl apply -f -
	kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml 
	kubectl wait --namespace ingress-nginx  --for=condition=ready pod --selector=app.kubernetes.io/component=controller --timeout=90s

# apply the changes
kind-update-apply: all kind-load kind-apply

kind-restart:
	kubectl rollout restart deploy yify-pod -n yify-system

# Will build the image, load into the nodes and restart the sales
kind-update: all kind-load kind-restart