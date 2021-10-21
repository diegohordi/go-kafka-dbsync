run:
	docker-compose -f ./deployments/docker-compose.yml up --build -d

stop:
	docker-compose -f ./deployments/docker-compose.yml down

create_source_connector:
	sh ./scripts/createconnectors.sh
