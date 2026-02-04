# cmd/shortener

В данной директории содержится код, который скомпилируется в бинарное приложение.

Рекомендуется помещать только код, необходимый для запуска приложения, но не бизнес-логику.

Название директории должно соответствовать названию приложения.

Директория `cmd/shortener` содержит:

- точку входа в приложение (функция `main`)
- инициализацию зависимостей (можно вынести в отдельный пакет `internal/app`)
- настройку и запуск HTTP-сервера (можно вынести в отдельный пакет `internal/router`)
- обработку сигналов завершения работы приложения

## RUN

export SERVER_ADDRESS=":8080"
export BASE_URL="http://127.0.0.1:8080/"
#export FILE_STORAGE_PATH="./tamers.json"
export DATABASE_DSN="postgres://admin:admin@192.168.0.20:5432/demo?sslmode=disable"
export AUDIT_FILE="./tamers_audit.json"
export AUDIT_URL="http://192.168.0.20:8088"
go run ./cmd/shortener/main.go

curl -v -X POST -H 'Content-Type: application/json' -d 'http://lib001.ru' http://127.0.0.1:8080
curl -v -X POST -c 'cookie.txt' -H 'Content-Type: application/json' -d 'http://lib005.ru' http://127.0.0.1:8080 # create cookie
curl -v -X POST -b 'cookie.txt' -H 'Content-Type: application/json' -d 'http://lib005.ru' http://127.0.0.1:8080 # use cookie

### Build
go build -o /tmp/shortener.bin ./cmd/shortener

### Load 
BASE_URL=http://127.0.0.1:8080
SHORT=w1BhF3b5
LONG=http://example.com
TMO=30s
COOKIE_FILE=/tmp/cookie_hey.txt
curl -v -X POST -c $COOKIE_FILE -H 'Content-Type: application/json' -d 'http://lib005.ru' http://127.0.0.1:8080
MAIN_COOKIE=$(tail -1 "$COOKIE_FILE")

hey -z $TMO -m POST -d ${LONG} ${BASE_URL} &
hey -z $TMO -m GET ${BASE_URL}/${SHORT} &
hey -z $TMO -m POST -d "{\"URL\":\"${LONG}\"}" -T "application/json" ${BASE_URL}/api/shorten &
hey -z $TMO -m GET ${BASE_URL}/ping &
hey -z $TMO -m POST -d "[{\"correlation_id\":\"${LONG}\", \"original_url\":\"${LONG}\"},]" -H "Cookie: $MAIN_COOKIE" ${BASE_URL}/api/shorten/batch &
hey -z $TMO -m GET -H "Cookie: $MAIN_COOKIE" ${BASE_URL}/api/user/urls &
hey -z $TMO -m DELETE -d "[{\"${SHORT}\"]" -H "Cookie: $MAIN_COOKIE" ${BASE_URL}/api/user/urls &

### Warming up
sleep 10

### Save memory profile 
#curl -sSf http://127.0.0.1:9090/debug/pprof/heap -o profiles/base.pprof
curl -sSf http://127.0.0.1:9090/debug/pprof/heap -o profiles/result.pprof

### Interactive
#go tool pprof /tmp/shortener.bin profiles/base.pprof
go tool pprof /tmp/shortener.bin profiles/result.pprof
#### top list peek
go tool pprof -top -diff_base=profiles/base.pprof profiles/result.pprof > profiles/README.md
