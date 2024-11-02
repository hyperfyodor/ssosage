get-proto:
	GOPRIVATE=github.com/hyperfyodor/ssosage_proto \
	 go get github.com/hyperfyodor/ssosage_proto@dev

migrate:
	go run ./cmd/migrator --config=./config/migrations.json