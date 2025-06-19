package security

import (
    "RAAS/core/config"
    
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "encoding/hex"
    "fmt"
     // Replace with the actual import path
)

// EncryptData encrypts the provided data using AES encryption with the given key.
// It generates a random IV and appends it to the encrypted data for secure decryption later.
func EncryptData(data []byte) (string, error) {
    // Retrieve the secret key from the global config
    secretKey := getSecretKey()

    block, err := aes.NewCipher([]byte(secretKey))
    if err != nil {
        return "", err
    }

    // Generate a random IV (Initialization Vector)
    iv := make([]byte, aes.BlockSize)
    if _, err := rand.Read(iv); err != nil {
        return "", err
    }

    // Encrypt the data using the CFB mode
    stream := cipher.NewCFBEncrypter(block, iv)
    encrypted := make([]byte, len(data))
    stream.XORKeyStream(encrypted, data)

    // Append IV to encrypted data and return it as a hex string
    return hex.EncodeToString(append(iv, encrypted...)), nil
}

// DecryptData decrypts the provided encrypted data using AES with the given key.
// It extracts the IV from the encrypted data and uses it for decryption.
func DecryptData(encryptedData string) ([]byte, error) {
    // Retrieve the secret key from the global config
    secretKey := getSecretKey()

    // Decode the hex string into a byte slice
    encryptedBytes, err := hex.DecodeString(encryptedData)
    if err != nil {
        return nil, err
    }

    // Extract the IV (first BlockSize bytes) and the actual encrypted data (rest of the bytes)
    if len(encryptedBytes) < aes.BlockSize {
        return nil, fmt.Errorf("invalid encrypted data")
    }
    iv := encryptedBytes[:aes.BlockSize]
    encrypted := encryptedBytes[aes.BlockSize:]

    // Create the AES cipher block with the provided key
    block, err := aes.NewCipher([]byte(secretKey))
    if err != nil {
        return nil, err
    }

    // Decrypt the data using CFB mode
    stream := cipher.NewCFBDecrypter(block, iv)
    decrypted := make([]byte, len(encrypted))
    stream.XORKeyStream(decrypted, encrypted)

    return decrypted, nil
}

// getSecretKey retrieves the SecretKey from the global config.
// getSecretKey retrieves the SecretKey from the global config.
func getSecretKey() string {
    return config.Cfg.Project.SecretKey
}