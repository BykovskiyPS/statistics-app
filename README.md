## **Install**
``` 
docker network create net-app 
```
 
```
docker build -t app-image . 
```

```
docker run --name db \
 -d --network net-app \
 --network-alias mysql \
 -v ${PWD}/init.sql:/docker-entrypoint-initdb.d/init.sql \
 --env-file env-db.txt \
mysql:5.7
```

```
docker run --name stats-app -dp 8080:8080 --env-file env-app.txt app-image 
```