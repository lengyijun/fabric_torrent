$(function() {
  loadPage();
  $('.nav-tabs li:eq(4)').addClass('active');

  function loadPage() {
    var query = {to: uname};
    $.ajax({
      url:'/query/request',
      type:'post',
      data:query,
      success: function(data, status) {
        var message = $('#message');
        var template = $('#fileTemplate');
        var date = new Date();
        message.empty();
        for (var i=0; i<data.length; i++) {
          template.find('.panel-title').text('request at '+ Date(parseInt(data[i].requestTime)));
          template.find('.tx_id').text(data[i].tx_id);
          template.find('.from').text(data[i].from);
          template.find('.to').text(data[i].to);
          template.find('.file').text(data[i].file);
          template.find('.name').text(data[i].name);
          template.find('.keyword').text(data[i].keyword);
          template.find('.owner').text(data[i].owner);


          if (data[i].responseTime == 0) {
            template.find('.response').text('No Response');
          } else {
            template.find('.response').text(Date(parseInt(data[i].responseTime)));
          }
          if (data[i].confirmTime == 0) {
            template.find('.confirm').text('No Confirmation');
          } else {
            template.find('.confirm').text(Date(parseInt(data[i].confirmationTime)));
          }

          template.find('.respond').removeAttr('disabled');
          if (data[i].responseTime != 0) {
            template.find('.respond').attr('disabled', 'true');
          }
          message.append(template.html());
        }

        $('.respond').each(function(index, element) {
          $(this).click(function() {
            var data = {};
            data.tx_id = $(this).siblings('.tx_id').text();
            data.secret = "secret";
            console.log(data);
            $.ajax({
              url: '/exchange',
              type: 'put',
              data: data,
              succuss: function(data) {
                if (data.success) {
                  alert('Operation succeed, transaction id: '+data.tx_id);
                } else {
                  alert(data.message);
                }
                $(this).attr('disabled', 'true');
              },
              error: function(data) {
                alert('something wrong');
                console.log('error', data);
              }
            });
          });
        });
      },
      error: function(data, status) {
        console.log('error', data);
      }
    });
  };
});
