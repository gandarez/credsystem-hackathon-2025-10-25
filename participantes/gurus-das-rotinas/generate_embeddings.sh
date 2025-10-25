#!/bin/bash

# Script to generate embeddings from both CSV files
# Usage: ./generate_embeddings.sh

set -e

echo "🚀 Generating embeddings from CSV datasets..."

# Check if OPENROUTER_API_KEY is set
if [ -z "$OPENROUTER_API_KEY" ]; then
    echo "❌ Error: OPENROUTER_API_KEY environment variable is required"
    exit 1
fi

# Create embeddings directory if it doesn't exist
mkdir -p embeddings

# Generate embeddings for the expanded dataset (more comprehensive)
echo "📊 Processing dataset_expandido.csv (184 examples)..."
cd cmd/generate-embeddings
go run main.go ../../../assets/dataset_expandido.csv > ../../embeddings/expanded_embeddings.log 2>&1

if [ $? -eq 0 ]; then
    echo "✅ Successfully generated embeddings for expanded dataset"
    mv service_embeddings.json ../../embeddings/expanded_service_embeddings.json
else
    echo "❌ Failed to generate embeddings for expanded dataset"
    exit 1
fi

# Generate embeddings for the original dataset
echo "📊 Processing intents_pre_loaded.csv (95 examples)..."
go run main.go ../../../assets/intents_pre_loaded.csv > ../../embeddings/original_embeddings.log 2>&1

if [ $? -eq 0 ]; then
    echo "✅ Successfully generated embeddings for original dataset"
    mv service_embeddings.json ../../embeddings/original_service_embeddings.json
else
    echo "❌ Failed to generate embeddings for original dataset"
    exit 1
fi

cd ../..

echo "🎉 All embeddings generated successfully!"
echo "📁 Files created:"
echo "   - embeddings/expanded_service_embeddings.json"
echo "   - embeddings/original_service_embeddings.json"
echo "   - embeddings/expanded_embeddings.log"
echo "   - embeddings/original_embeddings.log"

echo ""
echo "💡 Next steps:"
echo "   1. Copy the expanded_service_embeddings.json to your main application"
echo "   2. Update your application to use embedding-based classification"
echo "   3. Test the improved accuracy!"
