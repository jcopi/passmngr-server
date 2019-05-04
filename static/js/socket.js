class SecureSocket {
    constructor () {
        this.websocket = null;
        this.openhandler = null;
        this.closehandler = null;
        this.errorhandler = null;
        this.messagehandler = null;

        this.handshakeInitiated = false;
        this.handshakeCompleted = false;
    }

    set onopen(fn) {
        this.openhandler = fn;
    }

    set onclose(fn) {
        this.closehandler = fn;
    }

    set onerror(fn) {
        this.errorhandler = fn;
    }

    set onmessage(fn) {
        this.messagehandler = fn;
    }

    Open () {
        // Debug
        // this.websocket = new WebSocket("ws://" + window.location.host + "/socket")
        this.websocket = new WebSocket("wss://" + window.location.host + "/socket")
        this.websocket.onopen((ev) => {
            this.internalOpenHandler()
        });
        this.websocket.onerror((ev) => {
            this.errorhandler(ev);
        });
        this.websocket.onclose((ev) => {
            this.closehandler(ev);
        });
        this.websocket.onmessage((ev) => {
            this.internalMessageHandler(ev)
        });
    }

    internalOpenHandler () {

    }

    internalMessageHandler () {
        if (this.handshakeInitiated && !this.handshakeCompleted) {
            // This message must be a valid ECDH response
        }
    }

}

class SecureChannel {
    constructor () {
        // Cryptographical Context
        this.saltByteCount = 32;
        this.currentECDHContext = null;
        this.currentPublicKey = null;
        this.currentPrivateKey = null;
        this.currentSymmetricKey = null;


        // Event Listeners
        this.openResolver = null;
        this.openRejecter = null;

        // Internal WebSocket
        this.socket = null;
        this.socketUpgradeAddress = "wss://" + window.location.host + "/socket";
    }

    Open () {
        return new Promise((resolve, reject) => {
            this.openResolver = resolve;
            this.openRejecter = reject;

            this.socket = new WebSocket(this.socketUpgradeAddress);
            this.socket.onopen = (ev) => { this.internalOpenHandler(ev); };
            this.socket.onmessage = (ev) => { this.internalMessageHandler(ev); };
            this.socket.onerror = (ev) => { this.internalErrorHandler(ev); };
            this.socket.onclose = (ev) => { this.internalCloseHandler(ev); };
        });
    }

    Send () {
        return new Promise((resolve, reject) => {

        });
    }

    internalMessageHandler (ev) {
        
    }

    internalErrorHandler (ev) {

    }

    internalCloseHandler (ev) {

    }

    internalOpenHandler (ev) {

    }

    async internalGenerateECDHHeader () {
        let keys = await this.internalGenerateKeys();
        let salt = this.internalGenerateSalt();

        this.publicKey  = keys.public;
        this.privateKey = keys.private;

        let handshake = new Uint8Array(keys.public.byteLength + salt.byteLength)
        handshake.set(new Uint8Array(salt, 0, salt.byteLength), 0);
        handshake.set(new Uint8Array(keys.public, 0, keys.public.byteLength), salt.byteLength);

        return handshake;
    }

    async internalGenerateKeys() {
        let keys = await window.crypto.subtle.generateKey(
            { name:"ECDH", namedCurve:"P-521" }, true, ["deriveKey", "deriveBits"]
        );
        let public = await window.crypto.subtle.exportKey(
            "raw",
            keys.publicKey
        );

        return { public:public, private:keys.privateKey };
    }

    internalGenerateSalt() {
        let result = new Uint8Array(this.saltByteCount);
        window.crypto.getRandomValues(result);

        return result;
    }
}