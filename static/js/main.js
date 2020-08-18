var SubmitForm = function(action) {
  $(".submitbutton").addClass('disabled');
  var message = $('#message').val();
  if (action == "sendmessage" && !message) {
    $(".submitbutton").removeClass('disabled');
    $("#warning").text("Message is Empty").removeClass("hidden").addClass("visible");
    return false;
  }
  const data = {action, message};
  request(data, (res)=>{
    $("#result").text(res.message);
    $("#info").removeClass("hidden").addClass("visible");
    $(".submitbutton").removeClass('disabled');
  }, (e)=>{
    console.log(e.responseJSON.message);
    $("#warning").text(e.responseJSON.message).removeClass("hidden").addClass("visible");
    $(".submitbutton").removeClass('disabled');
  });
};

var request = function(data, callback, onerror) {
  $.ajax({
    type:          'POST',
    dataType:      'json',
    contentType:   'application/json',
    scriptCharset: 'utf-8',
    data:          JSON.stringify(data),
    url:           App.url
  })
  .done(function(res) {
    callback(res);
  })
  .fail(function(e) {
    onerror(e);
  });
};
var App = { url: location.origin + {{ .ApiPath }} };
