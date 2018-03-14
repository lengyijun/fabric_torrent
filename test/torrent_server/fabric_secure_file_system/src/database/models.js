// database schema
//
module.exports = {

  file: {
    name: {type:String, required:true},
    hash: {type:String, required:true},
    keyword: {type:String, required:true},
    summary: {type:String, required:true},
    owner: {type:String, required:true},
    magnet: {type:String}
  },

  request: {
    from: {type:String, required:true},
    to: {type:String, required: true},
    file: {type:String, required: true},
    name: {type:String},
    keyword: {type:String},
    owner: {type:String},
    tx_id: {type:String, required:true},
    secret: {type:String},
    requestTime: {type:Number},
    responseTime: {type:Number},
    confirmationTime: {type:Number}
  }

};
