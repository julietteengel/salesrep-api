# Sales Rep API - AI Assistant Context

## Project Overview
Building a MeetRox.ai-inspired platform for analyzing sales calls and tracking performance.

## Tech Stack
- **Backend**: Go 1.24.2, Echo framework, GORM, PostgreSQL
- **Architecture**: Service-Repository pattern with fx dependency injection
- **Auth**: JWT-based authentication
- **Docs**: Swagger/OpenAPI

## Project Structure
```
pkg/
  [domain]/
    - [domain].model.go       # Data models
    - [domain].repository.go  # Database operations
    - [domain].service.go     # Business logic
    - [domain].controller.go  # HTTP handlers
```

## Key Commands
- `make run` - Start the server
- `make swagger` - Generate API documentation
- `make test` - Run tests
- `docker-compose up` - Start with dependencies

## Current Models
- User (with roles: Admin, Manager, Sales Rep)
- Conversation (sales interactions)
- Call (recordings with metrics, transcripts, analysis)
- Performance (user metrics)
- Coaching (feedback and notes)

## Coding Standards
- Use service-repository pattern consistently
- All models extend `common.BaseModel` with ID, timestamps
- Controllers use fx dependency injection
- Error handling via `pkg/utils/errors.go`
- JWT middleware for protected routes

## Database
- PostgreSQL with GORM
- Migrations handled automatically
- Relationships defined in `pkg/models/relationships.go`

## API Patterns
- RESTful endpoints: `/api/v1/[resource]`
- Private routes require JWT: `/api/private/[resource]`
- Consistent error responses using utils.ErrorResponse

## Testing Requirements
- Unit tests for services
- Integration tests for repositories
- Mock external dependencies

## Environment Variables
Required in `.env`:
- DATABASE_URL
- JWT_SECRET
- PORT (default: 8080)

## Next Features to Build
1. Recording upload system (S3 integration)
2. Transcription service (OpenAI/AWS)
3. Call analysis pipeline
4. Performance metrics calculation
5. Dashboard APIs

## Important Notes
- Never commit secrets or API keys
- Follow existing code patterns
- Check existing implementations before creating new files
- Use GORM conventions for database operations