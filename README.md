# supreme-eureka

An API to create and manage file storage in a Docker managed volume.

Add your own authentication method.

## installation

docker pull peterjbishop/supreme-eureka:latest 
docker-compose build --no-cache 
docker-compose up

## Store files in a Docker Container.

POST   /upload     
GET    /files              
GET    /download/:filename       
DELETE /delete/:filename  

## Download a file to your local device with CURL

example:
 cd {your_download_directory}
 curl -o "The Home Depot - Cart.png" "http://localhost:8080/download/The%20Home%20Depot%20-%20Cart.png"

## Notes

docker build -t peterjbishop/supreme-eureka:latest . 
docker push peterjbishop/supreme-eureka:latest 
docker pull peterjbishop/supreme-eureka:latest 
docker-compose down 
docker-compose build --no-cache 
docker-compose up

