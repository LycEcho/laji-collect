layui.define(['jquery','laytpl'],function(exports) {

    $ = layui.$;
    var laytpl = layui.laytpl
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
    ,$.renderDiy = function(tplHtml,data,openObj)
    {
        laytpl(tplHtml).render(data, function(html){
            var a = {content:html}
            var newA = Object.assign(a,openObj)
            return layer.open(newA)
        })
    }
})