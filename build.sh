#!/bin/bash
APP_NAME="STEM-EMBASSY_WEBSERVER" 
MAIN_FILE="main.go"  
OUTPUT_DIR="./bin"

mkdir -p $OUTPUT_DIR

echo "Building $APP_NAME for multiple platforms..."

# Build for Windows (amd64)
echo "Building for Windows..."
GOOS=windows GOARCH=amd64 go build -o $OUTPUT_DIR/$APP_NAME-windows-amd64.exe $MAIN_FILE
if [ $? -eq 0 ]; then
    echo "✅ Windows build successful"
else
    echo "❌ Windows build failed"
fi

# Build for Raspberry Pi (ARM)
echo "Building for Raspberry Pi..."
GOOS=linux GOARCH=arm GOARM=7 go build -o $OUTPUT_DIR/$APP_NAME-raspberrypi-arm $MAIN_FILE
if [ $? -eq 0 ]; then
    echo "✅ Raspberry Pi build successful"
else
    echo "❌ Raspberry Pi build failed"
fi

# Build for MacBook (both Intel and Apple Silicon)
echo "Building for MacBook (Intel)..."
GOOS=darwin GOARCH=amd64 go build -o $OUTPUT_DIR/$APP_NAME-macos-amd64 $MAIN_FILE
if [ $? -eq 0 ]; then
    echo "✅ MacBook Intel build successful"
else
    echo "❌ MacBook Intel build failed"
fi

echo "Building for MacBook (Apple Silicon)..."
GOOS=darwin GOARCH=arm64 go build -o $OUTPUT_DIR/$APP_NAME-macos-arm64 $MAIN_FILE
if [ $? -eq 0 ]; then
    echo "✅ MacBook Apple Silicon build successful"
else
    echo "❌ MacBook Apple Silicon build failed"
fi

echo "Build process completed. Binaries are located in $OUTPUT_DIR/"
ls -la $OUTPUT_DIR