KIND_VERSION := 0.23.0
KUBERNETES_VERSION := 1.29.2
KUSTOMIZE_VERSION := 5.4.2


BINDIR := $(abspath $(PWD)/bin)

KIND := $(BINDIR)/kind
KUBECTL := $(BINDIR)/kubectl
KUSTOMIZE := $(BINDIR)/kustomize

PROJECT_NAME = sample-cni
KIND_CONFIG = kind.yaml

.PHONY: start
start: setup
	$(KIND) create cluster --image kindest/node:v$(KUBERNETES_VERSION) --config=$(KIND_CONFIG) --name $(PROJECT_NAME)

.PHONY: stop
stop:
	$(KIND) delete cluster --name $(PROJECT_NAME)


.PHONY: build
build:
	go build -o $(PROJECT_NAME) main.go

.PHONY: install
install: build
	docker cp $(PROJECT_NAME) $(PROJECT_NAME)-control-plane:/opt/cni/bin/sample-cni
	docker exec $(PROJECT_NAME)-control-plane ls /etc/cni/net.d/
	docker exec $(PROJECT_NAME)-control-plane rm /etc/cni/net.d/10-kindnet.conflist || true
	docker cp 10-$(PROJECT_NAME).conflist $(PROJECT_NAME)-control-plane:/etc/cni/net.d/10-$(PROJECT_NAME).conflist
	docker exec $(PROJECT_NAME)-control-plane ls /etc/cni/net.d/
	docker exec $(PROJECT_NAME)-control-plane ls /opt/cni/bin/

.PHONY: setup
setup: $(KIND) $(KUBECTL) $(KUSTOMIZE)

$(KIND):
	mkdir -p $(dir $@)
	curl -sfL -o $@ https://github.com/kubernetes-sigs/kind/releases/download/v$(KIND_VERSION)/kind-linux-arm64
	chmod a+x $@

$(KUBECTL):
	mkdir -p $(dir $@)
	curl -sfL -o $@ https://dl.k8s.io/release/v$(KUBERNETES_VERSION)/bin/linux/arm64/kubectl
	chmod a+x $@

$(KUSTOMIZE):
	mkdir -p $(dir $@)
	curl -sfL https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2Fv$(KUSTOMIZE_VERSION)/kustomize_v$(KUSTOMIZE_VERSION)_linux_arm64.tar.gz | tar -xz -C $(BINDIR)
	chmod a+x $@
