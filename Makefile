.PHONY: container publish serve serve-container clean

app        := daptin
static-app := build/linux-amd64/$(app)
docker-tag := daptin/daptin

bin/$(app): *.go
	go build -o $@

$(static-app): *.go
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
		go build -ldflags "-s" -a -installsuffix cgo -o $(static-app)

container: $(static-app)
	docker build -t $(docker-tag) .

publish: container
	docker push $(docker-tag)

serve: bin/$(app)
	env PATH=$(PATH):./bin forego start web

serve-container:
	docker run -it --rm --env-file=.env -p 8081:8080 $(docker-tag)

clean:
	rm -rf bin build