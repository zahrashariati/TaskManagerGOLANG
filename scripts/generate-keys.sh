#!/bin/bash

# Generate RSA keys for JWT RS256

echo "🔐 Generating RSA keys for JWT authentication..."

# Create keys directory if it doesn't exist
mkdir -p keys

# Generate private key (2048 bits)
echo "Generating private key..."
openssl genrsa -out keys/private.pem 2048

# Extract public key from private key
echo "Extracting public key..."
openssl rsa -in keys/private.pem -pubout -out keys/public.pem

# Set proper permissions
chmod 600 keys/private.pem  # Only owner can read/write
chmod 644 keys/public.pem   # Readable by all

echo "✅ Keys generated successfully!"
echo ""
echo "Private key: keys/private.pem (KEEP SECRET!)"
echo "Public key:  keys/public.pem (can be shared)"
echo ""
echo "⚠️  IMPORTANT: Add keys/private.pem to .gitignore!"


