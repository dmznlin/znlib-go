(function () {
    var terminal = new Terminal({
        screenKeys: true,
        useStyle: true,
        cursorBlink: true,
        fullscreenWin: true,
        maximizeWin: true,
        screenReaderMode: true,
        cols: 128,
    });

    terminal.open(document.getElementById("terminal"));
    // xterm 终端附加到 HTML 元素上

    var protocol = (location.protocol === "https:") ? "wss://" : "ws://";
    var url = protocol + location.host + "/terminal"
    var ws = new WebSocket(url);
    //创建WebSocket连接

    var attachAddon = new AttachAddon.AttachAddon(ws);
    //将 xterm 终端附加到 WebSocket 流，与后台进行交互

    var fitAddon = new FitAddon.FitAddon();
    terminal.loadAddon(fitAddon);
    //调整 xterm 终端的大小，以适合父元素的大小

    var webLinksAddon = new WebLinksAddon.WebLinksAddon();
    terminal.loadAddon(webLinksAddon);
    //允许在 xterm 终端会话中创建可点击的超链接

    var unicode11Addon = new Unicode11Addon.Unicode11Addon();
    terminal.loadAddon(unicode11Addon);
    //在 xterm 终端中支持 Unicode 11,用于表示各种文字、符号和表情等.

    var serializeAddon = new SerializeAddon.SerializeAddon();
    terminal.loadAddon(serializeAddon);
    //Serializes the terminal's buffer to a VT sequences or HTML

    const cmdTag = "\0\1\2\3\4\5"
    //命令前缀

    ws.onclose = function (event) {
        console.log(event);
        terminal.write('\r\n\n连接已断开(刷新页面重新登录)\n')
    };

    ws.onopen = function () {
        terminal.loadAddon(attachAddon);
        terminal._initialized = true;
        terminal.focus();

        setTimeout(function () {
            fitAddon.fit()
        });

        terminal.onResize(function (event) {
            var rows = event.rows;
            var cols = event.cols;
            ws.send(cmdTag + "resize" + cols + "," + (rows + 1));
        });

        terminal.onTitleChange(function (event) {
            console.log(event);
        });

        window.onresize = function () {
            fitAddon.fit();
        };

        ws.send(cmdTag + "conn"); //连接ssh
        ws.send("\n"); //若已连接,回车后刷新提示符
    };
})();
