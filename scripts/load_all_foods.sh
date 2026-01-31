#!/bin/bash

# Helper script to load all generated SQL files into the database
# Usage: ./load_all_foods.sh

set -e  # Exit on any error

echo "Loading food databases into PostgreSQL..."
echo "=========================================="

# Counter for successful loads
loaded=0
skipped=0

# Loop through all .sql files in the current directory
for file in *.sql; do
    # Skip if no .sql files found
    if [ ! -f "$file" ]; then
        echo "No SQL files found"
        exit 0
    fi

    # Check if file contains INSERT statements
    if grep -q "INSERT INTO" "$file"; then
        echo "Loading $file..."
        docker exec -i fooddb-postgres psql -U postgres -d fooddb < "$file"
        ((loaded++))
    else
        echo "Skipping $file (no INSERT statements)"
        ((skipped++))
    fi
done

echo "=========================================="
echo "Summary:"
echo "  - Loaded: $loaded file(s)"
echo "  - Skipped: $skipped file(s)"
echo "All food databases loaded successfully!"
