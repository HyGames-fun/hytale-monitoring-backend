# hytale-monitoring

Бэкенд платформы мониторинга серверов Hytale на Go.

## Архитектура сервисов

Проект разбит на независимые микросервисы, объединённые общим API Gateway (Nginx):

- **API Gateway (`gateway`)**: Nginx на порту `4242`. Единая точка входа для клиентов. Маршрутизирует `/auth/*` и `/user/*` в сервис авторизации, а `/server/*` и `/` — в каталог.
- **Каталог (`cmd/catalog`)**: API серверов на Chi, поиск, фильтры, пагинация, лайки с проверкой Cloudflare Turnstile, лидерборд, подтверждение владения через DNS TXT и секретный токен агента.
- **Авторизация (`cmd/auth`)**: регистрация с подтверждением почты 6-значным кодом (SMTP), OAuth2 (Discord, Google), JWT access/refresh токены в HTTP-only cookies.
- **Сборщик метрик (`cmd/collector`)**: периодический опрос узлов через worker pool, кэширование лидерборда в Redis, gRPC сервер на порту `50051` для телеметрии из мода Hytale.

## Быстрый старт

### В Docker (микросервисы + Nginx Gateway)
```bash
docker compose up -d
```
После запуска все HTTP-эндпоинты доступны по единому адресу `http://localhost:4242`.

### Локальная разработка (монолитный режим)
Для быстрой разработки без поднятия отдельных контейнеров можно запустить единый рантайм:
```bash
# запуск базы и редиса
docker compose up -d postgres redis

# запуск всех сервисов в одном процессе
go run ./cmd/all
```

Либо запускать каждый бинарник отдельно:
```bash
go run ./cmd/catalog
go run ./cmd/auth
go run ./cmd/collector
```

## Переменные окружения

| Переменная | По умолчанию | Описание |
|---|---|---|
| `PORT` | `4242` | Порт API каталога |
| `AUTH_PORT` | `4243` | Порт сервиса авторизации |
| `GRPC_PORT` | `50051` | Порт gRPC для мода серверов |
| `DATABASE_URL` | `postgres://...` | Подключение к PostgreSQL |
| `REDIS_URL` | `redis://...` | Подключение к Redis |
| `JWT_SECRET` | `...` | Секретный ключ подписи JWT |
| `COOKIE_DOMAIN` | `localhost` | Домен для cookie |
| `DISCORD_CLIENT_ID` | `...` | Discord OAuth2 client id |
| `GOOGLE_CLIENT_ID` | `...` | Google OAuth2 client id |
