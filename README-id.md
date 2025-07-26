# Galus - Alat Live Reloading untuk Aplikasi Go

<img src="./public/banner.png">

[**English**](README.md) | [**Bahasa Indonesia**](README-id.md)

Galus adalah alat _live reloading_ untuk aplikasi Go, mirip dengan Air atau CompileDaemon. Alat ini memantau perubahan file di proyek Anda, secara otomatis mengompilasi ulang, dan memulai ulang aplikasi.

## Fitur

- **Live Reloading**: Secara otomatis mendeteksi perubahan pada jenis file tertentu (misalnya, `.go`) dan mengompilasi ulang serta memulai ulang aplikasi.
- **Dapat Dikonfigurasi via `.galus.toml`**: Sesuaikan direktori yang dipantau, ekstensi file, perintah build, dan argumen runtime menggunakan file konfigurasi TOML.
- **Output CLI Berwarna**: Menyediakan output dengan kode warna yang jelas untuk pengalaman pengguna yang lebih baik (misalnya, hijau untuk sukses, merah untuk error).
- **Perintah CLI yang Kaya**: Mendukung `init` untuk membuat file konfigurasi default, `version` untuk menampilkan versi aplikasi, dan `help` untuk detail penggunaan perintah.
- **Penghentian Aman (Graceful Shutdown)**: Menggunakan `SIGTERM` untuk menghentikan proses yang sedang berjalan dengan aman, dengan cadangan penghentian paksa jika diperlukan.
- **Validasi Konfigurasi**: Memastikan `build_cmd` dan `command_args` valid sebelum memulai proses _live reload_.
- **Lintas Platform**: Berfungsi di semua platform yang didukung oleh Go, dengan instalasi mudah melalui `go get`.

## Instalasi

1. Instal Galus menggunakan `go get`:

   ```bash
   go get github.com/aliftech/galus
   ```

2. Instal binary untuk membuat `galus` tersedia secara global:

   ```bash
   go install github.com/aliftech/galus
   ```

3. Pastikan `$GOPATH/bin` ada di `$PATH` Anda:
   ```bash
   export PATH=$PATH:$(go env GOPATH)/bin
   ```

## Penggunaan

1. **Inisialisasi file konfigurasi**: Di direktori proyek Anda, jalankan:

   ```bash
   galus init
   ```

   Ini akan membuat file konfigurasi `.galus.toml` dengan pengaturan default.

2. **Mulai live reloading**: Di direktori proyek Anda, jalankan:

   ```bash
   galus
   ```

   Galus akan memantau perubahan file (misalnya, file `.go`), mengompilasi ulang, dan memulai ulang aplikasi Anda.

3. **Cek versi**:

   ```bash
   galus version
   ```

4. **Lihat bantuan**:
   ```bash
   galus help
   ```

## Konfigurasi

File `.galus.toml` memungkinkan Anda untuk menyesuaikan perilaku Galus. Contoh konfigurasi:

```toml
root_dir = "."
tmp_dir = "tmp"
include_ext = ["go"]
exclude_dir = [".git", "vendor", "tmp"]
build_cmd = "go build -o ./tmp/main ."
binary_name = "tmp/main"
command_args = ["tmp/main"]
```

- `root_dir`: Direktori yang dipantau untuk perubahan.
- `tmp_dir`: Direktori sementara untuk binary yang dikompilasi.
- `include_ext`: Ekstensi file yang dipantau (misalnya, `["go", "html"]`).
- `exclude_dir`: Direktori yang diabaikan.
- `build_cmd`: Perintah untuk mengompilasi aplikasi.
- `binary_name`: Nama binary yang dikompilasi.
- `command_args`: Argumen untuk menjalankan binary.
