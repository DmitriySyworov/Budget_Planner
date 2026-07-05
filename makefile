VERSION ?= v1
helm-test-manifestos:
	helm template ./helm-chart -f ./helm-chart/values.test.yaml
helm-replace-services:
	 helm uninstall budget-app --namespace default && helm install budget-app ./helm-chart -f ./helm-chart/values.yaml --namespace default
helm-upgrade-services:
	helm upgrade budget-app ./helm-chart -f ./helm-chart/values.yaml --namespace default
rebuild-all-containers:
	docker build -t dmitriysyworov/auth-user-service:$(VERSION) -f ./services/auth-service/Dockerfile . && docker build -t dmitriysyworov/budget-planner-service:$(VERSION) -f ./services/budget-planner-service/Dockerfile .
push-all-containers:
	docker push dmitriysyworov/auth-user-service:$(VERSION) && docker push dmitriysyworov/budget-planner-service:$(VERSION)
rebuild-push-all-containers:
	docker build -t dmitriysyworov/auth-user-service:$(VERSION) -f ./services/auth-service/Dockerfile . && docker build -t dmitriysyworov/budget-planner-service:$(VERSION) -f ./services/budget-planner-service/Dockerfile . && docker push dmitriysyworov/auth-user-service:$(VERSION) && docker push dmitriysyworov/budget-planner-service:$(VERSION)
rebuild-push-all-containers-and-replace-all-manifestos:
	docker build -t dmitriysyworov/auth-user-service:$(VERSION) -f ./services/auth-service/Dockerfile . && docker build -t dmitriysyworov/budget-planner-service:$(VERSION) -f ./services/budget-planner-service/Dockerfile . && docker push dmitriysyworov/auth-user-service:$(VERSION) && docker push dmitriysyworov/budget-planner-service:$(VERSION) && kubectl replace --force -f services/budget-planner-service/k8s/ && kubectl replace --force -f services/auth-service/k8s/
get-services-port:
	minikube service ingress-nginx-controller --namespace=ingress-nginx