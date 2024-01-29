.PHONY: test
test:
	docker compose up -d --wait
	go test -v -race ./...
	docker compose down
