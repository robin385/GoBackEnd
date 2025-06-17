# Build the image
docker build -t trashman .

# Run the container
docker run -p 3000:3000 trashman