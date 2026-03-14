.PHONY: up down logs db-shell

up:
	docker-compose up -d --build

down:
	docker-compose down

logs:
	docker-compose logs -f

db-shell:
	docker exec -it ticket_db psql -U user -d ticket_db

load-test:
	docker run --rm -i grafana/k6 run - < tests/load/flash_sale_test.js
