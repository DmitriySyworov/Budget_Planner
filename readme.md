# 💰 GO Budget Planner

<details>
<summary><b>🇷🇺 Нажмите сюда, чтобы читать на русском языке (RU)</b></summary>

## 📊 О проекте
**GO Budget Planner** — это высоконагруженная микросервисная экосистема, предназначенная для оперативного мониторинга расходов, учета финансовых транзакций по категориям и расчета процентной аналитики.

### 🚀 Ключевые фичи:
* **Мониторинг транзакций:** Регистрация расходов, распределение по категориям и расчет аналитических срезов в процентах по дням.
* **Асинхронные нотификации:** Мгновенная отправка email-уведомлений при изменении баланса счета. Логика реализована через паттерн Transactional Outbox (Брокер очередей Redis `List` + пакетный `LMPOP` ➡️ Apache Kafka).
* **Агрегированный Swagger-шлюз:** Централизованная точка сборки документации API, развернутая внутри кластера.

---

## 🛠️ Развертывание и универсальный реплейс инфраструктуры (Makefile)

Полносистемный запуск экосистемы возможен **исключительно внутри контура Kubernetes (Minikube)** из-за жесткой связанности компонентов с инфраструктурным слоем (кластер Apache Kafka с SCRAM-авторизацией, инстансы Redis и базы данных PostgreSQL).

Конфигурация `values.yaml` содержит флаг безопасности **`isLocal`**. При значении `true` манифесты требуют принудительного монтирования локальных директорий хоста в ноду Minikube. Это гарантирует персистентность данных PostgreSQL и Redis при полном пересоздании кластера.

### 1. Локальный запуск Минкуба с монтированием томов

```bash
# Старт кластера с привязкой локальной папки к точке монтирования /mnt/data
make minikube-start-local
```
⚠️ **ВАЖНО:** Перед запуском этой команды обязательно откройте `Makefile` и **замените значение переменной `LOCAL_PATH_VOLUMES` на реальный абсолютный путь к папке на вашем жестком диске** (по умолчанию там прописан путь разработчика проекта) [factual].

*Примечание: Если персистентность данных на реальном железе хоста не требуется, переключите флаг `isLocal` в положение `false` внутри `app-chart/values.yaml` и выполните стандартный `minikube start`.*

### 2. Универсальная сборка и жесткий деплой (Первый запуск / Полный сброс)

Для **первичного развертывания системы с нуля**, а также для полной автоматической пересборки, обновления версий (сброс в дефолтную `v1`) и жесткого передеплоя всех компонентов кластера используется одна универсальная команда:

```bash
# Полный перепуш Helm-чартов и принудительный хард-реплейс всей инфраструктуры
make push-all-helm-hard-replace-all
```

### 3. ⚠️ КРИТИЧЕСКИЙ ШАГ: Мониторинг логов сразу после запуска/реплейса

Ввиду высокой ресурсоемкости инфраструктурного слоя, в момент применения ядерного пайплайна отдельные консьюмеры могут аварийно завершать работу. **Необходимо принудительно запустить диагностику сразу после выполнения команды `hard-replace-all`**, чтобы исключить непредсказуемое поведение системы [factual]:

```bash
# Чтение логов консьюмера авторизации сразу после деплоя (переменная TAIL задает количество строк)
make get-logs-auth-user TAIL=100
```

**Действие при падении консьюмеров:**
Если логи на старте сигнализируют о потере соединения с брокером или падении горутин консьюмера, выполните принудительный перезапуск кластера для восстановления внутренних сетевых маршрутов. **Важно учитывать режим локали при старте [factual]:**

```bash
# 1. В любом режиме принудительно тушим зависший кластер
minikube stop

# 2. Поднимаем обратно в зависимости от флага isLocal:
make minikube-start-local   # 💡 Если работаем С локальными томами (isLocal: true)
# ИЛИ
minikube start              # 💡 Если работаем БЕЗ локальных томов (isLocal: false)
```

---

## 🌐 Маршрутизация REST API 

Для выполнения прямых HTTP REST запросов к API через внешние клиенты (Postman, cURL) необходимо получить актуальный сетевой маппинг Ingress-контроллера, который динамически выделяется при старте Nginx [factual].

```bash
# Получить актуальный IP и порт Ingress-шлюза
make get-services-port
```

**Системный вызов под капотом:**
```bash
minikube service ingress-nginx-controller --namespace=ingress-nginx
```
Полученный из таблицы `IP:PORT` (соответствующий внутреннему HTTP-порту `:80`) является **единственной и универсальной точкой входа** для взаимодействия со всеми REST-ручками микросервисов экосистемы [factual].

---

## 🔐 Документация API (Swagger) и Query-фильтрация

