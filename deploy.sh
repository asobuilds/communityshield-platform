#!/bin/bash

echo "🚀 Starting deployment..."

# Build backend
echo "📦 Building backend..."
cd backend
go build -o main .

# Build frontend
echo "📦 Building frontend..."
cd ../frontend
npm install
npm run build

echo "✅ Build complete!"

# Push to GitHub
echo "📤 Pushing to GitHub..."
cd ..
git add .
git commit -m "Deploy: $(date)"
git push origin main

echo "✅ Deployment complete!"
echo "🔗 Visit: https://communityshield-frontend-app.onrender.com"