build:
	go build -o ./shortener cmd/shortener/*.go
re_build:
	rm ./shortener && go build -o ./shortener cmd/shortener/*.go
run:
	go run  $(find cmd/shortener -maxdepth 1 -name "*.go" ! -name "*_test.go")
delete:
	rm ./shortener
test:
	go test -v ./handlers/
init_db:
	docker run --name shortener-postgres \
	-e POSTGRES_USER=${POSTGRES_USER} \
	-e POSTGRES_PASSWORD=${POSTGRES_PASSWORD} \
	-e POSTGRES_DB=${POSTGRES_DB} \
	-p ${POSTGRES_PORT}:5432 \
	-v ${POSTGRES_DATA}:/var/lib/postgresql/data \
	-d postgres:16
