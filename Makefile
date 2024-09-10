build:
	go build -o ./shortener cmd/shortener/*.go
re_build:
	rm ./shortener && go build -o ./shortener cmd/shortener/*.go
run:
	go run  $(find cmd/shortener -maxdepth 1 -name "*.go" ! -name "*_test.go")
delete:
	rm ./shortener
