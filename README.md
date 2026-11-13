# go_home

## Run

1. Create the config files and fill in the values (they must match each other):

   ```bash
   cp config/main.example.yaml config/main.yaml
   cp .env.docker.example .env.docker
   ```

2. Start PostgreSQL and run the app:

   ```bash
   make db-up
   go run ./cmd
   ```