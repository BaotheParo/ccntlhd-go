.PHONY: up down logs db-shell

up:
	docker-compose up -d --build

down:
	docker-compose down

logs:
	docker-compose logs -f

db-shell:
	docker exec -it ticket_db psql -U user -d ticket_db
