package main

import (
	"context"
	"fmt"
	"time"

	"github.com/PetaFlops-web/backend-shop-smbk/internal/pkg/notifier"
)

func main() {
	n := notifier.NewFonnte(notifier.FonnteConfig{Token: "2mhRFXqcqjN1vCSKmksh"})
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	err := n.SendReminder(ctx, notifier.Reminder{
		To:      "6289673302577",
		Message: "Halo, stok Kecap di toko kami mungkin sudah mulai menipis. Anda dapat membeli kembali Kecap dengan potongan harga 20% karena Anda sudah berbelanja 3 kali. Prediksi waktu pembelian ulang Anda: 2026-08-25. Sampai jumpa!",
	})
	if err != nil {
		fmt.Println("ERROR:", err)
		return
	}
	fmt.Println("SUCCESS: message sent via Go notifier")
}
