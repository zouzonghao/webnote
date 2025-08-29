#!/bin/sh
# This script installs dependencies and minifies static assets.

echo "📦 Installing dependencies..."
npm install

echo "🎨 Minifying CSS and JavaScript files..."
npm run minify

echo "✅ Minification complete."