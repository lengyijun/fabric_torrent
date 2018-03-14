var express = require('express');
var router = express.Router();

/* GET home page. */
router.get('/', function(req, res, next) {
  res.render('index', { title: 'Home Page' });
});

router.get('/file-list', function(req, res, next) {
  res.render('filelist', { title: 'File List'} );
});

router.get('/my-file', function(req, res, next) {
  res.render('myfile', { title: 'My File'} );
});

router.get('/request', function(req, res, next) {
  res.render('request', { title: 'My Request'} );
});

router.get('/message', function(req, res, next) {
  res.render('message', { title: 'My Message'} );
});

module.exports = router;
