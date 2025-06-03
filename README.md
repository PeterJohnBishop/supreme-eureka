# supreme-eureka

## Store files in a Docker Container.

POST   /upload     
GET    /files              
GET    /download/:filename       
DELETE /delete/:filename     

docker build -t peterjbishop/supreme-eureka:latest . 
docker push peterjbishop/supreme-eureka:latest 
docker pull peterjbishop/supreme-eureka:latest 
docker-compose down 
docker-compose build --no-cache 
docker-compose up

## Confirm all files are deleted

docker run --rm -it -v uploader-data:/data alpine sh
ls -lh /data/uploads

## Check storage usage 

docker volume inspect uploader-data
sudo du -sh /var/lib/docker/volumes/uploader-data/_data/uploads
