# SocialPredict Example Markets

This directory contains example prediction markets and scripts to populate the SocialPredict database.

## Files

- `example_markets.sql` - SQL file containing 10 example prediction markets
- `populate_markets.sh` - Bash script to populate markets via SQL
- `seed_markets.go` - Go program to populate markets using GORM
- `README.md` - This file

## Example Markets Included

The following 10 diverse prediction markets are included:

1. **Technology**: Will OpenAI release GPT-5 by end of 2025?
2. **Sports**: Will Lionel Messi score 15+ goals in MLS 2025 regular season?
3. **Politics**: Will Donald Trump be the Republican nominee for President in 2028?
4. **Economics**: Will Bitcoin price exceed $150,000 by end of 2025?
5. **Entertainment**: Will Taylor Swift announce a new studio album in 2025?
6. **Science**: Will a cure for Type 1 Diabetes receive FDA approval by 2030?
7. **Weather/Climate**: Will 2025 be the hottest year on record globally?
8. **Space**: Will SpaceX successfully land humans on Mars by 2030?
9. **Business**: Will Tesla stock price exceed $500 per share in 2025?
10. **Gaming**: Will Grand Theft Auto 6 be released in 2025?

Each market includes:
- Clear resolution criteria
- Appropriate resolution dates
- Realistic initial probabilities
- Detailed descriptions

## Usage

### Option 1: Using the Bash Script (Recommended)

The bash script provides comprehensive checks and error handling:

```bash
# Navigate to the project root
cd /path/to/socialpredict

# Run the script with default settings
./scripts/populate_markets.sh

# Check database connectivity only
./scripts/populate_markets.sh --check-only

# With custom .env file
./scripts/populate_markets.sh --env-file /path/to/custom/.env

# With environment variable override
DB_HOST=db.example.com ./scripts/populate_markets.sh
```

### Option 2: Using the Go Program

The Go program integrates with the existing codebase and provides better error handling:

```bash
# Navigate to the project root
cd /path/to/socialpredict

# Run the Go seeder script (handles dependencies automatically)
./scripts/seed_markets_go.sh

# Or manually (requires Go modules setup)
cd scripts
go run seed_markets.go
```

### Option 3: Direct SQL Import

For direct SQL import (advanced users):

```bash
# Make sure your database is running and accessible
psql -h localhost -U user -d devdb -f scripts/example_markets.sql
```

## Prerequisites

Before running any script, ensure:

1. **Database is running**: PostgreSQL server must be accessible
2. **Tables exist**: Run the main application first to create tables via GORM auto-migration
3. **Admin user exists**: The markets reference 'admin' as creator (or modify the SQL/Go code)
4. **Configuration**: Set up your environment using a `.env` file (recommended) or environment variables

### Configuration Setup

#### Option 1: Using .env File (Recommended)

1. Copy the example environment file:
```bash
cp .env.example .env
```

2. Edit `.env` with your actual values:
```bash
# Database connection settings
DB_HOST=localhost
DB_PORT=5432
POSTGRES_USER=user
POSTGRES_PASSWORD=your_actual_password
POSTGRES_DATABASE=devdb

# Admin user configuration
ADMIN_PASSWORD=your_admin_password
```

3. The scripts will automatically load this file

#### Option 2: Environment Variables

If you prefer not to use a `.env` file, set these environment variables:

```bash
export DB_HOST="localhost"
export DB_PORT="5432"
export POSTGRES_USER="user"
export POSTGRES_PASSWORD="password"
export POSTGRES_DATABASE="devdb"
```

**Note**: `.env` file values take precedence over environment variables.

#### Security Considerations

- **Never commit `.env` files**: Add `.env` to your `.gitignore` file
- **Use strong passwords**: Especially for `POSTGRES_PASSWORD` and `ADMIN_PASSWORD`
- **Restrict file permissions**: `chmod 600 .env` (readable only by owner)
- **Use different credentials**: For development vs production environments

## Script Features

### Bash Script (`populate_markets.sh`)

- ✅ Automatic `.env` file loading
- ✅ Database connectivity check
- ✅ Table existence verification
- ✅ User existence check
- ✅ Existing market count
- ✅ Confirmation prompts
- ✅ Colored output
- ✅ Error handling
- ✅ Help documentation
- ✅ Custom `.env` file support

### Go Program (`seed_markets.go` + wrapper)

- ✅ Automatic `.env` file loading
- ✅ GORM integration
- ✅ Proper Go struct usage
- ✅ Database transaction handling
- ✅ Error reporting
- ✅ Interactive prompts
- ✅ Existing data checks
- ✅ Automatic dependency management

## Market Data Structure

Each market includes the following fields:

```go
type Market struct {
    QuestionTitle           string    // The main question
    Description             string    // Detailed resolution criteria
    OutcomeType             string    // "BINARY" for yes/no markets
    ResolutionDateTime      time.Time // When market closes
    FinalResolutionDateTime time.Time // Final resolution deadline
    UTCOffset               int       // Timezone offset
    IsResolved              bool      // Resolution status
    ResolutionResult        string    // Final outcome (null until resolved)
    InitialProbability      float64   // Starting probability (0.0-1.0)
    CreatorUsername         string    // User who created the market
}
```

## Troubleshooting

### Common Issues

1. **Database Connection Error**
   - Check if PostgreSQL is running
   - Verify `.env` file exists and has correct values
   - Ensure database exists
   - Test connection: `./scripts/populate_markets.sh --check-only`

2. **Table Not Found**
   - Run the main application first: `go run main.go`
   - This will create tables via GORM auto-migration

3. **Admin User Not Found**
   - Create admin user by running main application with ADMIN_PASSWORD set in `.env`
   - Or modify the scripts to use a different username

4. **Permission Denied**
   - Make sure the script is executable: `chmod +x populate_markets.sh`
   - Check PostgreSQL user permissions

5. **.env File Issues**
   - Ensure `.env` file is in the project root directory
   - Check file permissions: `ls -la .env`
   - Verify no syntax errors (no spaces around `=`)
   - Use `.env.example` as a template

6. **Go Module Issues**
   - Use the wrapper script: `./scripts/seed_markets_go.sh`
   - Or manually run: `cd scripts && go mod tidy && go run seed_markets.go`

### Getting Help

- Use `./scripts/populate_markets.sh --help` for bash script options
- Check application logs for detailed error messages
- Verify database schema matches the Go model

## Customization

To create your own markets:

1. **Modify SQL file**: Edit `example_markets.sql` with your markets
2. **Update Go program**: Modify the `markets` slice in `seed_markets.go`
3. **Follow the schema**: Ensure all required fields are included
4. **Test thoroughly**: Use `--check-only` flag to verify setup

## Market Categories

The example markets cover diverse categories to demonstrate the platform's versatility:

- **Technology & AI**: GPT-5 release
- **Sports**: Messi goals
- **Politics**: Presidential nominations
- **Finance**: Bitcoin price
- **Entertainment**: Taylor Swift album
- **Science**: Medical breakthroughs
- **Climate**: Temperature records
- **Space**: Mars missions
- **Business**: Stock prices
- **Gaming**: Game releases

Feel free to modify or expand these examples based on your needs!