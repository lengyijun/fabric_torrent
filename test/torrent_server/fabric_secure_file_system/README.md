# RESTful API for fabric sample chaincode
## Introduction
This repo use Hyperledger Fabric 1.0.5 to configure the blockchain network.  
Set up a single node test network.  Code the chaincode in golang. Use nodejs SDK to implement APIs
and event listeners.  
My references: /fabric-samples/basic-network  &&  /fabric-samples/fabcar  
## Dependency
* npm & node
* docker & docker-compose
* For other prerequisites please view [fabric Doc](http://hyperledger-fabric-doc.readthedocs.io/en/latest/prereqs.html)
## Get Started
* Download fabric images
[view the Doc](http://hyperledger-fabric-doc.readthedocs.io/en/latest/samples.html#binaries)
* Clear outdated containers and images
```shell
./clear.sh
```
warning: This script will remove all docker containers and chaincode images. Be careful or do the clear-up job by yourself.
* Setup fabric network
```shell
./startNetwork.sh
```
* Setup chaincode
```shell
./startChaincode.sh
```
chaincode files are in directory ./chaincode
* Install nodejs packages
```shell
npm install
```
* Resgister at least one user first
```shell
# register an admin user
node enrollAdmin.js
# register a user named 'user1'. The name can be edited in the script
node registerUser.js
```
* Run express server as a RESTful API
```shell
npm run dev
# listen on port 3000
```
## Usage
### login
POST http://hostname:3000/users/login  
params: username  
e.g.
```
POST http://127.0.0.1:3000/users/login?username=user1
Expected result:
{
    "success": true,
    "username": "user1"
}
```
### logout
POST http://hostname:3000/users/logout
### create a file
POST http://hostname:3000/file  
params: name, hash, keyword, summary  
e.g.  
```
POST http://127.0.0.1:3000/file?name=name1&hash=hash1&keyword=key1&summary=sum1
Expected result:
{"success":true,"tx_id":"cded19e5da0c61854c9049a27001ac5ee757f3bc67f84b2e1fbd9e494884e66b"}
```
File object has another field: owner  
owner = creator of the transaction  
### delete a file
DELETE http://hostname:3000/file  
params: keyword, name, owner  
chaincode will check transaction creator's cretificate. If the creator of the transaction differ file.owner, the transaction will fail
e.g.  
```
DELETE http://127.0.0.1:3000/file?keyword=key1&name=name1&owner=user1
Expexted result:
{"success":true,"tx_id":"3d72f82dcadfd71082caf2d268dccc80d95addd58e72b342ae71dd359d179468"}
```
### Query files
GET http://hostname:3000/file  
params: keyword, name, owner  
if three params are all empty, it will query all files  
if only keyword is provided, it will query all files with this keyword  
In the same way, if all three params are provided, it will query the exact target file  
key-value database leveldb does not support conplicated retrivals.  
e.g.  
```
GET http://127.0.0.1:3000/file?keyword=key1&name=&owner=
Expected result:
{"success":true,"message":"[{\"Key\":{\"objectType\":File\", \"attributes\":[\"key1\", \"name1\",\"user1\"]},\"Record\":{\"hash\":\"hash1\",\"keyword\":\"key1\",\"name\":\"name1\",\"owner\":\"user1\",\"summary\":\"sum1\"}},
{\"Key\":{\"objectType\":File\",\"attributes\":[\"key1\", \"name2\", \"user1\"]},\"Record\":{\"hash\":\"hash2\",\"keyword\":\"key1\",\"name\":\"name2\",\"owner\":\"user1\",\"summary\":\"sum2\"}}]"}
```

### About the prototype of KEY exchange protocal
* Request struct
```
type Request struct {
    From string `json:"from"`
    To string `json:"To"`
    File string `json:"file"`
    RequestTime int64 `json:"requestTime"`
    ResponseTime int64 `json:"responseTime"`
    ConfirmationTime int64 `json:"confirmationTime"`
}
```
* step 1 Request  
User1 request File1 and provide File1.key  
Create RequestTime timestamp  
* step 2 Respond  
User2 is the owner of File1. After he receives the request proposal(using events), he should upload Secret and create a Respond event.  
ResponseTime timestamp is created.
A locktime is created as an attribute of File1. Assume File1.locktime = t and T is a constant time period  
When t &lt; current time &lt; t + 2T, other Request cannot be accepted.
* step 3 Confirm  
when current time &lt; t + T, File1 cannot be edited or deleted and it is for User1 to confirm the secret.  
User1 can invoke Confirm function and create ConfirmationTime timestamp.  
If time is out, User1 abandon terminate the protocol.  
* step 4 Preparation for next request.  
When t + T &lt; current time &lt; t + 2T, User2(the owner) can freely edit the File1.  
For example, User2 can change the secret password of File1 when he realized the leak of secret.  


### Event Listener
in ./src/routes/event.js  
```js
var promise = new Promise( (resolve, reject) => {
    event_hub.registerChaincodeEvent(network.app_name[1], 'requestSecret', function(ev) {
        console.log("catch requestSecret event", ev.payload.toString());
        // do something
    },
    function() {
        console.log("event listener stopped");
    }); 

    event_hub.registerChaincodeEvent(network.app_name[1], 'respondSecret', function(ev) {
        console.log("catch respondSecret event", ev.payload.toString());
        // do something
    },
```
requestSecret and respondSecret are two custom EVENT object. When the user successfully invoke requestSecret and respondSecret function, these two EVENT will be recorded on ledger.  
And the client can listen to these events and do something to respond.

### Request a file
POST http://hostname:3000/exchange  
params: keyword, name, owner  
e.g.  
```
login as user2
POST http://127.0.0.1:3000/exchange?keyword=key1&name=name1&owner=user1
Expected result:
{"success":true,"tx_id":"57e0fa4935f8feac17678d2f2675dd1ce32ed3b7ef7f040755fd44519a58672d"}
Caught event:
{"from":"user2","To":"user1","file":"\u0000File\u0000key1\u0000name1\u0000user1\u0000","tx_id":"57e0fa4935f8feac17678d2f2675dd1ce32ed3b7ef7f040755fd44519a58672d"}
```

### Respond a request
PUT http://hostname:3000/exchange  
params: tx_id (of the request transaction), secret (key to decrypt file)  
e.g.  
```
PUT http://127.0.0.1:3000/exchange?tx_id=57e0fa4935f8feac17678d2f2675dd1ce32ed3b7ef7f040755fd44519a58672d&secret=secret1
Expected result:
{"success":true,"tx_id":"90fd66c11cf11e4a6a6f12d5c300f6525ab3d08c2f115ac59f64401d9b77e6bf"}
Caught event:
{"from":"user1","To":"user2","file":"\u0000File\u0000key1\u0000name1\u0000user1\u0000","tx_id":"57e0fa4935f8feac17678d2f2675dd1ce32ed3b7ef7f040755fd44519a58672d","secret":"secret1"}
```

### Confirm secret received
DELETE http://hostname:3000/exchange  
params: tx_id (of the request transaction)  
e.g.  
```
DELETE http://127.0.0.1:3000/exchange?tx_id=57e0fa4935f8feac17678d2f2675dd1ce32ed3b7ef7f040755fd44519a58672d
Expected result:
{"success":true,"tx_id":"89e6ae05fd2552f782cb2ef57428ff0449587ff72f2b076480cf0c9f9f905391"}
```

### Query one Request
GET http://hostname:3000/exchange  
params: tx_id (of the request transaction)  
e.g.  
```
GET http://127.0.0.1:3000/exchange?tx_id=3a563d1cdc73b05bf8bbe8898ac433fd30603c0454e4a311a0342d2cb84da4e5
expected result:
{"success":true,"message":"{\"Key\":\"3a563d1cdc73b05bf8bbe8898ac433fd30603c0454e4a311a0342d2cb84da4e5\",
\"Record\":{\"To\":\"user1\",\"confirmationTime\":0,\"file\":\"\\u0000File\\u0000key1\\u0000name1\\u0000user1\\u0000\",\"from\":\"user1\",\"requestTime\":1519357346,\"responseTime\":1519357363}}"}
```

### Other considerations
For other Fabric networks, please edit the config in ./src/routes/setup.js
```js
var Network = {
    config: {channel:"mychannel", order_addr:'grpc://localhost:7050', peer_addr:'grpc://localhost:7051', event_addr:'grpc://localhost:7053'},
    app_name: "myapp",
    fabric_client: null,
    channel: null,
    peer: null,
    store_path: path.join(__dirname, "../../hfc-key-store"),
    user: null,
    order: null,
    event_hub: null,
...
```
## TODO list
* relational database implemented in chaincode
