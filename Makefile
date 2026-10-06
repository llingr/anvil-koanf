FUZZTIME ?= 30s

default: build test

all: build test lint fuzz mutation

build:
	go build ./...
	go vet ./...

test:
	go test -race -coverpkg=./conf/... -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

# same image and version as the CI lint job, nothing installed on the host
lint:
	docker run --rm -v "$(CURDIR)":/app -w /app golangci/golangci-lint:v2.12.2 golangci-lint run ./...

# each fuzz target runs on its own for FUZZTIME; FUZZTIME=5m make fuzz to lengthen the hunt
fuzz:
	@for target in $$(go test -list='Fuzz.*' ./conf/tests | grep '^Fuzz'); do \
		echo "=== $$target"; \
		go test -run=NONE -fuzz=$$target -fuzztime=$(FUZZTIME) ./conf/tests || exit 1; \
	done

# integration runs the conf/tests package too, which coverpkg counts against conf
mutation:
	gremlins unleash --integration --coverpkg=./conf/... --workers 2 \
		--invert-assignments --invert-bitwise --invert-bwassign --invert-logical \
		--invert-loopctrl --remove-self-assignments --invert-negatives \
		--threshold-efficacy 95 --threshold-mcover 90

.PHONY: default all build test lint fuzz mutation
