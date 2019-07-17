let ws = new WebSocket("wss://" + window.location.host + "/socket")
ws.addEventListener("open", function (ev) {
    console.log("Socket Opened");
    ws.send("Message Data");
});
ws.addEventListener("message", function (ev) {
    console.log(ev.data);
})
ws.addEventListener("error", function (ev) {
    console.log(ev);
});