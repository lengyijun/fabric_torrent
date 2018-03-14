var express = require('express');
var path = require('path');
var favicon = require('serve-favicon');
var logger = require('morgan');
var cookieParser = require('cookie-parser');
var bodyParser = require('body-parser');
var session = require('express-session');
var hbs = require('hbs');
var mongoose = require('mongoose');

var network = require('./routes/setup');
var event_listener = require('./routes/event');

var index = require('./routes/index');
var users = require('./routes/users');
var file = require('./routes/file');
var exchange = require('./routes/exchange');
var query = require('./routes/query');
var app = express();

// database settings
global.dbHandler = require('./database/dbHandler.js');
global.db = mongoose.connect("mongodb://localhost:27017/nodedb");

// view engine setup
app.set('views', path.join(__dirname, 'views'));
app.engine('html', hbs.__express);
app.set('view engine', 'html');

// uncomment after placing your favicon in /public
//app.use(favicon(path.join(__dirname, 'public', 'favicon.ico')));
app.use(logger('dev'));
app.use(bodyParser.json());
app.use(bodyParser.urlencoded({ extended: false }));
app.use(cookieParser());
app.use(express.static(path.join(__dirname, 'public')));

app.use(session({
    resave: false,
    saveUninitialized: true,
    secret: 'secret',
    cookie: {
        maxAge: 10000*60*30
    }
}));

app.use(function(req, res, next) {
    if (! req.session.user && req.originalUrl != '/' &&  req.originalUrl.slice(0,12) != '/users/login') {
        req.session.user = "";
        req.session.error = "please login first";
        res.redirect('/');
    } else {
        next();
    }
});

app.use(function(req, res, next) {
    res.locals.user = req.session.user;
    var err = req.session.error;
    var suc = req.session.success;
    delete req.session.error;
    delete req.session.success;
    res.locals.message = "";
    if (err) {
      res.locals.message = '<div class="alert alert-danger" style="width:100%;margin-bottm;20px;color:red;">'+err+'</div>';
    } else if (suc) {
     res.locals.message = '<div class="alert alert-success" style="width:100%;margin-bottm;20px;color:green;">'+suc+'</div>';
    }
    next();
});


app.use('/', index);
app.use('/query', query);
app.use('/file', file);
app.use('/users', users);
app.use('/exchange', exchange);

// catch 404 and forward to error handler
app.use(function(req, res, next) {
  var err = new Error('Not Found');
  err.status = 404;
  next(err);
});

// error handler
app.use(function(err, req, res, next) {
  // set locals, only providing error in development
  res.locals.message = err.message;
  res.locals.error = req.app.get('env') === 'development' ? err : {};

  // render the error page
  res.status(err.status || 500);
  res.render('error');
});

module.exports = app;
