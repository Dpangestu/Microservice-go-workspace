Run Tests Final

# Unit tests only

go test -v ./tests/unit/...

# With coverage

go test -cover ./tests/unit/...

# Coverage report

go test -coverprofile=coverage.out ./tests/unit/...
go tool cover -html=coverage.out

# Race detector

go test -race ./tests/unit/...

# Specific test

go test -v -run TestGetCurrentUserBundle ./tests/unit/...
