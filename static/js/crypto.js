class Crypto {
    static RandomBytes (count) {
        let random = new Uint8Array(count);
        crypto.getRandomValues(random);

        return random;
    }

    static async GetECDHComponents () {

    }

    static async CompleteECDHExchange (sPublic, cPrivate, salt, context) {

    }

    static async SymmetricEncrypt (aesKey, hmacKey, plain) {

    }

    static async SymmetricDecrypt (aesKey, hmacKey, cipher) {
        
    }

    static async GetAsymmetricPair () {
        let keys = await window.crypto.subtle.generateKey(
            { name:"ECDH", namedCurve:"P-521" },
            true,
            ["deriveKey", "deriveBits"]
        );
        let public = await window.crypto.subtle.exportKey(
            "raw",
            keys.publicKey
        );

        return { public:public, private:keys.privateKey };
    }

    static async GetSymmetricFromExchange (serverPublic, clientPrivate, salt, context) {        
        let shared = await window.crypto.subtle.deriveBits(
            {
                name:"ECDH",
                namedCurve:"P-521",
                public:serverKey 
            },
            privateKey,
            528
        );
        if ((new Uint8Array(shared, 0, shared.byteLength))[0] == 0)
            shared = shared.slice(1);
    
        let raw = await window.crypto.subtle.importKey(
            "raw",
            derivedBits,
            "HKDF",
            true,
            ["deriveKey", "deriveBits"]
        );

        let aesKey = await window.crypto.subtle.deriveKey(
            {
                name:"HKDF",
                hash:"SHA-256",
                salt:salt,
                info:context
            },
            raw,
            {
               name:"AES-GCM",
               length:256 
            },
            true,
            ["encrypt", "decrypt"]
        );
        let hmacKey = await window.crypto.subtle.deriveKey(
            {
                name:"HKDF",
                hash:"SHA-256",
                salt:salt,
                info:context
            },
            raw,
            {
               name:"HMAC",
               hash:"SHA-256",
               length:256 
            },
            true,
            ["encrypt", "decrypt"]
        );

        return {AES:aesKey, HMAC:hmacKey};
    }
}