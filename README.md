# DisturbedCLI (WIP) | Grep working, Other in future
**Важная ремарка, проверять задание можно, но планирую после курса доработать решение.**

Сборник распределенных CLI-утилит (grep, cut, sort) написанный на Go. 

* **Grep.** Принимает паттерн и файлы (или stdin), делит данные на чанки, обрабатывает параллельно через горутины или сеть HTTP-воркеров и возвращает результат при достижении кворума. Вывод совместим со стандартным grep.

## Содержание

- [Возможности](#возможности)
- [Архитектура](#архитектура)
- [Быстрый старт](#быстрый-старт)
- [Флаги](#флаги)
- [Примеры использования](#примеры-использования)
- [Кворум](#кворум)
- [Сравнительный тест](#сравнительный-тест)
- [Тесты](#тесты)
- [Структура проекта](#структура-проекта)

## Возможности

- **Локальный параллелизм** - `-workers N` делит stdin/файл на `N` чанков, каждый обрабатывается в отдельной горутине
- **Распределённый режим** - `-peers :5001,:5002,:5003` fan-out чанки на HTTP-воркеры, запущенные на любых хостах
- **Кворум** - результат принимается только если ≥ `-quorum` воркеров успешно ответили; при недоборе - ошибка
- **Совместимые флаги** - `-i`, `-v`, `-n`, `-c`, `-o`, `-q` работают как в GNU grep
- **Горутины + каналы** - параллелизм без мьютексов, результаты собираются через `chan worker.Result`
- **Корректный exit code** - `0` при совпадении, `1` при отсутствии, `2` при ошибке

## Архитектура

```
stdin / file
      │
      ▼
  Splitter - делит []string на N равных чанков с сохранением номеров строк
      │
      ├─── [local mode]  ──────────────────────────────────────────────────────
      │    goroutine 1 → worker.Run(chunk1, re) ─┐
      │    goroutine 2 → worker.Run(chunk2, re) ─┤
      │    goroutine N → worker.Run(chunkN, re) ─┤
      │                                          ▼
      │                                    chan Result
      │                                          │
      │                                        Quorum.Collect()
      │
      └─── [distributed mode]  ────────────────────────────────────────────────
      │      goroutine 1 → POST :5001/grep ─┐
      │      goroutine 2 → POST :5002/grep ─┤
      │      goroutine 3 → POST :5003/grep ─┤
      │                                     ▼
      │                               chan Result
      │                                     │
      │                                   Quorum.Collect()
      │
      ▼
  Merger    - сортирует совпадения по оригинальному номеру строки
      │
      ▼
  Formatter - выводит с учётом -n, -c, -o, -q, имени файла
```

Каждый воркер-узел (`-addr :5001`) - отдельный HTTP-сервер, принимающий `POST /grep` с JSON-телом и возвращающий список совпадений.

## Быстрый старт

### Требования

- Go 1.26.1+

### Сборка

```bash
git clone https://github.com/RidusM/DisturbedCLI
cd disturbedcli
make build
# бинарь: ./grep
```

### Запуск локально

```bash
# просто как grep
echo "hello world\nerror here" | ./grep "error"

# 4 горутины параллельно
cat large.log | ./grep -workers 4 "panic"
```

### Запуск в распределённом режиме

```bash
# Терминал 1, 2, 3 - запустить воркеры
./grep -addr :5001
./grep -addr :5002
./grep -addr :5003

# Терминал 4 - координатор
cat app.log | ./grep -peers :5001,:5002,:5003 -quorum 2 "error"
```

## Флаги

| Флаг | Описание |
|---|---|
| `-i` | Игнорировать регистр |
| `-v` | Инвертировать совпадение (вывести несовпадающие строки) |
| `-n` | Выводить номер строки перед каждым совпадением |
| `-c` | Вывести только количество совпадающих строк |
| `-o` | Вывести только совпавшую часть строки |
| `-q` | Тихий режим: не выводить ничего, exit 0/1 |
| `-workers N` | Число параллельных горутин в локальном режиме (по умолчанию: 1) |
| `-quorum N` | Минимум успешных воркеров для принятия результата (по умолчанию: workers/2+1) |
| `-addr :PORT` | Запустить как HTTP-воркер на указанном адресе |
| `-peers a,b,c` | Адреса воркеров для координатора (включает распределённый режим) |

## Примеры использования

### Базовый поиск

```bash
# поиск в stdin
echo -e "ok\nerror found\nall good" | ./grep "error"
# error found

# поиск в файле
./grep "panic" app.log

# несколько файлов - выводится имя файла
./grep "error" app.log access.log
```

### Флаги совместимые с grep

```bash
# без учёта регистра
./grep -i "error" app.log

# инвертировать: строки без "ok"
cat app.log | ./grep -v "ok"

# с номерами строк
./grep -n "error" app.log
# 42:error: connection refused
# 87:error: timeout

# только количество совпадений
./grep -c "error" app.log
# 17

# только совпавшая часть (-o)
echo "fatal error: disk full" | ./grep -o "error: [a-z ]+"
# error: disk full

# тихий режим (для скриптов)
./grep -q "panic" app.log && echo "found panics!"
```

### Локальный параллельный режим

```bash
# 8 горутин, кворум по умолчанию (5/8)
cat 10gb.log | ./grep -workers 8 "OutOfMemory"

# явный кворум: достаточно 3 из 8
cat big.log | ./grep -workers 8 -quorum 3 "error"
```

### Распределённый режим

```bash
# запустить воркеры (на разных машинах или портах)
./grep -addr :5001 &
./grep -addr :5002 &
./grep -addr :5003 &

# поиск с кворумом 2 из 3
cat app.log | ./grep -peers :5001,:5002,:5003 -quorum 2 "error"

# с удалёнными хостами
./grep -peers 10.0.0.1:5001,10.0.0.2:5001,10.0.0.3:5001 -quorum 2 "panic" app.log
```

### Health check воркера

```bash
curl http://localhost:5001/health
# ok
```

## Кворум

Кворум - минимальное число воркеров, которые должны успешно вернуть результат. Если меньше - команда завершается с ошибкой, результат не выводится.

```
3 воркера, quorum=2:

worker1 → OK  ✓
worker2 → OK  ✓   → кворум достигнут → вывод результата
worker3 → fail ✗

3 воркера, quorum=3:

worker1 → OK  ✓
worker2 → fail ✗  → кворум НЕ достигнут → exit 2
worker3 → fail ✗
```

По умолчанию кворум = `workers/2 + 1` (строгое большинство).

## Сравнительный тест

500 000 строк, ~10% содержат паттерн `ERROR`.

```
goos: windows
goarch: amd64
cpu: Intel(R) Core(TM) i5-4570T CPU @ 2.90GHz

BenchmarkLocal_1Worker-4             100          50572690 ns/op        13805915 B/op      50394 allocs/op
BenchmarkLocal_4Workers-4              6        1039424417 ns/op        15705541 B/op      50460 allocs/op
BenchmarkLocal_8Workers-4              6        1484240367 ns/op        16903466 B/op      50526 allocs/op
BenchmarkSystemGrep-4                232          91463766 ns/op           45600 B/op        320 allocs/op
```

**Важно!** При `-workers > 1` накладные расходы на запуск горутин и конкуренцию за append к результатам заметны на таком объёме данных. В реальном сценарии выигрыш проявляется при CPU-тяжёлых regexp или при I/O-параллелизме (чтение нескольких файлов).  Системный `grep` использует C-реализацию Boyer-Moore с буферизацией на уровне ОС - прямое сравнение показывает порядок отставания Go regexp.

Запустить самостоятельно:

```bash
make bench
```

## Тесты

```bash
# unit-тесты
make test

# с -race detector
go test ./internal/... -race -v
```

Покрытые сценарии:

| Пакет | Тесты |
|---|---|
| `splitter` | Равное деление, нечётное, больше воркеров чем строк, пустой ввод |
| `worker` | Базовый матч, `-i`, `-v`, `-o`, смещение StartLine |
| `quorum` | Все успешны, кворум достигнут, кворум не достигнут |

## Структура проекта

```
DisturbedCLI/
├── cmd/
│   └── GrepRequest/
│       └── main.go              # Точка входа: разбор флагов, выбор режима
├── internal/
│   ├── cli/
│   │   └── flags.go             # Парсинг и валидация флагов командной строки
│   ├── coordinator/
│   │   ├── coordinator.go       # Распределённый режим: fan-out на peers
│   │   ├── proto.go             # JSON-типы запроса/ответа (GrepRequest/Response)
│   │   ├── local.go             # Локальный режим: fan-out на горутины
│   │   └── worker_node.go       # HTTP-сервер воркера: POST /grep
│   └── merger/
│   │    ├── merger.go           # Форматирование вывода (-n, -c, -o, -q)
│   │    └── reader.go           # Чтение stdin и файлов
│   ├── quorum/
│   │    ├── quorum_test.go      # Тест quorum_test.go
│   │    └── quorum.go           # Сборка результатов с проверкой кворума
│   ├── splitter/
│   │   ├── splitter_test.go     # Тест splitter_test.go
│   │   └── splitter.go          # Деление []string на N чанков с StartLine
│   └── worker/
│   │   ├── worker_test.go.go    # Тест worker_test.go
│       └── worker.go            #  Grep-логика: Run(), CompilePattern(), Match
├── benchmark/
│   └── benchmark_test.go        # Сравнение 1/4/8 горутин vs system grep
├── Makefile
└── go.mod
```