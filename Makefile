cleanupt:
	./scripts/cleanup_temp.sh

testall:
	GOGACHE=off GO_ENV=test go test -v -bench=. -benchmem ./pkg/...

test:
	GOGACHE=off GO_ENV=test go test -v -bench=^$ -benchmem ./pkg/...

dev:
	DEBUG=1 air

build-static:
	rm -rf ./dist && ./node_modules/.bin/gulp
