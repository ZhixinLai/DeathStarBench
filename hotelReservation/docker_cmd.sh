sudo docker-compose stop
sudo docker container prune
sudo docker image prune
sudo docker container rm $(sudo docker container ls)
sudo docker image rm $(sudo docker image ls)
sudo docker volume rm $(sudo docker volume ls)
sudo docker-compose up -d
