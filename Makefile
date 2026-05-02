# Local infrastructure for monorepo apps (Postgres/Timescale, Redis, RabbitMQ, Prometheus, Loki, Promtail, Grafana).
# Aligns with apps/template/go-api/.env.example and 2chi-framework/go-api defaults:
#   POSTGRES_DSN=postgres://postgres:postgres@localhost:5432/example?sslmode=disable
#   REDIS_*_DSN=redis://localhost:6379/0
#   RABBITMQ_URL=amqp://guest:guest@localhost:5672/

.PHONY: infra-up infra-down infra-ps infra-pull \
	infra-postgres infra-redis infra-rabbitmq infra-rabbitmq-jobs-topology \
	monitoring-up monitoring-down monitoring-logs infra-grafana-recreate all-up

POSTGRES_CONTAINER := postgres_timescale_postgis
POSTGRES_IMAGE     := timescale/timescaledb-ha:pg16-oss
POSTGRES_PORT      := 5432
POSTGRES_USER      := postgres
POSTGRES_PASSWORD  := postgres
POSTGRES_DB        := example

REDIS_CONTAINER := redis
REDIS_IMAGE     := redis:8-alpine
REDIS_PORT      := 6379

RABBIT_CONTAINER := rabbitmq
RABBIT_IMAGE     := rabbitmq:3-management
RABBIT_AMQP_PORT := 5672
RABBIT_UI_PORT   := 15672

# RabbitMQ management HTTP API (declares job topology; must match apps/template/go-api/internals/jobs queue names).
RABBIT_VHOST        ?= %2F
RABBIT_API          ?= http://localhost:$(RABBIT_UI_PORT)
RABBIT_API_USER     ?= guest
RABBIT_API_PASS     ?= guest
JOB_DLX_EXCHANGE    ?= jobs.dlx
INFRA_COMPOSE       := infra/docker-compose.yml

## infra-up: start Postgres, Redis, RabbitMQ, Prometheus, and Grafana
infra-up:
	docker compose -f $(INFRA_COMPOSE) up -d
	@echo "Infrastructure is up. Postgres :$(POSTGRES_PORT), Redis :$(REDIS_PORT), AMQP :$(RABBIT_AMQP_PORT), RabbitMQ UI http://localhost:$(RABBIT_UI_PORT) (guest/guest)"
	@echo "Optional local proxy (port 80 → Vite): http://localhost — App: http://localhost:3000, API: http://localhost:1337"
	@echo "Monitoring is up. Prometheus http://localhost:9090, Loki http://localhost:3100, Grafana http://localhost:3030 (admin/admin)"
	@echo "Logs: Grafana → Explore → datasource \"Loki\" (Promtail ships Docker container logs)."
	@echo "Grafana dashboard \"Go API — Observability\" is under Dashboards → Browse (not Explore)."
	@echo "If Loki/dashboards are missing after git pull, run: make infra-grafana-recreate"
	@echo "Job consumers expect queues to exist; run once: make infra-rabbitmq-jobs-topology"

## all-up: bring up all local infra and initialize RabbitMQ job topology
all-up: infra-up infra-rabbitmq-jobs-topology
	@echo "All local services are ready."

## infra-down: stop all local infra containers
infra-down:
	docker compose -f $(INFRA_COMPOSE) down

## infra-ps: show these containers
infra-ps:
	docker compose -f $(INFRA_COMPOSE) ps

## infra-pull: refresh images
infra-pull:
	docker compose -f $(INFRA_COMPOSE) pull

infra-postgres:
	docker compose -f $(INFRA_COMPOSE) up -d postgres

infra-redis:
	docker compose -f $(INFRA_COMPOSE) up -d redis

infra-rabbitmq:
	docker compose -f $(INFRA_COMPOSE) up -d rabbitmq

