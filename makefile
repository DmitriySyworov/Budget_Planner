VERSION_AUTH ?= v1
VERSION_BUDGET ?= v1
KAFKA_USER ?= example_user
KAFKA_PASS ?=example_password
TAIL ?=1000
helm-test-manifestos:
	helm template ./helm-chart -f ./helm-chart/values.test.yaml
helm-install-services:
	helm install budget-app ./helm-chart \
    	-f ./helm-chart/values.yaml \
    	--namespace default \
        --set versions.authUserVersion="$(VERSION_AUTH)" \
        --set versions.budgetPlannerVersion="$(VERSION_BUDGET)"
helm-replace-services:
	helm uninstall budget-app --namespace default && \
  	helm install budget-app ./helm-chart \
        -f ./helm-chart/values.yaml \
        --namespace default \
        --set versions.authUserVersion="$(VERSION_AUTH)" \
        --set versions.budgetPlannerVersion="$(VERSION_BUDGET)"
helm-upgrade-services:
	helm upgrade budget-app ./helm-chart \
    	-f ./helm-chart/values.yaml \
    	--namespace default \
        --cleanup-on-fail \
        --set versions.authUserVersion="$(VERSION_AUTH)" \
        --set versions.budgetPlannerVersion="$(VERSION_BUDGET)"
rebuild-all-containers:
	docker build -t dmitriysyworov/auth-user-service:$(VERSION_AUTH) -f ./services/auth-service/Dockerfile . && \
 	docker build -t dmitriysyworov/budget-planner-service:$(VERSION_BUDGET) -f ./services/budget-planner-service/Dockerfile .
push-all-containers:
	docker push dmitriysyworov/auth-user-service:$(VERSION_AUTH) && \
 	docker push dmitriysyworov/budget-planner-service:$(VERSION_BUDGET)
rebuild-push-all-containers:
	docker build -t dmitriysyworov/auth-user-service:$(VERSION_AUTH) -f ./services/auth-service/Dockerfile . && \
  	docker build -t dmitriysyworov/budget-planner-service:$(VERSION_BUDGET) -f ./services/budget-planner-service/Dockerfile . && \
 	docker push dmitriysyworov/auth-user-service:$(VERSION_AUTH) && \
 	docker push dmitriysyworov/budget-planner-service:$(VERSION_BUDGET)
rebuild-push-all-containers-and-upgrade-all-services:
	docker build -t dmitriysyworov/auth-user-service:$(VERSION_AUTH) -f ./services/auth-service/Dockerfile . && \
 	docker build -t dmitriysyworov/budget-planner-service:$(VERSION_BUDGET) -f ./services/budget-planner-service/Dockerfile . && \
 	docker push dmitriysyworov/auth-user-service:$(VERSION_AUTH) && \
 	docker push dmitriysyworov/budget-planner-service:$(VERSION_BUDGET) && \
	helm upgrade budget-app ./helm-chart \
		-f ./helm-chart/values.yaml \
		--namespace default \
        --cleanup-on-fail \
        --set versions.authUserVersion="$(VERSION_AUTH)" \
        --set versions.budgetPlannerVersion="$(VERSION_BUDGET)"
get-services-port:
	minikube service ingress-nginx-controller --namespace=ingress-nginx
helm-hard-replace-all:
	-helm uninstall my
	-helm uninstall budget-app --namespace default
	-kubectl delete pvc --all --force --grace-period=0
	-kubectl get pvc --no-headers | awk '{print $$1}' | xargs -I {} kubectl patch pvc {} -p '{"metadata":{"finalizers":null}}' --type=merge
	 helm install budget-app ./helm-chart \
		-f ./helm-chart/values.yaml \
		--namespace default \
		--set versions.authUserVersion="$(VERSION_AUTH)" \
		--set versions.budgetPlannerVersion="$(VERSION_BUDGET)"
	 helm template budget-app ./helm-chart -f ./helm-chart/values.yaml --show-only templates/ingress.yaml | kubectl apply -f -
	 helm install my oci://registry-1.docker.io/soldevelo/kafka-chart --version 32.4.4 -f ./helm-chart/values.yaml
	 kubectl wait --namespace default --for=condition=Ready pod/my-kafka-chart-controller-0 --timeout=180s
	 kubectl exec my-kafka-chart-controller-0 -- kafka-configs.sh \
		--bootstrap-server localhost:9094 \
		--entity-type users \
		--entity-name $(KAFKA_USER) \
		--alter \
		--add-config "SCRAM-SHA-512=[password=$(KAFKA_PASS)]" && \
	echo "security.protocol=SASL_PLAINTEXT\nsasl.mechanism=SCRAM-SHA-512\nsasl.jaas.config=org.apache.kafka.common.security.scram.ScramLoginModule required username=\"$(KAFKA_USER)\" password=\"$(KAFKA_PASS)\";" | kubectl exec -i my-kafka-chart-controller-0 -- kafka-topics.sh --bootstrap-server localhost:9092 --command-config /dev/stdin --list

rebuild-push-all-helm-hard-replace-all:
	docker build -t dmitriysyworov/auth-user-service:$(VERSION_AUTH) -f ./services/auth-service/Dockerfile . && \
	docker build -t dmitriysyworov/budget-planner-service:$(VERSION_BUDGET) -f ./services/budget-planner-service/Dockerfile . && \
	docker push dmitriysyworov/auth-user-service:$(VERSION_AUTH) && \
	docker push dmitriysyworov/budget-planner-service:$(VERSION_BUDGET)
	-helm uninstall my
	-helm uninstall budget-app --namespace default
	-kubectl delete pods --all --force --grace-period=0
	-kubectl delete pvc --all --force --grace-period=0
	-kubectl get pvc --no-headers | awk '{print $$1}' | xargs -I {} kubectl patch pvc {} -p '{"metadata":{"finalizers":null}}' --type=merge

	helm install my oci://registry-1.docker.io/soldevelo/kafka-chart --version 32.4.4 -f ./helm-chart/values.yaml
	kubectl wait --namespace default --for=condition=Ready pod/my-kafka-chart-controller-0 --timeout=180s

	helm install budget-app ./helm-chart \
		-f ./helm-chart/values.yaml \
		--namespace default \
		--set versions.authUserVersion="$(VERSION_AUTH)" \
		--set versions.budgetPlannerVersion="$(VERSION_BUDGET)"
	helm template budget-app ./helm-chart -f ./helm-chart/values.yaml --show-only templates/ingress.yaml | kubectl apply -f -

get-logs-auth-user:
	kubectl logs deploy/app-auth-user --tail=$(TAIL) -f | grep -v -E "health|ready"
get-logs-budget-planner:
	kubectl logs deploy/app-budget-planner --tail=$(TAIL) -f | grep -v -E "health|ready"
get-logs-all:
	kubectl logs deploy/app-auth-user,deploy/app-budget-planner --tail=100 -f | grep -v -E "health|ready"
