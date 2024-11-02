get-proto:
	GOPRIVATE=github.com/hyperfyodor/ssosage_proto \
	 go get github.com/hyperfyodor/ssosage_proto@dev

run:
	go build
	./ssosage

test:
	go test . -count=1 -v