// setup fabric basic network
var Fabric_Client = require('fabric-client');
var path = require('path');
var util = require('util');
var os = require('os');

var config = {channel:"mychannel", order_addr:'grpc://localhost:7050', peer_addr:'grpc://localhost:7051', event_addr:'grpc://localhost:7053'};
var app_name = ["myapp", "keyExchange"];
var store_path = path.join(__dirname, "../../hfc-key-store");
var fabric_client = new Fabric_Client();
var event_hub = fabric_client.newEventHub();
event_hub.setPeerAddr(config.event_addr);
// create the key value store as default config
Fabric_Client.newDefaultKeyValueStore({ path: store_path }).then((state_store) => {
    fabric_client.setStateStore(state_store);
    var crypto_suite = Fabric_Client.newCryptoSuite();
    var crypto_store = Fabric_Client.newCryptoKeyStore({path: store_path});
    crypto_suite.setCryptoKeyStore(crypto_store);
    fabric_client.setCryptoSuite(crypto_suite);
    return fabric_client.getUserContext('admin', true);
}).then((user_from_store) => {
    event_hub.connect();
    var promise = new Promise( (resolve, reject) => {
        event_hub.registerChaincodeEvent(app_name[1], 'requestSecret', function(ev) {
            console.log("catch requestSecret event", ev.payload.toString());
            var message = JSON.parse(ev.payload.toString());
            var request = global.dbHandler.getModel('request');
            message.responseTime = 0;
            message.confirmationTime = 0;
            message.secret = "";
            var myfile = message.file.split('\u0000');
            message.name = myfile[3];
            message.keyword = myfile[2];
            message.owner = myfile[3];
            request.create(message, function (err, doc) {
                if (err) {
                    console.log("requestSecret", err);
                } else {
                  console.log("requestSecret", message.tx_id);
                }
            });
        },
        function() {
            console.log("event listener stopped");
        }); 

        event_hub.registerChaincodeEvent(app_name[1], 'respondSecret', function(ev) {
            console.log("catch respondSecret event", ev.payload.toString());
            // do something
            var message = JSON.parse(ev.payload.toString());
            var request = global.dbHandler.getModel('request');
            request.update({tx_id:{"$in":message.tx_id} }, 
                {responseTime:message.responseTime, secret:message.secret}, 
            function (err, doc) {
                if (err) {
                    console.log("respondSecret", err);
                } else {
                  console.log("respondSecret", message.tx_id);
                }
            });
        },
        function() {
            console.log("event listener stopped");
        }); 

        event_hub.registerChaincodeEvent(app_name[1], 'confirmSecret', function(ev) {
            console.log("catch confirmSecret event", ev.payload.toString());
            // do something
            var message = JSON.parse(ev.payload.toString());
            var request = global.dbHandler.getModel('request');
            request.update({tx_id:message.tx_id}, 
                {confirmationTime: message.confirmationTime}, 
            function (err, doc) {
                if (err) {
                    console.log("respondSecret", err);
                } else {
                  console.log("respondSecret", message.tx_id);
                }
            });
        },
        function() {
            console.log("event listener stopped");
        }); 

        event_hub.registerChaincodeEvent(app_name[0], 'createFile', function(ev) {
            console.log("catch createFile event", ev.payload.toString());
            // do something
            var message = JSON.parse(ev.payload.toString());
            var owner = message.owner.split('@')[0];
            console.log(owner);
            delete message['owner'];
            owner = owner.toLowerCase();
            message.owner = owner;
            delete message["locktime"];
            var file = global.dbHandler.getModel('file');
            file.create(message, function (err, doc) {
              if (err) {
                console.log("createFile", err);
              } else {
                console.log("createFile", message.name);
              }
            });
        },
        function() {
            console.log("event listener stopped");
        }); 

        event_hub.registerChaincodeEvent(app_name[0], 'deleteFile', function(ev) {
            console.log("catch deleteFile event", ev.payload.toString());
            // do something
            var attr = ev.payload.toString().split('\u0000');
            var file = global.dbHandler.getModel('file');
            var condition = {keyword:attr[2], name:attr[3], owner:attr[4]};
            file.remove(condition, function (err, doc) {
                if (err) {
                     console.log("deleteFile", err);
                } else {
                     console.log("deleteFile", attr[1]);
                }
            });
        },
        function() {
            console.log("event listener stopped");
        }); 

    });
});

var event_listener = {

};

module.exports = event_listener;
