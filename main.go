package main

import (
    "context"
    "crypto/ecdsa"
    "fmt"
    "log"
    "math/big"
    "net/http"
    "net/url"
    "io/ioutil"
    "os"
    "time"

    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/crypto"
    "github.com/ethereum/go-ethereum/ethclient"
)

const (
    telegramToken = "token" // Token bot Telegram Anda
    chatID        = int64(id)                                // ID chat Telegram Anda dalam format int64
)

// Fungsi untuk mengirim pesan ke Telegram
func sendTelegramMessage(message string) {
    apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", telegramToken)
    values := url.Values{}
    values.Set("chat_id", fmt.Sprintf("%d", chatID)) // Konversi int64 ke string
    values.Set("text", message)

    resp, err := http.PostForm(apiURL, values)
    if err != nil {
        log.Println("ERROR: Gagal mengirim pesan ke Telegram:", err)
        return
    }
    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        log.Println("ERROR: Gagal membaca respons dari Telegram:", err)
        return
    }

    log.Println("INFO: Pesan Telegram terkirim:", string(body))
}

func main() {
    // Buat log file untuk mencatat error
    logFile, err := os.OpenFile("bot-error.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
    if err != nil {
        fmt.Println("ERROR: Gagal membuka file log:", err)
        log.Fatal("Gagal membuka file log:", err)
    } else {
        fmt.Println("INFO: Berhasil membuka file log")
    }
    defer logFile.Close()

    // Konfigurasi logger untuk mencetak ke file log
    log.SetOutput(logFile)

    for {
        // Koneksi ke provider (ETH Mainnet)
        fmt.Println("INFO: Memulai koneksi ke jaringan ETH Mainnet")
        client, err := ethclient.Dial("https://ethereum.blockpi.network/v1/rpc/public")
        if err != nil {
            log.Println("ERROR: Gagal menghubungkan ke jaringan:", err)
            fmt.Println("Gagal menghubungkan ke jaringan. Lihat bot-error.log untuk detailnya.")
            continue // Menggunakan continue untuk melanjutkan ke iterasi berikutnya jika koneksi gagal
        }
        fmt.Println("INFO: Berhasil terhubung ke jaringan ETH")

        // Generate private key dan address acak
        fmt.Println("INFO: Menghasilkan Private Key")
        privateKey, err := crypto.GenerateKey()
        if err != nil {
            log.Println("ERROR: Gagal membuat private key:", err)
            fmt.Println("Gagal membuat private key. Lihat bot-error.log untuk detailnya.")
            continue
        }
        privateKeyBytes := crypto.FromECDSA(privateKey)
        privateKeyHex := fmt.Sprintf("%x", privateKeyBytes)

        publicKey := privateKey.Public()
        publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
        if !ok {
            log.Println("ERROR: Gagal casting public key ke ECDSA")
            fmt.Println("Gagal memproses public key. Lihat bot-error.log untuk detailnya.")
            continue
        }
        address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()

        // Print address, private key, and balance
        fmt.Printf("INFO: Address: %s\n", address)
        fmt.Printf("INFO: Private Key: %s\n", privateKeyHex)

        // Cek saldo address
        accountAddress := common.HexToAddress(address)
        fmt.Println("INFO: Memeriksa saldo address:", address)
        balance, err := client.BalanceAt(context.Background(), accountAddress, nil)
        if err != nil {
            log.Println("ERROR: Gagal memeriksa saldo:", err)
            fmt.Println("Gagal memeriksa saldo. Lihat bot-error.log untuk detailnya.")
            continue
        }
        balanceInETH := new(big.Float).Quo(new(big.Float).SetInt(balance), big.NewFloat(1e18))
        fmt.Printf("INFO: Saldo: %f ETH\n", balanceInETH)

        // Kirim pemberitahuan jika saldo lebih besar dari 0.0000001 ETH
        threshold := big.NewFloat(0.0000001)
        if balanceInETH.Cmp(threshold) > 0 {
            message := fmt.Sprintf("🔑 Private Key: %s\n🏦 Address: %s\n💰 Balance: %f ETH", privateKeyHex, address, balanceInETH)
            sendTelegramMessage(message)
            log.Println("INFO: Mengirim pesan ke Telegram karena saldo melebihi batas 0.0000001 ETH")
        } else {
            log.Println("INFO: Saldo kurang dari 0.0000001 ETH, tidak mengirim pesan")
        }

        // Tunggu beberapa detik sebelum loop berikutnya
        fmt.Println("INFO: Cari tolen")
        time.Sleep(1 * time.Second)
    }
}
