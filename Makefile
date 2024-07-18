IMAGE:="docker.io/log/alot"

check-docker-context:
	@if [ -z "${MINIKUBE_ACTIVE_DOCKERD}" ]; then \
		echo 'Please set docker-desktop as the current context. Run:\n'; \
		echo '  eval $$(minikube docker-env)\n\n'; \
		exit 1; \
	fi

deploy: image

.PHONY: image
image: binary check-docker-context
	docker build -t $(IMAGE) .

.PHONY: binary
binary:
	CGO_ENABLED=0 go build -v .

deploy: image
	kubectl apply -f config.yaml
	kubectl apply -f deployment.yaml
	kubectl delete -l app=logalot pod

delete: check-docker-context
	kubectl delete deployment/logalot

logs: check-docker-context
	kubectl logs -f deployment/logalot
