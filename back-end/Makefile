run-server:
	docker compose up --build

run-import-kaggle:
	make reset-db && make run-seed && go run cmd/import-kaggle/main.go

reset-db:
	go run cmd/reset-db/main.go

run-seed:
	go run cmd/seed/main.go