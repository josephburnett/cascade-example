goals = help all a b c ns clean
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
a : WEIGHT = 20
a : CPU = 20m
a : DEPS = b,c

b : SERVICE = b
b : WEIGHT = 20
b : CPU = 20m
b : DEPS = 

c : SERVICE = c
c : WEIGHT = 20
c : CPU = 20m
c : DEPS = 

a : export PARAMS = s/SERVICE/$(SERVICE)/g; s/WEIGHT/$(WEIGHT)/g; s/DEPS/$(DEPS)/g; s/CPU/$(CPU)/g
a : ns
	echo $$PARAMS
	sed "$$PARAMS" deploy/template.yaml | ko apply -f -

b : export PARAMS = s/SERVICE/$(SERVICE)/g; s/WEIGHT/$(WEIGHT)/g; s/DEPS/$(DEPS)/g; s/CPU/$(CPU)/g
b : ns
	echo $$PARAMS
	sed "$$PARAMS" deploy/template.yaml | ko apply -f -

c : export PARAMS = s/SERVICE/$(SERVICE)/g; s/WEIGHT/$(WEIGHT)/g; s/DEPS/$(DEPS)/g; s/CPU/$(CPU)/g
c : ns
	echo $$PARAMS
	sed "$$PARAMS" deploy/template.yaml | ko apply -f -

ns :
	kubectl create namespace cascade-example --dry-run=true -o yaml | kubectl apply -f -

clean :
	kubectl delete namespace cascade-example --ignore-not-found
