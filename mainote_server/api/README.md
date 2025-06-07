# Mainote Server API Documentation

This directory contains the OpenAPI 3.0.3 specification for the Mainote Server API, organized in a modular structure for maintainability.

## Structure

```
api/
├── src.yaml                    # Main OpenAPI specification file
├── path/                       # API endpoint definitions
│   ├── health.yaml            # Health check endpoint
│   ├── notes.yaml             # Notes collection endpoints
│   └── NoteById.yaml          # Individual note endpoints
└── schema/                    # Data schemas and models
    ├── components/            # Reusable schema components
    │   ├── index.yaml         # Components index
    │   ├── Note.yaml          # Core Note entity
    │   ├── NoteCategory.yaml  # Note category enum
    │   ├── NoteStatus.yaml    # Note status enum
    │   ├── HealthResponse.yaml # Health response schema
    │   ├── ErrorResponse.yaml  # Error response schema
    │   └── PaginationInfo.yaml # Pagination metadata
    ├── requests/              # Request schemas
    │   ├── CreateNoteRequest.yaml
    │   └── UpdateNoteRequest.yaml
    └── responses/             # Response schemas
        ├── index.yaml         # Responses index
        ├── NotesListResponse.yaml
        ├── BadRequest.yaml
        ├── NotFound.yaml
        ├── InternalServerError.yaml
        └── ServiceUnavailable.yaml
```

## API Endpoints

### Health
- `GET /health` - Service health check

### Notes
- `GET /api/v1/notes` - List notes with filtering and pagination
- `POST /api/v1/notes` - Create a new note
- `GET /api/v1/notes/{noteId}` - Get a specific note
- `PUT /api/v1/notes/{noteId}` - Update a note
- `DELETE /api/v1/notes/{noteId}` - Delete a note

## Usage

To view the API documentation:

1. **Using online tools**: Copy the content of `src.yaml` to [Swagger Editor](https://editor.swagger.io/)
2. **Using local tools**: Install swagger-ui or redoc and point to `src.yaml`
3. **Generate code**: Use openapi-generator to generate client/server code

## Validation

The specification follows OpenAPI 3.0.3 standards with:
- Proper HTTP status codes
- Request/response validation
- Comprehensive error handling
- Clear documentation and examples
