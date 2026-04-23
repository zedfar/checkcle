# CheckCle Internal Build — Deployment Guide

Panduan untuk menjalankan branch **internal** dari CheckCle yang telah dimodifikasi.
Build ini dari source code sehingga dapat dikustomisasi sepenuhnya.

> Untuk instalasi versi publik resmi (single container), lihat [README.md](README.md).

---

## Struktur Stack

| Container | Port | Deskripsi |
|---|---|---|
| `checkcle-modified-pocketbase` | `18090` | Backend API + Database (PocketBase v0.28.4, official binary) |
| `checkcle-modified-service-operation` | internal | Worker monitoring (HTTP, ping, DNS, TCP, SSL) |
| `checkcle-modified-frontend` | `18080` | Frontend React (Nginx) |

Port 5 digit dipilih agar tidak konflik jika stack publik juga jalan di server yang sama.

---

## Prasyarat

- Docker Engine 24+ dan Docker Compose v2
- Git (untuk clone repo)
- Port `18080` dan `18090` tersedia di server

---

## Deploy Pertama Kali (Fresh Server)

### 1. Clone dan masuk ke branch internal

```bash
git clone <repo-url>
cd checkcle
git checkout internal
```

### 2. Buat file konfigurasi

```bash
cp .env.build.example .env.build
```

Edit `.env.build`:

```env
# Port akses
FRONTEND_PORT=18080
PB_PORT=18090

# Akun admin — dibuat otomatis saat pertama kali deploy
# Juga dipakai oleh service-operation untuk auth ke PocketBase API
PB_ADMIN_EMAIL=admin@domain.com
PB_ADMIN_PASSWORD=PasswordKuat!123
PB_ADMIN_FULL_NAME=Administrator
PB_ADMIN_USERNAME=admin
```

### 3. Build dan jalankan

```bash
docker compose -f docker-compose.build.yml --env-file .env.build up -d --build
```

Proses ini akan otomatis:
1. Build semua image dari source code
2. Jalankan semua migrasi database (schema + seed data)
3. Buat akun admin sesuai env
4. Seed default notification templates (service + server)
5. Jalankan service-operation dengan auth ke PocketBase

Akses setelah selesai:
- **Frontend:** `http://server-ip:18080`
- **PocketBase Admin Dashboard:** `http://server-ip:18090/_/`

Login dengan `PB_ADMIN_EMAIL` dan `PB_ADMIN_PASSWORD` yang di-set di `.env.build`.

---

## Update / Rebuild (Data Tidak Hilang)

**Data aman selama tidak ada `-v` di perintah `down`.**

Data tersimpan di Docker named volume `checkcle-modified_pb_data` — terpisah dari container.
Rebuild image apapun (termasuk pocketbase) **tidak menghapus data**.

### Update kode dan deploy ulang

```bash
git pull
docker compose -f docker-compose.build.yml --env-file .env.build up -d --build
```

### Paksa rebuild tanpa Docker cache (jika ada perubahan Go/dependency)

```bash
docker compose -f docker-compose.build.yml --env-file .env.build build --no-cache
docker compose -f docker-compose.build.yml --env-file .env.build up -d
```

### Rebuild satu service saja

```bash
# Hanya frontend (misal ada perubahan UI)
docker compose -f docker-compose.build.yml --env-file .env.build up -d --build frontend

# Hanya service-operation (misal ada perubahan Go)
docker compose -f docker-compose.build.yml --env-file .env.build up -d --build service-operation

# Hanya pocketbase (misal ada migration baru)
docker compose -f docker-compose.build.yml --env-file .env.build up -d --build pocketbase
```

---

## Persistensi Data

Data di volume **tidak hilang** kecuali dihapus secara eksplisit.

| Perintah | Data |
|---|---|
| `up -d --build` | ✅ Aman |
| `down` (tanpa `-v`) | ✅ Aman |
| `restart <container>` | ✅ Aman |
| `down -v` | ❌ **Data hilang** |
| `docker volume rm checkcle-modified_pb_data` | ❌ **Data hilang** |

