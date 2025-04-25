# Word of Wisdom TCP Server

## Описание проекта

TCP-сервер "Word of Wisdom" — это сервис на языке Go, который предоставляет клиентам уникальные  цитаты через TCP-соединение. Сервис включает механизм Proof of Work (PoW) для защиты от DDoS-атак, гарантируя доступ к цитатам только для легитимных клиентов. Сервер генерирует уникальные цитаты, комбинируя слова, и кэширует их, чтобы избежать повторений.

### Основные возможности

- **TCP-сервер**: Слушает и обрабатывает подключения клиентов.
- **Защита от DDoS**:
  - Предварительное PoW (сложность=3, таймаут 0.5 сек) 
  - Основное PoW (сложность=4 или 6, таймаут 2 сек) обеспечивает ресурсозатратную проверку.
  - Ограничение в 20 соединений с одного IP, очередь на 2000 слотов.
  - Подозрительные IP (после 5 неудачных попыток) временно блокируются.
- **Уникальные цитаты**:
  - Генерирует цитаты в формате `прилагательное существительное глагол #счётчик — автор`.
  - Поддерживает \~100,000,000 уникальных комбинаций с низкой вероятностью повторов.
  - Цитаты кэшируются в памяти (`sync.Map`) для обеспечения уникальности.
- **Контейнеризация**:
  - Сервер и клиент упакованы в Docker-контейнеры.
  - Управление через Docker Compose для упрощённого запуска.

## Структура проекта

```
word-of-wisdom-project/
├── docker-compose.yml          # Docker Compose
├── server/                     # Код сервера
│   ├── Dockerfile              # Dockerfile для сервера
│   ├── go.mod                  # Go mod
│   ├── main.go                 # Точка входа сервера
│   ├── domain/                 # Модели данных (Challenge, Quote)
│   ├── usecase/                # Бизнес-логика (PoW, генерация цитат)
│   ├── repository/             # Генерация и хранение цитат
│   └── delivery/tcp/           # Реализация TCP-сервера
├── client/                     # Код клиента
│   ├── Dockerfile              # Dockerfile для клиента
│   ├── go.mod                  # Go mod
│   └── main.go                 # Реализация клиента
├── README.md                   # Документация
```

## Установка и запуск

### 1. Клонирование репозитория

```bash
git clone <repository-url>
cd word-of-wisdom-project
```

### 2. Проверка работы Docker и доступности порта

- 

  ```bash
  docker --version
  docker-compose --version
  ```

Сервер использует порт 8080. он должен быть свободен:

```bash
lsof -i :8080
```

### 4. Сборка и запуск через Docker Compose

Соберите и запустите сервер и клиент:

```bash
docker-compose up --build
```

- Команда:
  - Соберёт образы `wisdom-server` и `wisdom-client`.
  - Запустит сервер (на порту 8080) и один клиент.
  - Клиент подключится, решит PoW-задачи и получит цитату.

**Ожидаемый вывод**:

```
Starting word-of-wisdom-project_server_1 ... done
Starting word-of-wisdom-project_client_1 ... done
server-1  | 2025/04/25 02:00:00 TCP server running on :8080
server-1  | Connection from 172.18.0.3:12345
client-1  | Attempting to connect to server at server:8080
client-1  | Received preliminary challenge: PRE:abc123:xyz789:3
client-1  | Sent preliminary nonce: 1234
client-1  | Received main challenge: def456:uvw012:4
client-1  | Sent main nonce: 5678
client-1  | Server response: adj1 noun2 verb3 #1 — author4
word-of-wisdom-project_client_1 exited with code 0
```

### 5. Проверка логов

- Логи сервера:

  ```bash
  docker logs word-of-wisdom-project_server_1
  ```

  Ожидаемый вывод:

  ```
  2025/04/25 02:00:00 TCP server running on :8080
  Connection from 172.18.0.3:12345
  ```

- Логи клиента:

  ```bash
  docker logs word-of-wisdom-project_client_1
  ```

  Ожидаемый вывод:

  ```
  Attempting to connect to server at server:8080
  Received preliminary challenge: PRE:abc123:xyz789:3
  Sent preliminary nonce: 1234
  Received main challenge: def456:uvw012:4
  Sent main nonce: 5678
  Server response: adj1 noun2 verb3 #1 — author4
  ```

### 7. Остановка сервиса

Чтобы остановить контейнеры:

```bash
docker-compose down
```

## Тестирование защиты от DDoS

Сервер защищён от DDoS-атак с помощью PoW и ограничений на соединения. Для симуляции нагрузки:

```bash
for i in {1..1000}; do
  docker run --network word-of-wisdom-project_wisdom-net -e SERVER_ADDR=server:8080 --rm wisdom-client &
done
```

- **Ожидаемое поведение**:

  - Сервер отфильтрует \~90% "злоумышленников" через предварительное PoW.
  - Легитимные клиенты (решающие PoW в рамках таймаутов) получат цитаты.
  - Подозрительные IP (с 5+ неудачными попытками) блокируются на 30 секунд.

## Дополнительные рекомендации

- **Масштабирование**:

  - Протестируйте с 10,000 клиентов для оценки защиты от DDoS при высокой нагрузке:

    ```bash
    for i in {1..10000}; do
      docker run --network word-of-wisdom-project_wisdom-net -e SERVER_ADDR=server:8080 --rm wisdom-client &
    done
    ```