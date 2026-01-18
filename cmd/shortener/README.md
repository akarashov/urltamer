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

