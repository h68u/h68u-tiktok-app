# DB Config

1. GET MySQL

```
docker pull mysql
docker run --name mysql -p 你的主机端口:3306 -e MYSQL_ROOT_PASSWORD='your password' -d mysql:latest
dockr exec -it 'CONTAINER ID' bash
mysql -u root -p
```

2. CREATE DB
```
CREATE DATABASE tiktok;
USE tiktok;
```