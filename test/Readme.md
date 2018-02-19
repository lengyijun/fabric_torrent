##related folder
- torrent_cli
- torrent_server

##how to run
first put seeded file under torrent_server/data
```$xslt
make  dockerenv-devstable-up
```
```$xslt
docker exec -it fabsdkgo_server_1 bash
go run multiple_server.go server.go
```
```$xslt
#run after server is ready
docker exec -it fabsdkgo_cli_1 bash
go run multiple_cli.go client.go
```
