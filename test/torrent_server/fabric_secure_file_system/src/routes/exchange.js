var express = require('express');
var router = express.Router();
var network = require('./setup.js');
var chaincode = require('./chaincode.js');

/* GET & POST invoke createFile and queryFileByPartialKey function in chaincode  */
router.route('/').put(function(req, res, next) {
    // PUT /exchange
    // invoke respondSecret
    // params: tx_id, secret

    var _query = [];
    if (! req.body) var data = req.query;
    else var data = req.body

    var tx_ids = data.tx_id.split(',');
    for (var i = 0; i < tx_ids.length; i++) {
        _query.push(tx_ids[i]);
    }
    _query.push(data.secret);

    var _txId = network.clientList[req.session.user].fabric_client.newTransactionID();
    const request = {
        chaincodeId: network.clientList[req.session.user].app_name[1],
        fcn: 'respondSecret',
        args: _query,
        chainId: network.clientList[req.session.user].config.channal,
        txId: _txId
    };
    return chaincode.invoke(req, res, next, request);

}).post(function(req, res, next) {
    // POST /exchange
    // invoke requestSecret
    // params: keyword, name, owner

    var _query = [];
    if (! req.body) var data = req.query;
    else var data = req.body
    _query.push(data.keyword);
    _query.push(data.name);
    _query.push(data.owner);

    var _txId = network.clientList[req.session.user].fabric_client.newTransactionID();

    const request = {
        chaincodeId: network.clientList[req.session.user].app_name[1],
        fcn: 'requestSecret',
        args: _query,
        chainId: network.clientList[req.session.user].config.channel,
        txId: _txId
    };
    return chaincode.query(req, res, next, request);
}).delete(function(req, res, next) {
    // DELETE /exchange
    // invoke confirmSecret
    // params: tx_id

    var _query = [];
    if (! req.body) var data = req.query;
    else var data = req.body

    _query.push(data.tx_id);
    console.log(_query);
    var _txId = network.clientList[req.session.user].fabric_client.newTransactionID();

    const request = {
        chaincodeId: network.clientList[req.session.user].app_name[1],
        fcn: 'confirmSecret',
        args: _query,
        chainId: network.clientList[req.session.user].config.channel,
        txId: _txId
    };
    return chaincode.invoke(req, res, next, request);
}).get(function(req, res, next) {
    // GET /exchange
    // invoke queryRequest
    // params: tx_id

    var _query = [];
    _query.push(req.query.tx_id);

    const request = {
        chaincodeId: network.clientList[req.session.user].app_name[1],
        fcn: 'queryRequest',
        args: _query,
    };
    return chaincode.query(req, res, next, request);
});

module.exports = router;
