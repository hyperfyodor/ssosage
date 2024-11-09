get-proto:
	 go get github.com/hyperfyodor/ssosage_proto

migrate:
	go run ./cmd/migrator --config=./config/migrations.json

run:
	go run ./cmd/ssosage --config=./config/ssosage.json

test:
	go test ./tests -count=1 -v