> **Kenapa PocketBase aman di-rebuild?**
> Container PocketBase hanya menjalankan binary. Data SQLite ada di volume `/app/pb_data` yang di-mount dari host. Rebuild image tidak menyentuh volume.

---

## Perintah Berguna

```bash
# Status semua container
docker compose -f docker-compose.build.yml ps

# Log real-time semua service
docker compose -f docker-compose.build.yml --env-file .env.build logs -f

# Log per service
docker compose -f docker-compose.build.yml logs -f pocketbase
docker compose -f docker-compose.build.yml logs -f service-operation
docker compose -f docker-compose.build.yml logs -f frontend

# Stop semua (data aman)
docker compose -f docker-compose.build.yml down

# Stop dan hapus semua data (HATI-HATI)
docker compose -f docker-compose.build.yml down -v
```

---

## Manajemen Akun Admin

```bash
# Buat atau update akun admin via CLI
docker exec checkcle-modified-pocketbase \
  ./pocketbase superuser upsert email@domain.com password-baru

# Ganti password
docker exec checkcle-modified-pocketbase \
  ./pocketbase superuser update email@domain.com password-baru
```

---

## Tambah Migrasi Database

Buat file di `server/pb_migrations/` dengan timestamp lebih besar dari yang ada:

```
server/pb_migrations/1900000010_nama_perubahan.js
```

Migrasi berjalan otomatis saat PocketBase start dan tidak dijalankan ulang (tracked di tabel `_migrations`).

Pola upsert untuk seed data:

```js
/// <reference path="../pb_data/types.d.ts" />
migrate((app) => {
  const col = app.findCollectionByNameOrId("nama_collection");
  let recs;
  try { recs = app.findRecordsByFilter(col.id, 'name = "Nama Record"', "", 1, 0); } catch(e) { recs = []; }
  const rec = (recs && recs.length > 0) ? recs[0] : new Record(col);
  rec.set("name", "Nama Record");
  // set fields lain...
  app.save(rec);
}, (app) => {});
```

---

## Fitur Modifikasi (vs Versi Publik)

### Tambahan Fitur
- **Telegram Thread ID** — kirim notifikasi ke topic/thread tertentu di Telegram Group (field `telegram_thread_id` di notification channel)
- **MOD badge** — indikator visual di sidebar dan halaman login bahwa ini build modifikasi
- **Port 5 digit** — frontend `18080`, backend `18090` agar bisa jalan berdampingan dengan stack publik

### Fix Arsitektur (root cause: semua masalah berasal dari DB kosong + collection rules)

- **PocketBase binary** — diganti dari custom multi-db fork ke official binary v0.28.4. Fork lama menyimpan SQLite di path yang tidak masuk volume → data hilang tiap rebuild
- **Service-operation auth** — service-operation kini autentikasi ke PocketBase menggunakan admin credentials (`PB_ADMIN_EMAIL`/`PB_ADMIN_PASSWORD`). Diperlukan karena migrasi `1772*` mengubah semua collection rules dari open ke require-auth
- **Notification field `status` vs `enabled`** — migrasi `1752921327` rename field `enabled` (boolean) ke `status` (select: `enabled`/`disabled`). Fix diterapkan di:
  - Frontend: `alertConfigService.ts` — mapping `status` → boolean `enabled` untuk UI
  - Service-operation: `notification/shared.go` — `isNotificationEnabled()` dan semua `http.Get` ke PocketBase kini pakai authenticated client
- **Date format** — `serviceService.ts` pakai `toISOString()` bukan `toLocaleString()` untuk field date PocketBase
- **Frontend port** — `pocketbase.ts` baca `VITE_PB_PORT` build arg, bukan hardcode `8090`
- **Seed otomatis** — migration `1900000000` seed default notification templates, `1900000001` seed akun admin dari env vars
