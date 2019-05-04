class Crypto {
    static async AsymmetricGetKeyPair () {
        let keys = await window.crypto.subtle.generateKey({name:"ECDH",namedCurve:"P-521"},true,["deriveKey", "deriveBits"]);
        let public = await window.crypto.subtle.exportKey("raw",keys.publicKey);
    
        return {public:public, private:keys.privateKey};
    }

    static async AsymmetricImportPublicKey (raw) {
        return await window.crypto.subtle.importKey("raw",raw,{name:"ECDH",namedCurve:"P-521"},true,[]);
    }

    static async AsymmetricComputeSharedSecret (public, private) {
        let shared = await window.crypto.subtle.deriveBits({name:"ECDH",namedCurve:"P-521",public:public},private,528);
        if ((new Uint8Array(shared, 0, shared.byteLength))[0] == 0) {
            shared = shared.slice(1);
        }
    
        return shared;
    }

    static async SymmetricKeysFromSharedSecret (bits, salt, info) {
        let raw = await window.crypto.subtle.importKey("raw",bits,"HKDF",true,["deriveKey", "deriveBits"]);
    
        let aes = await window.crypto.subtle.deriveKey({name:"HKDF",hash:"SHA-256",salt:salt,info:info},raw,{name:"AES-GCM",length:256},false,["encrypt", "decrypt"]);
        let hmac = await window.crypto.subtle.deriveKey({name:"HKDF",hash:"SHA-256",salt:salt,info:info},raw,{name:"HMAC",hash:"SHA-256",length:256},false,["sign", "verify"]);
        
        return {aes:aes, hmac:hmac};
    }

    static async SymmetricEncrypt (aeskey, hmackey, data) {
        let iv = Crypto.GetSalt(StandardGCMNonceLength);
        let ct = await crypto.subtle.encrypt({name:"AES-GCM", iv:iv, tagLength:128},aeskey,data);

        let cipher = new Uint8Array(data.byteLength + iv.byteLength);
        cipher.set(new Uint8Array(iv), 0);
        cipher.set(new Uint8Array(ct), iv.byteLength);

        let signature = await crypto.subtle.sign("HMAC",hmackey,message);
        
        let signed = new Uint8Array(cipher.byteLength + signature.byteLength);
        signed.set(new Uint8Array(signature), 0);
        signed.set(new Uint8Array(message), signature.byteLength);

        return signed.buffer;
    }

    static async SymmetricDecrypt (aeskey, hmackey, data) {

    }

    static GetSalt (byteLength) {
        let random = new Uint8Array(byteLength);
        window.crypto.getRandomValues(random);

        return random;
    }
};

Crypto.StandardGCMNonceLength = 16;
Crypto.StandardHashLength     = 32;