APP_NAME=splose-clone-be

dev:
	docker compose -f docker-compose.dev.yaml up --build 

prod:
	docker compose -f docker-compose.prod.yaml up --build -d

stop:
	docker compose down

logs:
	docker compose logs -f

clean:
	docker compose down -v --remove-orphans

ps:
	docker compose ps