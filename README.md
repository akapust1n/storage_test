Программу можно доработать в плане отказоустойчивости, обработки потоков данных,более удачных протоколов, ошибок пользователя и других нюансов, но для ожидаемого эстимейта на выполнение задание "не более одного дня" - это ок, как мне кажется.

## Запуск всех сервисов

```bash
docker compose up --build -d
```
## Запуск тестов

```bash
go build -o server ./server
go test -v ./server```
```

## Добавление нового storage сервера

1. Собрать образ storage сервера :
```bash
docker build -t storage-image -f storage/Dockerfile .
```

2. Запустить новый storage сервер:
```bash
docker run -d -p 8088:8088 -e PORT=8088 storage-image
```

3. Зарегистрировать сервер через API:
```bash
curl -X POST "http://localhost:8080/register?url=http://localhost:8088"
```

Новый сервер сразу начнет использоваться для хранения чанков новых файлов.