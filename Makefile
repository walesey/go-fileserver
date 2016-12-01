CONTAINER_NAME = gofileserver
IMAGE_NAME = gofileserver
HOST_PORT = 3000

stopDockerApp:
	-docker rm -f $(CONTAINER_NAME)

# USAGE: make startDockerApp HOST_PORT=8080
startDockerApp: stopDockerApp 
	docker build --tag $(IMAGE_NAME) .
	docker run -d -p $(HOST_PORT):3000 --name $(CONTAINER_NAME) $(IMAGE_NAME)