# Build the image
docker build -t trashman .

# Run the container
docker run -p 8080:8080 trashman