Сборка и агрегация OpenAPI/Swagger спецификаций со всех внутренних микросервисов осуществляется централизованным шлюзом на базе единого Ingress-адреса [factual].

* **Точка доступа (Документация):** `http://<АКТУАЛЬНЫЙ_IP>:<АКТУАЛЬНЫЙ_ПОРТ>/swagger/index.html`

Шлюз обрабатывает запросы к спецификации в двух режимах [factual]:
1. **Агрегированный:** Без передачи параметров возвращается общая схема, объединяющая эндпоинты всех активных микросервисов [factual].
2. **Изолированный (Query-фильтрация):** Передача параметра `?service=<name>` позволяет выгрузить спецификацию конкретного компонента [factual] (например, `?service=auth` или `?service=budget`).

</details>

<details>
<summary><b>🇺🇸 Click here to read in English (EN)</b></summary>

## 📊 About The Project
**GO Budget Planner** is a high-performance microservice ecosystem designed for real-time expense monitoring, transaction logging by categories, and percentage-based daily analytics calculation.

### 🚀 Key Features:
* **Transaction Monitoring:** Expense registration, category distribution, and daily financial analytics.
* **Asynchronous Notifications:** Instant email notifications triggered by balance changes. Implemented via the Transactional Outbox pattern (Redis `List` broker + batched `LMPOP` ➡️ Apache Kafka).
* **Aggregated Swagger Gateway:** A centralized API documentation assembly layer deployed directly inside the cluster.

---

## 🛠️ Deployment & Universal Infrastructure Hard Replace (Makefile)

Running the entire architecture is **only possible inside the Kubernetes (Minikube) cluster** due to tight integration with infrastructure primitives (Apache Kafka with SCRAM auth, Redis instances, and PostgreSQL engines).

The `values.yaml` configuration file includes an infrastructure safety guard flag: **`isLocal`**. When set to `true`, the manifests require local host directories to be explicitly mounted into the Minikube node. This guarantees that your PostgreSQL and Redis storage states survive a complete cluster teardown.

### 1. Local Deployment with Host Hardware Mounting

```bash
# Start the cluster with automated local folder mapping to /mnt/data
make minikube-start-local
```
⚠️ **IMPORTANT:** Before running this command, open the `Makefile` and **replace the `LOCAL_PATH_VOLUMES` variable value with your actual local absolute directory path** (by default, it contains the core project developer's path) [factual].

*Note: If data persistence on the host machine's hardware is not required, switch the `isLocal` flag to `false` in `app-chart/values.yaml` and run standard `minikube start`.*

### 2. Universal Rebuild and Force Deployment (First Run / Full Reset)

To execute a **clean system deployment from scratch**, or to run a complete automated rebuild, push configurations, reset tag versions to default `v1`, and force-redeploy all cluster components at once, trigger the unified command:

```bash
# Complete Helm charts push and infrastructure hard-replace
make push-all-helm-hard-replace-all
```

### 3. ⚠️ CRITICAL STEP: Post-Deployment Log Auditing

Due to the high resource overhead of the stateful infrastructure layer, some message consumers might drop or crash during a hard replacement cycle. **You must explicitly review the consumer logs immediately after triggering the `hard-replace-all` command** to prevent unpredictable system states [factual]:

```bash
# Stream Auth Service logs immediately after deployment (The TAIL variable configures the buffer size)
make get-logs-auth-user TAIL=100
```

**Recovery Procedure:**
If logs indicate dead consumers or connection timeouts to the Kafka broker, trigger an immediate Minikube node restart to gracefully rebuild the internal cluster routing tables. **Make sure to respect your deployment environment mode when starting the node [factual]:**

```bash
# 1. Force stop the freezing cluster state across all modes
minikube stop

# 2. Spin the cluster back up according to your active isLocal flag value:
make minikube-start-local   # 💡 If working WITH mounted host persistent volumes (isLocal: true)
# OR
minikube start              # 💡 If working WITHOUT host storage attachments (isLocal: false)
```

---

## 🌐 REST API Routing 

To interact with the microservice REST API using external clients (such as Postman or cURL), you must retrieve the active port mapping assigned dynamically during the Ingress controller startup [factual].

```bash
# Retrieve the active Ingress routing table
make get-services-port
```

**Under the hood Kubernetes routine:**
```bash
minikube service ingress-nginx-controller --namespace=ingress-nginx
```
The resulting `IP:PORT` mapped to internal HTTP port `:80` acts as your **sole and universal base URL** for executing all incoming REST operations across the entire microservice ecosystem [factual].

---

## 🔐 API Documentation (Swagger) & Query Filtering

API specifications from all internal microservices are dynamically aggregated by a centralized routing gateway running behind the unified Ingress address [factual].

