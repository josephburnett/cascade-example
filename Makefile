goals = help all a b c setup bench nobench dash shell logs clean
.DEFAULT_GOAL : help
.PHONY : $(goals)
.ONESHELL : $(goals)

help :
	@echo "Usage: make all"
	@echo "Prerequisites:"
	@echo " - kubectl installed and pointing to a Kubernetes cluster"
	@echo " - ko installed and pointing to a repo (https://github.com/google/ko)"

all : a b c

a : SERVICE = a
a : WEIGHT = 80ms
a : GEN = 0
a : LIMIT = 1
a : CPU = 100m
a : DEPS = b

b : SERVICE = b
b : WEIGHT = 80ms
b : GEN = 0
b : LIMIT = 1
b : CPU = 100m
b : DEPS = c

c : SERVICE = c
c : WEIGHT = 80ms
c : GEN = 0
c : LIMIT = 1
c : CPU = 100m
c : DEPS = 

a : export PARAMS = s/SERVICE/$(SERVICE)/g; s/WEIGHT/$(WEIGHT)/g; s/DEPS/$(DEPS)/g; s/CPU/$(CPU)/g; s/GEN/$(GEN)/g; s/LIMIT/$(LIMIT)/g
a : setup
	echo $$PARAMS
	sed "$$PARAMS" deploy/template.yaml | ko apply -f -

b : export PARAMS = s/SERVICE/$(SERVICE)/g; s/WEIGHT/$(WEIGHT)/g; s/DEPS/$(DEPS)/g; s/CPU/$(CPU)/g; s/GEN/$(GEN)/g; s/LIMIT/$(LIMIT)/g
b : setup
	echo $$PARAMS
	sed "$$PARAMS" deploy/template.yaml | ko apply -f -

c : export PARAMS = s/SERVICE/$(SERVICE)/g; s/WEIGHT/$(WEIGHT)/g; s/DEPS/$(DEPS)/g; s/CPU/$(CPU)/g; s/GEN/$(GEN)/g; s/LIMIT/$(LIMIT)/g
c : setup
	echo $$PARAMS
	sed "$$PARAMS" deploy/template.yaml | ko apply -f -

bench : setup
	for i in `seq 1 1`; do kubectl -n cascade-example-bench run ab$$i --image=gcr.io/josephburnett-gke-dev/ab --restart=Always -- -s 120 -t 3600 -n 999999 -c 10 "http://a.cascade-example.svc.cluster.local/" ; done

nobench :
	kubectl delete namespace cascade-example-bench --ignore-not-found

setup :
	kubectl create namespace cascade-example --dry-run=true -o yaml | kubectl apply -f -
	kubectl create namespace cascade-example-bench --dry-run=true -o yaml | kubectl apply -f -

dash :
	watch 'kubectl get svc -n cascade-example; echo; kubectl get hpa -n cascade-example; echo; kubectl get pods -n cascade-example; echo; kubectl get pods -n cascade-example-bench'

shell :
	kubectl delete pod shell --ignore-not-found
	kubectl run --generator=run-pod/v1 -it --rm shell --image=busybox /bin/sh

logs :
	kubectl -n cascade-example logs -f deployment/a cascade-example --since=10m

clean : nobench
	kubectl delete namespace cascade-example --ignore-not-found




gen : SERVICE = gen
gen : WEIGHT = 0
gen : GEN = 1
gen : CPU = 20m
gen : DEPS = a

gen : export PARAMS = s/SERVICE/$(SERVICE)/g; s/WEIGHT/$(WEIGHT)/g; s/DEPS/$(DEPS)/g; s/CPU/$(CPU)/g; s/GEN/$(GEN)/g; s/LIMIT/$(LIMIT)/g
gen : setup
	echo $$PARAMS
	sed "$$PARAMS" deploy/template.yaml | ko apply -f -

hey : setup
	kubectl apply -f deploy/load/load.yaml