## infra-rabbitmq-jobs-topology: declare DLX, DLQs, bindings, and work queues for template go-api jobs (idempotent PUTs)
infra-rabbitmq-jobs-topology:
	@echo "Declaring job topology (exchange $(JOB_DLX_EXCHANGE)) via $(RABBIT_API) ..."
	@echo "Waiting for RabbitMQ management API to be ready..."
	@attempt=0; until curl -sfu "$(RABBIT_API_USER):$(RABBIT_API_PASS)" "$(RABBIT_API)/api/overview" >/dev/null; do \
		attempt=$$((attempt + 1)); \
		if [ $$attempt -ge 30 ]; then \
			echo "RabbitMQ management API did not become ready in time."; \
			exit 1; \
		fi; \
		sleep 1; \
	done
	curl -sfu "$(RABBIT_API_USER):$(RABBIT_API_PASS)" -H "content-type: application/json" \
		-X PUT "$(RABBIT_API)/api/exchanges/$(RABBIT_VHOST)/$(JOB_DLX_EXCHANGE)" \
		-d '{"type":"direct","durable":true}'
	curl -sfu "$(RABBIT_API_USER):$(RABBIT_API_PASS)" -H "content-type: application/json" \
		-X PUT "$(RABBIT_API)/api/queues/$(RABBIT_VHOST)/cleanup_dlq" \
		-d '{"durable":true,"auto_delete":false,"arguments":{}}'
	curl -sfu "$(RABBIT_API_USER):$(RABBIT_API_PASS)" -H "content-type: application/json" \
		-X PUT "$(RABBIT_API)/api/queues/$(RABBIT_VHOST)/apply_scheduled_plan_changes_dlq" \
		-d '{"durable":true,"auto_delete":false,"arguments":{}}'
	curl -sfu "$(RABBIT_API_USER):$(RABBIT_API_PASS)" -H "content-type: application/json" \
		-X POST "$(RABBIT_API)/api/bindings/$(RABBIT_VHOST)/e/$(JOB_DLX_EXCHANGE)/q/cleanup_dlq" \
		-d '{"routing_key":"cleanup.dlq","arguments":{}}'
	curl -sfu "$(RABBIT_API_USER):$(RABBIT_API_PASS)" -H "content-type: application/json" \
		-X POST "$(RABBIT_API)/api/bindings/$(RABBIT_VHOST)/e/$(JOB_DLX_EXCHANGE)/q/apply_scheduled_plan_changes_dlq" \
		-d '{"routing_key":"apply_scheduled_plan_changes.dlq","arguments":{}}'
	curl -sfu "$(RABBIT_API_USER):$(RABBIT_API_PASS)" -H "content-type: application/json" \
		-X PUT "$(RABBIT_API)/api/queues/$(RABBIT_VHOST)/cleanup" \
		-d '{"durable":true,"auto_delete":false,"arguments":{"x-dead-letter-exchange":"$(JOB_DLX_EXCHANGE)","x-dead-letter-routing-key":"cleanup.dlq"}}'
	curl -sfu "$(RABBIT_API_USER):$(RABBIT_API_PASS)" -H "content-type: application/json" \
		-X PUT "$(RABBIT_API)/api/queues/$(RABBIT_VHOST)/apply_scheduled_plan_changes" \
		-d '{"durable":true,"auto_delete":false,"arguments":{"x-dead-letter-exchange":"$(JOB_DLX_EXCHANGE)","x-dead-letter-routing-key":"apply_scheduled_plan_changes.dlq"}}'
	curl -sfu "$(RABBIT_API_USER):$(RABBIT_API_PASS)" -H "content-type: application/json" \
		-X PUT "$(RABBIT_API)/api/queues/$(RABBIT_VHOST)/playbook_jobs_dlq" \
		-d '{"durable":true,"auto_delete":false,"arguments":{}}'
	curl -sfu "$(RABBIT_API_USER):$(RABBIT_API_PASS)" -H "content-type: application/json" \
		-X POST "$(RABBIT_API)/api/bindings/$(RABBIT_VHOST)/e/$(JOB_DLX_EXCHANGE)/q/playbook_jobs_dlq" \
		-d '{"routing_key":"playbook_jobs.dlq","arguments":{}}'
	curl -sfu "$(RABBIT_API_USER):$(RABBIT_API_PASS)" -H "content-type: application/json" \
		-X PUT "$(RABBIT_API)/api/queues/$(RABBIT_VHOST)/playbook_jobs" \
		-d '{"durable":true,"auto_delete":false,"arguments":{"x-dead-letter-exchange":"$(JOB_DLX_EXCHANGE)","x-dead-letter-routing-key":"playbook_jobs.dlq"}}'
	@echo "Job queues ready."

## monitoring-up: start Prometheus, Loki, Promtail, Grafana
monitoring-up:
	docker compose -f $(INFRA_COMPOSE) up -d prometheus loki promtail grafana

## monitoring-down: stop local monitoring stack
monitoring-down:
	docker compose -f $(INFRA_COMPOSE) stop prometheus loki promtail grafana

## monitoring-logs: tail monitoring container logs
monitoring-logs:
	docker compose -f $(INFRA_COMPOSE) logs -f prometheus loki promtail grafana

## infra-grafana-recreate: reload Grafana provisioning (datasources + dashboards)
infra-grafana-recreate:
	docker compose -f $(INFRA_COMPOSE) up -d --force-recreate grafana
