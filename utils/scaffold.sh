#!/bin/bash

# Exit immediately if any command fails
set -e
cd ..
echo "🚀 Starting Project Scaffolding for Event Management System..."

# 1. Create Main Project Root Directories
mkdir -p apps/web-react/src/components apps/web-react/src/graphql apps/web-react/src/public
mkdir -p apps/mobile-native/src/components apps/mobile-native/src/graphql
mkdir -p services/gateway-graphql/graph/generated services/gateway-graphql/graph/model
mkdir -p services/service-event/db services/service-event/models
mkdir -p services/service-booking/cache services/service-booking/flags services/service-booking/storage services/service-booking/models
mkdir -p infrastructure/mongo-init

# 2. Touch Files in client apps
touch apps/web-react/src/public/index.html
touch apps/web-react/src/components/BookingButton.js
touch apps/web-react/src/graphql/queries.js
touch apps/web-react/src/graphql/mutations.js
touch apps/web-react/src/App.js
touch apps/web-react/src/index.js
touch apps/web-react/package.json

touch apps/mobile-native/src/components/BookingButton.js
touch apps/mobile-native/src/graphql/queries.js
touch apps/mobile-native/src/graphql/mutations.js
touch apps/mobile-native/src/App.js
touch apps/mobile-native/package.json

# 3. Touch Files in service-event (Go)
touch services/service-event/db/mongodb.go
touch services/service-event/models/event.go
touch services/service-event/main.go

# 4. Touch Files in service-booking (Go)
touch services/service-booking/cache/redis.go
touch services/service-booking/flags/flagsmith.go
touch services/service-booking/storage/s3.go
touch services/service-booking/models/booking.go
touch services/service-booking/main.go

# 5. Touch Files in gateway-graphql (Go)
touch services/gateway-graphql/graph/schema.graphqls
touch services/gateway-graphql/graph/resolver.go
touch services/gateway-graphql/main.go

# 6. Touch Infrastructure configurations
touch infrastructure/docker-compose.yml
touch infrastructure/mongo-init/init.js

# 7. Root Documentation
touch README.md

# ----------------------------------------------------
# Go Modules Initialization
# ----------------------------------------------------
echo "Initializing Go modules..."

cd services/gateway-graphql
go mod init ems-platform/services/gateway-graphql
cd ../..

cd services/service-event
go mod init ems-platform/services/service-event
# Fetch Go driver for MongoDB
go get go.mongodb.org/mongo-driver/mongo
cd ../..

cd services/service-booking
go mod init ems-platform/services/service-booking
# Fetch Redis driver and Flagsmith SDK
go get github.com/redis/go-redis/v9
go get github.com/Flagsmith/flagsmith-go-client/v3
cd ../..

echo "✅ Scaffolding complete! Structure is created, Go modules are initialized, and dependencies are downloaded."