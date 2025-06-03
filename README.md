# supreme-eureka

## Store files in a Docker Container.

POST   /upload                  
GET    /download/:filename       
DELETE /delete/:filename     

docker build -t peterjbishop/supreme-eureka:latest . 
docker push peterjbishop/supreme-eureka:latest 
docker pull peterjbishop/supreme-eureka:latest 
docker-compose down 
docker-compose build --no-cache 
docker-compose up