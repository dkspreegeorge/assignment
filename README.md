# JSON/HTTP Timestamp Service

This is an implementation of a JSON/HTTP service in Golang that returns the matching timestamps of a periodic task. It listens to port 8080 if you do not give an argument, or you can give as command-line argument the listen addr/port.

## Usage

1. Clone the repository to your local machine.

```
git clone https://github.com/dkspreegeorge/assignment
```

2. Change into the project directory.


```
cd assignment
```

3. Build and run the service.

```
go mod init TaskService
go mod tidy
go build TaskService.go
./TaskService
```


Then use curl to make the call
```
curl -X GET "http://localhost:8089/ptlist?period=1h&tz=Europe/Athens&t1=20210714T204603Z&t2=20210715T123456Z"
```





## DOCKER

In order to run it on docker you first need to cd into project folder and then

```
docker build -t periodtask
docker run periodtask 0.0.0.0:8089
```
