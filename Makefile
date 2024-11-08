get-proto:
	GOPRIVATE=github.com/hyperfyodor/ssosage_proto \
	 go get github.com/hyperfyodor/ssosage_proto@apps_roles_model

migrate:
	go run ./cmd/migrator --config=./config/migrations.json

run:
	go run ./cmd/ssosage --config=./config/ssosage.json

test:
	go test ./tests -count=1 -v