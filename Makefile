SERVICES := auth user habit goal task finance ai notification scheduler

.PHONY: up down logs build test lint clean $(SERVICES)

# Full stack
up:
	docker compose up -d

down:
	docker compose down

logs:
	docker compose logs -f

# Build all service images
build:
	@for svc in $(SERVICES); do \
		echo "=== building $$svc ===" && \
		docker compose build $$svc-service 2>/dev/null || \
		docker compose build $$svc-worker 2>/dev/null || true; \
	done

# Run tests across all services
test:
	@for svc in $(SERVICES); do \
		echo "=== testing $$svc ===" && \
		(cd services/$$svc && go test ./...) || exit 1; \
	done

# Run go vet across all services
lint:
	@for svc in $(SERVICES); do \
		echo "=== vetting $$svc ===" && \
		(cd services/$$svc && go vet ./...) || exit 1; \
	done

# Tidy all go.mod files
tidy:
	@for svc in $(SERVICES); do \
		(cd services/$$svc && go mod tidy); \
	done

# Reset volumes (destructive)
reset: down
	docker compose down -v

# Per-service shortcuts: make auth, make user, etc.
$(SERVICES):
	docker compose up -d $@-service 2>/dev/null || docker compose up -d $@-worker
