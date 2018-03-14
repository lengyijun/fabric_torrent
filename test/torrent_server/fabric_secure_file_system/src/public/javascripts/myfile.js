$(function() {
  loadPage();
  $('.nav-tabs li:eq(2)').addClass('active');
  function loadPage() {
    $.ajax({
      url:'/query/file',
      type:'post',
      data:{owner:uname},
      success: function(data, status) {
        var myfile = $('#myfile');
        var template = $('#fileTemplate');
        for (var i=0; i<data.length; i++) {
          template.find('.name').text(data[i].name);
          template.find('.keyword').text(data[i].keyword);
          template.find('.owner').text(data[i].owner);
          template.find('.summary').text(data[i].summary);
          myfile.append(template.html());
        }

        $('.delete').each(function(index, element) {
          $(this).click(function() {
            var data = {};
            data.keyword = $(this).siblings('.keyword').text();
            data.name = $(this).siblings('.name').text();
            data.owner = $(this).siblings('.owner').text();
            $.ajax({
              url: '/file',
              type: 'delete',
              data: data,
              succuss: function(data, status) {
                if (data.success) {
                  alert("Operation succeed, transaction id: "+data.tx_id);
                } else {
                  alert(data.message);
                }
                $(this).attr('disabled', 'true');
              },
              error: function(data, status) {
                console.log('error', data);
                alert("something wrong");
              }
            })
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
        console.log("error in loadpage", data);
      }
    });
  }
});

$(document).ready(function() {

  $('#upload-form').on('click', '#upload', function() {
    var formData = new FormData($("#upload-form")[0]);
    $.ajax({
      url: '/file',
      type: 'POST',
      data: formData,
      timeout: 100,
      processData: false,
      contentType: false,
      success: function(data) {
        if (data.success) {
          alert("upload succeed. Transaction id: "+data.tx_id);
        } else {
          alert(data.message);
        }
      },
      error: function(data) {
        console.log("error in upload", data);
      }
    });
  });

});
