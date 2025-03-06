BIN=go

build:
	${BIN} build -v ./...

test:
	gotest -race ./... -v -race

watch-test:
	wgo gotest ./... -v -race

bench:
	go test -benchmem -count 3 -bench ./...

profile:
	rm memprofile.out
	go test -bench=. -benchmem -memprofile memprofile.out
	# go tool pprof -http :8081 memprofile.out
	# go tool pprof memprofile.out

coverage:
	${BIN} test -v -coverprofile=cover.out -covermode=atomic .
	${BIN} tool cover -html=cover.out -o cover.html

tools:
	${BIN} install github.com/bokwoon95/wgo@latest
	${BIN} install github.com/rakyll/gotest@latest
	${BIN} install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	${BIN} get -t -u golang.org/x/tools/cmd/cover
	go mod tidy

lint:
	golangci-lint run --timeout 60s --max-same-issues 50 ./...
lint-fix:
	golangci-lint run --timeout 60s --max-same-issues 50 --fix ./...
