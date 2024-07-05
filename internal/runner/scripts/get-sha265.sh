#!/bin/sh

# Define patterns as a space-separated string
patterns="{{patterns}}"

# Initialize files as an empty string
files=""

# Use find to list files for each pattern and append them to the files string
for pattern in $patterns
do
    for file in $(find . -type f -name "$pattern"); do
        files="$files $file"
    done
done

# Check if no files were found
if [ -z "$files" ]; then
    echo "NONE"
    exit 0
fi

# String to store individual hashes
combined_hashes=""

# Compute SHA-256 for each file and append it to the combined_hashes string
for file in $files
do
    hash=$(sha256sum "$file" | grep -o "^[^ ]*")
    combined_hashes="$combined_hashes$hash"
done

# Compute the SHA-256 of the combined string
final_hash=$(echo -n "$combined_hashes" | sha256sum | grep -o "^[^ ]*")

# Print the combined hash
echo "$final_hash"
