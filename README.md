# Trip Service

Учебный сервис для лабораторной работы 1: HTTP API поверх PostgreSQL.

## Требования

- Go 1.24+
- Docker (для локального окружения через tripgoctl)
- tripgoctl (см. репозиторий course-infra)

## Запуск

1. Скопировать `environment.toml` в корень репозитория (уже сделано).
2. Поднять окружение:
   tripgoctl cluster start
   tripgoctl environment start
   tripgoctl connect
3. Загрузить переменные окружения в текущую сессию терминала:
   export $(grep -v '^#' .env | xargs)
4. Накатить миграции:
   make migrate
5. Запустить сервис:
   make run

Сервис слушает адрес из `HTTP_ADDR` (по умолчанию `:8080`).

## Переменные окружения

| Переменная | Назначение |
|---|---|
| `HTTP_ADDR` | адрес, на котором слушает HTTP-сервер |
| `DATABASE_URL` | строка подключения к PostgreSQL |
| `LOG_LEVEL` | уровень логирования |
| `SHUTDOWN_TIMEOUT` | таймаут graceful shutdown |
| `DATABASE_MAX_CONNS`, `DATABASE_MIN_CONNS` | размер пула соединений |
| `DATABASE_CONNECT_TIMEOUT`, `DATABASE_QUERY_TIMEOUT` | таймауты на подключение и запросы |
| `DATABASE_MAX_CONN_LIFETIME` | время жизни соединения в пуле |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | адрес OTLP-коллектора, куда сервис отправляет трейсы и метрики |
| `OTEL_EXPORTER_OTLP_PROTOCOL` | протокол экспорта в коллектор |
| `OTEL_TRACES_SAMPLER` | стратегия сэмплирования трейсов |

## Команды

- `make migrate` — накатить миграции
- `make migrate-down` — откатить миграции
- `make run` — запустить сервис локально
- `make generate` — сгенерировать код по OpenAPI-контракту
- `make build` — собрать бинарник
- `make test` — прогнать тесты с `-race`

## Решения

### Уровень изоляции транзакций

Выбран `Read Committed` — уровень изоляции по умолчанию в PostgreSQL. Он подходит, потому что невозможность двух активных поездок у водителя обеспечивается частитным уникальным индексом. 

### Менеджер транзакций

Реализован в `internal/db/tx_manager.go`. Интерфейс `TxManager.Do(ctx, fn)`:
- открывает транзакцию через `pgxpool.BeginTx`;
- кладёт её в `context.Context` через приватный ключ `txKey{}`, чтобы репозитории могли достать транзакцию из контекста, не получая её отдельным параметром;
- если `fn` вернула `nil` — коммитит, если ошибку или паникует — откатывает;
- вложенный вызов `Do` внутри уже открытой транзакции переиспользует её же, не открывая вторую (проверяется наличием транзакции в контексте перед началом новой).

Репозитории получают исполнителя запросов (`Executor`) через `ExecutorFromContext`: если в контексте есть транзакция — используют её, иначе — пул соединений напрямую.

### Запрет двух активных поездок у одного водителя

Обеспечен частичным уникальным индексом:
sql
```
CREATE UNIQUE INDEX trips_driver_active_uidx
    ON trips (driver_id)
    WHERE status = 'active';
```
Уникальность проверяется только среди строк со статусом active, поэтому у водителя может быть много завершённых поездок, но не больше одной активной 
одновременно. Проверка и вставка строки происходят в PostgreSQL как одно неразделимое действие, поэтому при двух одновременных запросах база сама 
гарантированно пропустит только один INSERT, а второй получит ошибку. Это отличается от проверки SELECT перед INSERT в коде: между этими двумя отдельными 
запросами есть промежуток времени, в который может успеть вклиниться второй параллельный запрос.

### Защита от гонки при завершении поездки

`UPDATE trips SET status = 'completed', ... WHERE id = ? AND status = 'active'` — обновление срабатывает только если поездка ещё активна. Если строка не была обновлена (`RowsAffected() == 0`), выполняется дополнительное чтение, чтобы различить два случая: поездки нет вообще (`404 trip_not_found`) или она уже завершена (`409 trip_completed`).

## Архитектура

### Слои и зависимости

