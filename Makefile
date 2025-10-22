BIN = cmd/gophermart/gophermart
PORT = 8000
DNNAME = praktikum
DSN = postgresql://domurdoc@localhost:5432/${DNNAME}?sslmode=disable
DIR = .
MAIN = cmd/gophermart/main.go
MNAME = unnamed
TESTBIN = gophermarttest-darwin-arm64
ACCRUALBIN = cmd/accrual/accrual_darwin_arm64
ACCRUAL_PORT = 8001

run:
	go run ${MAIN} -d ${DSN}

exe:
	./${BIN}

re:
	rm -f ${BIN}
	go build -o ${BIN} ${MAIN}

redb:
	dropdb ${DNNAME}
	createdb ${DNNAME}

accrual:
	./${ACCRUALBIN}

kill:
	killall -9 gophermart || true

mm:
	migrate create -ext sql -dir ./migrations -seq ${MNAME}

m:
	migrate -database "${DSN}" -path ./migrations up

md:
	migrate -database "${DSN}" -path ./migrations down 1

test: kill re
	./${TESTBIN} -test.v \
	-gophermart-binary-path=${BIN} \
	-gophermart-host=localhost \
	-gophermart-port=${PORT} \
	-gophermart-database-uri="${DSN}" \
	-accrual-binary-path=${ACCRUALBIN} \
	-accrual-host=localhost \
	-accrual-port=${ACCRUAL_PORT} \
	-accrual-database-uri="${DSN}"


PHONY: run exe re redb accrual kill m mm md test
