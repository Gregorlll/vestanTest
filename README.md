# Prerequisites
- Docker and Docker Compose installed
- Go 1.21 or higher

# 1. Create Docker network (if it doesn't exist)
docker network create chat-network

# 2. Start PostgreSQL and server through docker-compose
docker-compose up --build

# 3. In another terminal, start the client
go run cmd/client/main.go

# 4. In the client, connect with the command
/connect your_username

# 5. In another terminal, start the client
go run cmd/client/main.go

# 6. In another terminal, connect with the command
/connect your_username 