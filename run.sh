#!/bin/sh

echo "Choose an option to run the application:"
echo "1. Run with Docker"
echo "2. Run with go run main.go"
read -p "Enter your choice [1 or 2]: " choice

if [ "$choice" -eq 1 ]; then
    # Stop and remove any existing container named 'forum'
    docker stop forum 2>/dev/null || true
    docker rm forum 2>/dev/null || true

    # Remove unused Docker images to free up space
    docker image prune -f

    # Build the Docker container without using cache
    docker build --no-cache -t forum . 

    # Clear the terminal
    clear

    # Run the Docker container
    docker run -p 8080:8080 --name forum forum
elif [ "$choice" -eq 2 ]; then
    # Run the application using go run main.go
    clear
    go run main.go
else
    echo "Invalid choice. Please run the script again and choose either 1 or 2."
    exit 1
fi