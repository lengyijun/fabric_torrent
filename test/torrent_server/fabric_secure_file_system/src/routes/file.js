var express = require('express');
var router = express.Router();
var network = require('./setup.js');
var chaincode = require('./chaincode.js');
var fs = require('fs');
var path = require('path');
var crypto = require('crypto');
var multipart = require('connect-multiparty');
var multipartMid = multipart();

//var multer = require('multer');
//var upload = multer({dest: 'upload/'});

/* GET & POST invoke createFile and queryFileByPartialKey function in chaincode  */
router.route('/').get(function(req, res, next) {
    // GET /file
    // invoke queryFileByPartialKey
    // params: keyword, name, owner

    var _query = [];
    if (! req.body.name) var data = req.query;
    else var data = req.body;
    if (data.keyword.length > 0) _query.push(data.keyword);
    if (data.name.length > 0) _query.push(data.name);
    if (data.owner.length > 0) _query.push(data.owner);

    const request = {
        chaincodeId: network.clientList[req.session.user].app_name[0],
        fcn: 'queryFile',
        args: _query
    };
    return chaincode.query(req, res, next, request);

}).post(multipartMid, function(req, res, next) {
    // POST /file
    // invoke createFile
    // params: name, hash, keyword, summary
    var data = null;
    if (! req.body.name) data = req.query;
    else data = req.body;

    var datapath = path.join(__dirname, '../public/files/' + data.name);
    var oripath = req.files.file.path;
    fs.readFile(oripath, function(err, doc) {
        var md5 = crypto.createHash('md5');
        var _hash = md5.update(doc).digest('hex');
        fs.writeFile(datapath, doc, function(err) {
            if (err) {
                console.log(err);
                return res.send({success:false, message:"file upload failure"});
            } else {
                var _query = [];
                _query.push(data.name);
                _query.push(_hash);
                _query.push(data.keyword);
                _query.push(data.summary);

                var _txId = network.clientList[req.session.user].fabric_client.newTransactionID();

                const request = {
                    chaincodeId: network.clientList[req.session.user].app_name[0],
                    fcn: 'createFile',
                    args: _query,
                    chainId: network.clientList[req.session.user].config.channel,
                    txId: _txId
                };
                return chaincode.invoke(req, res, next, request);
            }
        });
    });

}).delete(function(req, res, next) {
    // DELETE /file
    // invoke deleteFile
    // params: keyword, name, owner

    if (! req.body.name) var data = req.query;
    else var data = req.body;
 
    var _query = [];
    _query.push(data.keyword);
    _query.push(data.name);
    _query.push(data.owner);

    var _txId = network.clientList[req.session.user].fabric_client.newTransactionID();

    const request = {
        chaincodeId: network.clientList[req.session.user].app_name[0],
        fcn: 'deleteFile',
        args: _query,
        chainId: network.clientList[req.session.user].config.channel,
        txId: _txId
    };
    return chaincode.invoke(req, res, next, request);
});

module.exports = router;
