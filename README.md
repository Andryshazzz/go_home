1. Create the config files and fill in the values (they must match each other):

   ```bash
   cp config/main.example.yaml config/main.yaml
   cp .env.docker.example .env.docker
   ```

   The JWT secret must be at least 32 bytes long.

2. Start PostgreSQL and apply the migrations:

   ```bash
   make db-up
   make migrate-up
   ```

3. Run the app:

   ```bash
   go run ./cmd
   ```

   The config path can be changed with a flag (default `./config/main.yaml`):

   ```bash
   go run ./cmd -config ./config/main.yaml
   ```