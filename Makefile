REGISTRY = norbega
PROJECT = eraser
VERSION = v1beta2
GIT_URL = github.com/igarridot
BASE_REPO = learning-k8s-crds
KIND = Environment

install-kubebuilder:
	curl -L -o kubebuilder https://go.kubebuilder.io/dl/latest/$$(go env GOOS)/$$(go env GOARCH)
	chmod +x kubebuilder && mv kubebuilder /usr/local/bin/
	kubebuilder completion zsh >> ~/.zshrc
kubebuilder-init-project:
	mkdir -p ${PROJECT}
	cd ${PROJECT} && kubebuilder init --domain ${BASE_REPO} --repo ${GIT_URL}/${BASE_REPO}/${PROJECT}
kubebuilder-create-api:
	cd ${PROJECT} && kubebuilder create api --group learning-k8s-crds --version ${VERSION} --kind ${KIND} --controller --resource
apply-cr-manifests:
	kubectl apply -f ${PROJECT}/config/samples/
apply-crd-components:
	cd ${PROJECT} && make deploy IMG=${REGISTRY}/${PROJECT}:${VERSION}
remove-crd-components:
	cd ${PROJECT} && make undeploy


init: kubebuilder-init-project kubebuilder-create-api
apply: apply-crd-components apply-cr-manifests
remove: remove-crd-components

build:
	cd ${PROJECT} && make docker-build IMG=${REGISTRY}/${PROJECT}:${VERSION}
push:
	cd ${PROJECT} && make docker-push IMG=${REGISTRY}/${PROJECT}:${VERSION}

