goals = help all a b c d metrics setup bench nobench dash shell logs clean
.DEFAULT_GOAL : help
.PHONY : $(goals)
.ONESHELL : $(goals)

help :
	@echo "Usage: make all"
	@echo "Prerequisites:"
	@echo " - kubectl installed and pointing to a Kubernetes cluster"
	@echo " - ko installed and pointing to a repo (https://github.com/google/ko)"

all : metrics a b c d

a : SERVICE = a
a : WEIGHT = 20ms
a : GEN = 0
a : LIMIT = 5
a : CPU = 200m
a : DEPS = b

b : SERVICE = b
b : WEIGHT = 20ms
b : GEN = 0
b : LIMIT = 5
b : CPU = 200m
b : DEPS = c

c : SERVICE = c
c : WEIGHT = 20ms
c : GEN = 0
c : LIMIT = 5
c : CPU = 200m
c : DEPS = d

d : SERVICE = d
d : WEIGHT = 20ms
d : GEN = 0
d : LIMIT = 5
d : CPU = 200m
d : DEPS = 

bench : setup
	for i in `seq 1 2`; do kubectl -n cascade-example-bench run ab$$i --image=gcr.io/josephburnett-gke-dev/ab --restart=Always -- -s 120 -t 3600 -n 999999 -c 10 "http://a.cascade-example.svc.cluster.local/" ; done







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

d : export PARAMS = s/SERVICE/$(SERVICE)/g; s/WEIGHT/$(WEIGHT)/g; s/DEPS/$(DEPS)/g; s/CPU/$(CPU)/g; s/GEN/$(GEN)/g; s/LIMIT/$(LIMIT)/g
d : setup
	echo $$PARAMS
	sed "$$PARAMS" deploy/template.yaml | ko apply -f -

metrics: setup
	ko apply -f deploy/metrics.yaml

nobench :
	kubectl delete namespace cascade-example-bench --ignore-not-found

setup :
	kubectl create namespace cascade-example --dry-run=true -o yaml | kubectl apply -f -
	kubectl create namespace cascade-example-bench --dry-run=true -o yaml | kubectl apply -f -

dash :
	watch 'kubectl get svc -n cascade-example; echo; kubectl get hpa -n cascade-example; echo; kubectl get pods -n cascade-example-bench'

#  while true; do wget -q -O- http://metrics.cascade-example.svc.cluster.local/metrics | grep total_qps; sleep 2; done
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

