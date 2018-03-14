// setup fabric basic network
var Fabric_Client = require('fabric-client');
var path = require('path');
var util = require('util');
var os = require('os');

var Network = {
    config: {channel:"orgchannel", order_addr:'grpc://localhost:7050', peer_addr:'grpc://localhost:7051', event_addr:'grpc://localhost:7053'},
    app_name: ["myapp", "keyExchange"],
    fabric_client: null,
    channel: null,
    peer: null,
    store_path: path.join(__dirname, "../../hfc-key-store"),
    user: null,
    order: null,
    event_hub: null,

    // init network
    init: function(uname) {
        console.log("initiating the fabric network");
        // init basic network config
        this.fabric_client = new Fabric_Client();
        this.channel = this.fabric_client.newChannel(this.config.channel);
        this.peer = this.fabric_client.newPeer(this.config.peer_addr);
        this.order = this.fabric_client.newOrderer(this.config.order_addr);
        this.event_hub = this.fabric_client.newEventHub();
        this.event_hub.setPeerAddr(this.config.event_addr);
        this.channel.addPeer(this.peer);
        this.channel.addOrderer(this.order);
        console.log("store path: ", this.store_path);

        // create the key value store as default config
        Fabric_Client.newDefaultKeyValueStore({ path: this.store_path }).then((state_store) => {
            this.fabric_client.setStateStore(state_store);
            var crypto_suite = Fabric_Client.newCryptoSuite();
            var crypto_store = Fabric_Client.newCryptoKeyStore({path: this.store_path});
            crypto_suite.setCryptoKeyStore(crypto_store);
            this.fabric_client.setCryptoSuite(crypto_suite);
            console.log("fabric network initialization done");
            return this.fabric_client;
        }).then(function(client) {
            client.getUserContext(uname, true).then((user_from_store) => {
                if (user_from_store && user_from_store.isEnrolled()) {
                    console.log("login as", uname);
                } else {
                    console.log("login failed");
                }
            });
        });
    }

}

var Networks = {
    clientList: [],
    addClient: function(uname) {
        if (typeof this.clientList[uname] !== 'undefined') {
            //pass
        } else {
            this.clientList[uname] = Network;
            this.clientList[uname].init(uname);
        }
    }
}


module.exports = Networks;
