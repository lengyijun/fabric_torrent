## related folder
- torrent_cli
- torrent_server

## how to run
first put seeded file under torrent_server/data
```$xslt
make  dockerenv-stable-up
```
```$xslt
#in another terminal
docker exec -it fabsdkgo_cli_1 bash
./torrent_cli
```