| Слой | Пакет | Ответственность |
|---|---|---|
| Домен | `internal/domain/trip` | модель `Trip`, переходы статусов (`Finish`), валидация, доменные ошибки, `ValidationError` |
| Usecase | `internal/usecase/trip/{create,get,finish}` | сценарий: транзакция, вызов репозиториев, метрики и бизнес-логи |
| Транспорт | `internal/http/trip/{create,get,finish}`, `server`, `response`, `middleware` | разбор запроса, вызов usecase, формирование ответа |
| Маппинг ошибок | `internal/http/domainerror` | единственное место, где доменная ошибка превращается в `Problem` |
| Репозиторий | `internal/repository/trip` | SQL через squirrel, реализует интерфейсы usecase |
| Инфраструктура | `internal/infra/postgres`, `postgres/transaction`, `telemetry` | пул pgx, менеджер транзакций, OpenTelemetry и slog |
| Сборка | `internal/container` | единственное место, где создаются пул и репозитории |


Зависимости направлены внутрь: транспорт и репозитории знают про usecase, usecase знает про домен, домен не знает ни про кого.

```
cmd/service (main)
      │
internal/container ── собирает все зависимости вручную
      │
internal/http/*  ──>  internal/usecase/*  ──>  internal/domain/trip
                              ^
internal/repository/trip ─────|  (реализует интерфейсы usecase)
internal/infra/postgres, transaction, telemetry
```
### Интерфейсы

Интерфейсы лежат в `contract.go` пакета-потребителя и неэкспортируемы (`tripRepository`, `historyRepository`, `txManager`, `metrics`). Потребитель сам описывает, что ему нужно, поэтому интерфейсы узкие: `get` знает только `GetByID`, `create` только `Create`. Реализаций может быть несколько (Postgres, мок в тестах), а потребности usecase остаются прежними. Моки генерирует `mockgen` через `go:generate` в `mocks_test.go` рядом с кодом.

Второй транспорт сможет вызывать те же usecase: нужно будет добавить пакет в `internal/consumer`, а usecase, домен и репозитории не изменятся.


### Ошибки

Домен возвращает типизированные ошибки (`ErrTripNotFound`, `ErrTripCompleted`, `ErrDriverBusy`, `*ValidationError`) и ничего не знает про HTTP. Хендлеры передают ошибку из usecase в `response.Error`, который вызывает `domainerror.Problem(err)`. Там единственный `switch` по `errors.Is` / `errors.As`. Неизвестные ошибки дают `500 internal_error` без деталей в теле, подробности попадают только в лог.

### Observability

- **Логи:** `log/slog`, JSON, stdout, логгер передаётся зависимостью. В записи добавляются `trace_id` и `span_id`. Middleware пишет одну запись `request handled` на запрос: `Info` для 2xx и 4xx, `Error` для 5xx.
- **Трейсы:** OTLP-экспорт, корневой спан `{method} {route}`, спаны БД создаёт `otelpgx` на уровне пула, репозитории про телеметрию не знают.
- **Метрики:** `http.server.request.duration`, `http.server.active_requests`, `db.client.operation.duration`, `tripgo.trips.created`, `tripgo.trips.completed`.
- **Middleware:** `Recover -> Tracing -> Logging -> Metrics -> роутер`.
- В `http.route` попадает только шаблон (`/api/v1/trips/{tripId}`). Реальный id дал бы неограниченное число значений лейбла (высокая кардинальность) и раздул бы хранилище метрик.
- Недоступность коллектора не влияет на работу сервиса: экспорт идёт в фоне.

## Тесты и линтер

| Команда | Что делает |
|---|---|
| `make test` | unit-тесты с `-race` и покрытием |
| `make test-integration` | интеграционные тесты на реальном PostgreSQL |
| `make lint` | `golangci-lint` с `depguard` и остальными линтерами |
| `make fmt` | форматтеры `gofumpt` и `gci` |
| `go generate ./...` | перегенерация моков |

Для `make test-integration` нужен запущенный PostgreSQL из `tripgoctl environment start` и переменная `DATABASE_URL` из `.env`:

```bash
export $(grep -v '^#' .env | xargs)
make test-integration
```

Каждый интеграционный тест создаёт свои данные с уникальными id и удаляет их после себя, поэтому тесты можно запускать повторно.