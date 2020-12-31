layui.define(['jquery'],function(exports) {

    $ = layui.$;
    $.postJson = function(data){
        $.ajax({
            url:data.url
            ,data: JSON.stringify(data.data)
            ,contentType: "application/json; charset=utf-8"
            ,dataType: 'json'
            ,type:"POST"
            ,success: function(res) {
                data.success(res)
            }
            , error: function (err) {
                data.error(err)
            }
        });
    }
})