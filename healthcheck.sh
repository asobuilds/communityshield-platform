#!/bin/bash

echo "🔍 Running health checks..."

# Check backend
echo "📊 Checking backend..."
curl -s https://communityshield-backend.onrender.com/health

# Check frontend
echo "📊 Checking frontend..."
curl -s https://communityshield-frontend-app.onrender.com

echo "✅ Health check complete!"
