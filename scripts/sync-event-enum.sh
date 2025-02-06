#!/bin/bash

# Exit on any error
set -e

# Function to extract event constants from Go file within the same const block
extract_events() {
    local file="$1"
    # Create a temporary file to store the const block
    local tempfile=$(mktemp)
    
    # Extract the const block containing Event definitions
    awk '/^const \($/ {p=1;next} /^\)$/ {p=0} p' "$file" > "$tempfile"
    
    # Extract event values from the const block
    grep -o 'Event = "[^"]*"' "$tempfile" | cut -d'"' -f2 | tr '\n' '=' | sed 's/=$//' | sed 's/=/,enum=/g'
    
    # Clean up
    rm "$tempfile"
}

# Function to update JSON schema in Go file
update_schema() {
    local file="$1"
    local events="$2"
    
    # Create temporary file
    local tempfile=$(mktemp)
    
    # Flag to track if we found the generate comment
    local next_line_update=0
    
    # Process the file line by line
    while IFS= read -r line; do
        if [ $next_line_update -eq 1 ]; then
            # Update the line after the comment
            echo "$line" | sed "s/jsonschema:\"enum=[^\"]*\"/jsonschema:\"enum=$events\"/" >> "$tempfile"
            next_line_update=0
        elif [[ $line =~ "next-line-generate event-enum-jsonschema-values" ]]; then
            # Mark the next line for update
            echo "$line" >> "$tempfile"
            next_line_update=1
        else
            echo "$line" >> "$tempfile"
        fi
    done < "$file"
    
    # Replace original file with updated content
    mv "$tempfile" "$file"
}

# Main script
main() {
    local file="${1:-events.go}"  # Default to events.go if no file specified
    
    if [ ! -f "$file" ]; then
        echo "Error: File $file not found"
        exit 1
    fi
    
    echo "Processing $file..."
    
    # Extract events and update schema
    local events=$(extract_events "$file")
    
    if [ -z "$events" ]; then
        echo "Error: No events found in const block"
        exit 1
    fi
    
    update_schema "$file" "$events"
    
    echo "Successfully updated JSON schema enum values"
    echo "New values: $events"
}

# Run the script
main "$@"
