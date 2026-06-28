VERSION ?= v1
apply-all-manifestos:
	kubectl apply -f services/budget-planner-service/k8s/ && kubectl apply -f services/auth-service/k8s/
replace-all-manifestos:
	kubectl replace --force -f services/budget-planner-service/k8s/ && kubectl replace --force -f services/auth-service/k8s/
rebuild-all-containers:
	docker build -t dmitriysyworov/auth-user-service:$(VERSION) -f ./services/auth-service/Dockerfile . && docker build -t dmitriysyworov/budget-planner-service:$(VERSION) -f ./services/budget-planner-service/Dockerfile .
push-all-containers:
	docker push dmitriysyworov/auth-user-service:$(VERSION) && docker push dmitriysyworov/budget-planner-service:$(VERSION)
rebuild-and-push-all-containers:
	docker build -t dmitriysyworov/auth-user-service:$(VERSION) -f ./services/auth-service/Dockerfile . && docker build -t dmitriysyworov/budget-planner-service:$(VERSION) -f ./services/budget-planner-service/Dockerfile . && docker push dmitriysyworov/auth-user-service:$(VERSION) && docker push dmitriysyworov/budget-planner-service:$(VERSION)