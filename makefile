VERSION_AUTH ?= v1
VERSION_BUDGET ?= v1
VERSION_NOTIFICATION ?= v1
REPLICAS_AUTH ?= 3
REPLICAS_BUDGET ?= 3
REPLICAS_NOTIFICATION ?= 3
TAIL ?=1000
build-auth:
	docker build -t dmitriysyworov/auth-user-service:$(VERSION_AUTH) -f ./services/auth/Dockerfile . && \
	docker push dmitriysyworov/auth-user-service:$(VERSION_AUTH)

build-budget:
	docker build -t dmitriysyworov/budget-planner-service:$(VERSION_BUDGET) -f ./services/budget/Dockerfile . && \
	docker push dmitriysyworov/budget-planner-service:$(VERSION_BUDGET)

build-notification:
	docker build -t dmitriysyworov/notification:$(VERSION_NOTIFICATION) -f ./services/notification/Dockerfile . && \
	docker push dmitriysyworov/notification:$(VERSION_NOTIFICATION)

build-all-images:
	@$(MAKE) -j 3 build-auth build-budget build-notification

rebuild-push-all-helm-hard-replace-all: build-all-images
	-helm uninstall my
	-helm uninstall budget-app --namespace default
	-kubectl delete jobs --all --force --grace-period=0
	-kubectl delete pods --all --force --grace-period=0
	-kubectl get pvc --no-headers | awk '{print $$1}' | xargs -I {} kubectl patch pvc {} -p '{"metadata":{"finalizers":null}}' --type=merge
	-kubectl delete pvc --all --force --grace-period=0

	helm install my oci://registry-1.docker.io/soldevelo/kafka-chart --version 32.4.4 -f ./helm-chart/values.yaml
	kubectl wait --namespace default --for=condition=Ready pod/my-kafka-chart-controller-0 --timeout=180s
	sleep 35
	helm install budget-app ./helm-chart \
		-f ./helm-chart/values.yaml \
		--namespace default \
		--set versions.authUserVersion="$(VERSION_AUTH)" \
		--set versions.budgetPlannerVersion="$(VERSION_BUDGET)" \
		--set versions.notificationVersion="$(VERSION_NOTIFICATION)" \
		--set replicasCount.authReplicas=0 \
		--set replicasCount.budgetReplicas=0 \
		--set replicasCount.notificationReplicas=0
	kubectl wait --namespace default --for=condition=complete job/kafka-create-topics-job --timeout=60s
	helm upgrade budget-app ./helm-chart \
		-f ./helm-chart/values.yaml \
		--namespace default \
		--reuse-values \
		--set versions.authUserVersion="$(VERSION_AUTH)" \
		--set versions.budgetPlannerVersion="$(VERSION_BUDGET)" \
		--set versions.notificationVersion="$(VERSION_NOTIFICATION)" \
		--set replicasCount.authReplicas="$(REPLICAS_AUTH)" \
		--set replicasCount.budgetReplicas="$(REPLICAS_BUDGET)" \
		--set replicasCount.notificationReplicas="$(REPLICAS_NOTIFICATION)"
	helm template budget-app ./helm-chart -f ./helm-chart/values.yaml --show-only templates/ingress.yaml | kubectl apply -f -

upgrade-helm-push-all: build-all-images
	helm upgrade budget-app ./helm-chart \
		-f ./helm-chart/values.yaml \
		--namespace default \
		--reuse-values \
		--set versions.authUserVersion="$(VERSION_AUTH)" \
		--set versions.budgetPlannerVersion="$(VERSION_BUDGET)" \
		--set versions.notificationVersion="$(VERSION_NOTIFICATION)" \
		--set replicasCount.authReplicas="$(REPLICAS_AUTH)" \
		--set replicasCount.budgetReplicas="$(REPLICAS_BUDGET)" \
		--set replicasCount.notificationReplicas="$(REPLICAS_NOTIFICATION)"
	helm template budget-app ./helm-chart -f ./helm-chart/values.yaml --show-only templates/ingress.yaml | kubectl apply -f -
get-services-port:
	minikube service ingress-nginx-controller --namespace=ingress-nginx
get-logs-auth-user:
	kubectl logs -l app=app-auth-user --tail=$(TAIL) -f
get-logs-budget-planner:
	kubectl logs -l app=app-budget-planner --tail=$(TAIL) -f
get-logs-notification:
	kubectl logs -l app=notification --tail=$(TAIL) -f

proto-update-all:
	protoc --go_out=. --go_opt=paths=source_relative ./shared/shprotos/event/user.proto
	protoc --go_out=. --go_opt=paths=source_relative ./shared/shprotos/event/letter.proto