# pr-reviewer-assignment
Сервис для назначения ревьюверов для Pull Request'ов. 

Сделано согласно [заданию](docs\Backend-trainee-assignment-autumn-2025.md). API можно посмотреть [здесь](api\openapi.yml).

## Инструкция по запуску
Есть [Makefile](Makefile) с разными командами. Поднять докер можно написав:
```
make up
```
или
```
docker-compose up
```
Перед этим нужно написать нужные переменные окружения в `.env` файле, [пример файла](example.env).