CURDIR=$(shell pwd)
BINDIR=${CURDIR}/bin
GOVER=$(shell go version | perl -nle '/(go\d\S+)/; print $$1;')
MOCKGEN=${BINDIR}/mockgen_${GOVER}
SMARTIMPORTS=${BINDIR}/smartimports_${GOVER}
LINTVER=v1.49.0
LINTBIN=${BINDIR}/lint_${GOVER}_${LINTVER}
PACKAGE=gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/cmd/bot

all: format build test lint

build: bindir
	go build -o ${BINDIR}/bot ${PACKAGE}
	go build -o ${BINDIR}/goose ./cmd/goose

build-goose: bindir
	go build -o ${BINDIR}/goose ./cmd/goose

build-seeder: bindir
	go build -o ${BINDIR}/seeder ./cmd/seeder

test:
	go test ./...

migration-new:
	${BINDIR}/goose create ${name} go

migration-up: build-goose
	${BINDIR}/goose up

migration-down: build-goose
	${BINDIR}/goose down

migration-status: build-goose
	${BINDIR}/goose status

run:
	go run ${PACKAGE}

generate: install-mockgen
	${MOCKGEN} -source=internal/model/messages/incoming_msg.go -destination=internal/mocks/messages/messages_mocks.go

lint: install-lint
	${LINTBIN} run

precommit: format build test lint
	echo "OK"

seed: build-seeder
	${BINDIR}/seeder

bindir:
	mkdir -p ${BINDIR}

format: install-smartimports
	${SMARTIMPORTS} -exclude internal/mocks

install-mockgen: bindir
	test -f ${MOCKGEN} || \
		(GOBIN=${BINDIR} go install github.com/golang/mock/mockgen@v1.6.0 && \
		mv ${BINDIR}/mockgen ${MOCKGEN})

install-lint: bindir
	test -f ${LINTBIN} || \
		(GOBIN=${BINDIR} go install github.com/golangci/golangci-lint/cmd/golangci-lint@${LINTVER} && \
		mv ${BINDIR}/golangci-lint ${LINTBIN})

install-smartimports: bindir
	test -f ${SMARTIMPORTS} || \
		(GOBIN=${BINDIR} go install github.com/pav5000/smartimports/cmd/smartimports@latest && \
		mv ${BINDIR}/smartimports ${SMARTIMPORTS})

docker-run:
	sudo docker compose up
