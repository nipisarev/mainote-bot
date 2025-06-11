# Flyway Database Migrations

This directory contains Flyway database migrations for the Mainote Bot project.

## Directory Structure

```
db/
├── flyway.conf          # Flyway configuration file
├── sql/                 # SQL migration files
│   ├── V1__Initial_schema.sql
│   └── V{version}__{description}.sql
└── README.md           # This file
```

## Migration Naming Convention

Flyway uses a specific naming convention for migration files:

- **Versioned migrations**: `V{version}__{description}.sql`
  - Example: `V1__Initial_schema.sql`
  - Example: `V2__Add_notes_table.sql`
  - Example: `V1.1__Add_user_email_column.sql`

- **Repeatable migrations**: `R__{description}.sql`
  - Example: `R__Create_views.sql`
  - Example: `R__Update_functions.sql`

## Version Numbering

- Use semantic versioning: `V1`, `V2`, `V3`, etc.
- For hotfixes: `V1.1`, `V1.2`, etc.
- For features: `V2`, `V3`, etc.

## Migration Guidelines

### 1. SQL Best Practices
- Use explicit column types and constraints
- Add indexes for frequently queried columns
- Include comments for documentation
- Use transactions where appropriate

### 2. Backward Compatibility
- Avoid breaking changes when possible
- Use `ALTER TABLE ADD COLUMN` instead of recreating tables
- Provide default values for new NOT NULL columns
- Consider data migration scripts for complex changes

### 3. Rollback Strategy
- Keep `undo` migrations for critical changes
- Test rollback procedures in development
- Document any manual rollback steps

### 4. Performance Considerations
- Create indexes concurrently in production
- Consider table locks during migrations
- Test migrations with realistic data volumes

## Development Workflow

### 1. Create New Migration
```bash
# Create new migration file
touch db/sql/V{next_version}__{description}.sql

# Example
touch db/sql/V2__Add_notes_table.sql
```

### 2. Write Migration SQL
```sql
-- Migration: V2__Add_notes_table.sql
-- Description: Add notes table for storing user notes
-- Author: Developer Name
-- Date: 2025-06-10

CREATE TABLE notes (
    id SERIAL PRIMARY KEY,
    chat_id TEXT NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (chat_id) REFERENCES user_preferences(chat_id)
);

CREATE INDEX idx_notes_chat_id ON notes(chat_id);
CREATE INDEX idx_notes_created_at ON notes(created_at);
```

### 3. Test Migration
```bash
# Run migration in development
mainote-cli db-migrate

# Check migration status
mainote-cli db-info

# Rollback if needed (if undo migration exists)
mainote-cli db-undo
```

## Production Deployment

Migrations are automatically applied during deployment:

1. **Development**: Run via `mainote-cli db-migrate`
2. **Production**: Run via Flyway Docker container during deployment

## Troubleshooting

### Common Issues

1. **Migration fails**: Check SQL syntax and database permissions
2. **Checksum mismatch**: Migration file was modified after being applied
3. **Out of order migrations**: Use `flyway repair` to fix metadata

### Recovery Commands

```bash
# Repair metadata table
flyway repair

# Baseline existing database
flyway baseline

# Validate migrations
flyway validate

# Get detailed info
flyway info
```

## Environment Variables

Flyway configuration can be overridden using environment variables:

- `FLYWAY_URL`: Database connection URL
- `FLYWAY_USER`: Database username  
- `FLYWAY_PASSWORD`: Database password
- `FLYWAY_SCHEMAS`: Database schema (default: public)
- `FLYWAY_LOCATIONS`: Migration locations (default: filesystem:./sql)

## Integration with Project

The Flyway migrations are integrated with:

- **Docker Compose**: Automatic migration during development startup
- **Fly.io**: Automatic migration during production deployment  
- **CLI**: `mainote-cli db-*` commands for manual migration management
- **CI/CD**: Validation and testing of migrations
