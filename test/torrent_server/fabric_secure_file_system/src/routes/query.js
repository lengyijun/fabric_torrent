var express = require('express');
var router = express.Router();

router.route('/file').post(function(req, res, next) {

    var query = req.body;
    var file = global.dbHandler.getModel('file');
    file.find(query, function(err, data) {
        if (err) {
            return res.sendStatus(500);
        } else {
            return res.send(data);
        }
    });
});


router.route('/request').post(function(req, res, next) {
    var query = req.body;
    var request = global.dbHandler.getModel('request');
    request.find(query, function(err, data) {
        if (err) {
            return res.sendStatus(500);
        } else {
            return res.send(data);
        }
    });

});

module.exports = router;
