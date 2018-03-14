$(function() {
  loadPage();

  $('.nav-tabs li:eq(3)').addClass('active');
  function loadPage() {
    var query = {from: uname};
    $.ajax({
      url:'/query/request',
      type:'post',
      data:query,
      success: function(data, status) {
        var request = $('#request');
        var template = $('#fileTemplate');
        request.empty();
        for (var i=0; i<data.length; i++) {
          template.find('.panel-title').text('Request at '+ Date(parseInt(data[i].requestTime)));
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
          if (data[i].responseTime == 0) {
            template.find('.confirmation').text('No Confirmation');
          } else {
            template.find('.confirmation').text(Date(parseInt(data[i].confirmationTime)));
          }

          template.find('.download').removeAttr('disabled');
          template.find('.confirm').removeAttr('disabled');
          if (data[i].responseTime == 0) {
            template.find('.download').attr('disabled', 'true');
            template.find('.confirm').attr('disabled', 'true');
          } else if (data[i].confirmationTime != 0) {
            template.find('.confirm').attr('disabled', 'true');
          }
          request.append(template.html());
        }
        $('.confirm').each(function(index, element) {
          $(this).click(function() {
            var data = {};
            data.tx_id = $(this).siblings('.tx_id').text();
            $.ajax({
              url: '/exchange',
              type: 'delete',
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
                console.log('error', data);
                alert('something wrong');
              }
            });
          });
        });

        $('.download').each(function(index, element) {
          $(this).click(function() {
            var filename = $(this).siblings('.name').text();
            window.location.href = "/files/"+filename;
          });
        });


      },
      error: function(data, status) {
        console.log("error", data);
      }
    });
  };


});
