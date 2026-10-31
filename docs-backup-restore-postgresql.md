# Dokumentasi Backup & Restore PostgreSQL (Docker)

**Logical Backup menggunakan `pg_dump` / `pg_restore`**

Dokumentasi ini menjelaskan **cara backup & restore PostgreSQL - TimescaleDB yang berjalan di dalam Docker container**

---

## 1. Ringkasan

- Backup dilakukan di **level database (logical)**, **bukan Docker volume**
- Tool utama:
  - `pg_dump` → backup
  - `pg_restore` → restore

- Aman dilakukan **saat database ramai transaksi** (IoT / ingest per detik)
- Tidak perlu mematikan database saat backup

**Catatan penting** Backup **TIDAK BOLEH** dilakukan dengan redirect STDOUT (`pg_dump > file`) untuk **custom format (`-F c`)** di Docker karena berisiko menghasilkan file **corrupt / invalid**.

Metode yang digunakan di dokumen ini:

- Backup ditulis **langsung ke file di dalam container** (`pg_dump -f`)
- File diverifikasi
- File disalin ke host dengan `docker cp`

---

## 2. Prasyarat

- PostgreSQL / TimescaleDB berjalan di Docker
- Container name: `emerti-db`
- Database name: `emerti_db`
- Database user: `user`
- Password tersedia (contoh: `password`)

---

## 3. Arsitektur

```text
Backend (Go + Gin + GORM)
        ↓
Docker Container
  emerti-db (PostgreSQL / TimescaleDB)
```

---

## 4. Struktur Folder di Host

```bash
./backup/
```

Folder ini digunakan untuk menyimpan file backup hasil copy dari container.

---

## 5. Prosedur Backup

### 5.1 Buat Backup di Dalam Container

Backup dibuat dalam format **custom archive** (`-F c`) dan disimpan langsung ke filesystem container.

```bash
docker exec emerti-db sh -c \
"PGPASSWORD=password pg_dump -U user -F c emerti_db \
 --no-acl --no-owner \
 -f /tmp/emerti_db_$(date +%F_%H-%M-%S).backup"
```

**Kenapa ini aman:**

- Menghindari corrupt binary akibat STDOUT
- Warning PostgreSQL tidak merusak file archive
- Stabil untuk Docker + TimescaleDB
- `--no-acl --no-owner` memastikan restore aman ke environment/user yang berbeda

---

### 5.2 Validasi Backup (WAJIB)

Cek nama file

```bash
docker exec emerti-db ls -lh /tmp
```

Ganti `<ganti-yang-sesuai-nama-file>` dengan nama file backup

```bash
docker exec emerti-db pg_restore --list /tmp/emerti_db_<ganti-yang-sesuai-nama-file>.backup
```

Jika berhasil:

- Akan muncul daftar schema, table, index, dll

Jika gagal:

- Muncul error `did not find magic string`
- Artinya backup **INVALID** dan tidak boleh digunakan

---

### 5.3 Salin Backup ke Host

```bash
mkdir -p ./backup

docker cp emerti-db:/tmp/emerti_db_<ganti-yang-sesuai-nama-file>.backup \
  ./backup/
```

---

### 5.4 Verifikasi di Host (Opsional)

```bash
ls -lh ./backup
```

Ukuran file normal:

- Puluhan MB hingga ratusan MB (tergantung data)

---

## 6. Restore Database

**PASTIKAN aplikasi TIDAK menulis ke database saat restore**

### 6.1 Recreate Database

```bash
docker exec -it emerti-db psql -U user -d postgres
```

```sql
DROP DATABASE IF EXISTS emerti_db;
CREATE DATABASE emerti_db;
```

---

### 6.2 Create TimescaleDB Extension

```bash
docker exec -it emerti-db psql -U user -d emerti_db -c "CREATE EXTENSION IF NOT EXISTS timescaledb;"
```

### 6.3 Restore dari File Backup

```bash
docker cp ./backup/emerti_db_2026-02-11_12-00-00.backup \
  emerti-db:/tmp/restore.backup
```

```bash
docker exec emerti-db pg_restore \
  -U user \
  -d emerti_db \
  --no-acl --no-owner \
  -v \
  /tmp/restore.backup
```

---

### 6.4 Verifikasi Restore ke Database Test (Sangat Dianjurkan)

Sebelum menggunakan backup di production, sebaiknya restore ke database test terlebih dahulu untuk memastikan data valid.

```bash
docker exec -it emerti-db psql -U user -d postgres -c "CREATE DATABASE emerti_db_test;"
docker exec -it emerti-db psql -U user -d emerti_db_test -c "CREATE EXTENSION IF NOT EXISTS timescaledb;"
```

```bash
docker exec emerti-db pg_restore \
  -U user \
  -d emerti_db_test \
  --no-acl --no-owner \
  -v \
  /tmp/restore.backup
```

Setelah verifikasi selesai, hapus database test:

```bash
docker exec -it emerti-db psql -U user -d postgres -c "DROP DATABASE IF EXISTS emerti_db_test;"
```

---

## 7. Restore Sebagian (Selective Restore)

```bash
# Restore satu table
docker exec emerti-db pg_restore -U user -d emerti_db -t sensor_data /tmp/restore.backup

# Schema saja
docker exec emerti-db pg_restore -U user -d emerti_db --schema-only /tmp/restore.backup

# Data saja
docker exec emerti-db pg_restore -U user -d emerti_db --data-only /tmp/restore.backup
```

---

## 8. Backup Otomatis (Cron)

**JANGAN gunakan redirect `>` untuk custom format**

### 8.1 Buat Script Backup

Buat script agar cron tetap rapi dan mudah diubah:

```bash
nano /usr/local/bin/emerti-backup.sh
```

Isi script:

```bash
#!/bin/bash
DATE=$(date +%F)
BACKUP_DIR=/var/www/emerti-backup

# Buat backup
docker exec emerti-db sh -c \
  "PGPASSWORD=password pg_dump -U user -F c emerti_db \
   --no-acl --no-owner \
   -f /tmp/emerti_db_${DATE}.backup"

# Copy ke host
docker cp emerti-db:/tmp/emerti_db_${DATE}.backup \
  ${BACKUP_DIR}/

# Hapus backup lebih dari 3 hari (simpan 4 backup terakhir)
find ${BACKUP_DIR} -type f -name "*.backup" -mtime +3 -delete
```

Simpan dan keluar: `Ctrl+O` → Enter → `Ctrl+X`

Beri izin eksekusi:

```bash
chmod +x /usr/local/bin/emerti-backup.sh
```

Pastikan folder backup ada:

```bash
mkdir -p /var/www/emerti-backup
```

### 8.2 Daftarkan ke Cron

```bash
crontab -e
```

Tambahkan baris berikut (jauh lebih pendek):

```
0 2 * * * /usr/local/bin/emerti-backup.sh
```

Simpan dan keluar: `Ctrl+O` → Enter → `Ctrl+X`

### 8.3 Verifikasi Cron Tersimpan

```bash
crontab -l
```

### 8.4 Test Script Manual

Sebelum nunggu jam 02:00, test dulu scriptnya:

```bash
/usr/local/bin/emerti-backup.sh
```

Cek hasilnya:

```bash
ls -lh /var/www/emerti-backup/
```

> **Penting:** Pastikan file di-copy ke host (`docker cp`) setelah backup. File yang hanya ada di dalam container akan hilang jika container dihapus.

---

## 9. Offsite Backup (Opsional tapi Sangat Dianjurkan)

```text
Server A (App + DB)
   │
   ├─ pg_dump
   │
   └─ scp / rsync
        ↓
Server B (Backup Storage)
```

```bash
scp ./backup/emerti_db_2026-02-11.backup \
  backupuser@backup-server:/data/postgres/
```

---

## 10. Retention Policy (WAJIB)

Backup **TIDAK boleh ditumpuk tanpa aturan**.

Kebijakan yang digunakan:

| Jenis | Disimpan |
| ----- | -------- |
| Daily | 4 hari   |

File backup ke-5 masuk → file backup hari pertama otomatis dihapus.

Retention sudah otomatis dijalankan di dalam script `emerti-backup.sh` (section 8.1):

```bash
find /var/www/emerti-backup -type f -name "*.backup" -mtime +3 -delete
```

> `-mtime +3` artinya hapus file yang **lebih dari 3 hari**, sehingga selalu tersisa **4 file terakhir**.

---

## 11. Catatan TimescaleDB

Saat backup, mungkin muncul warning:

```text
circular foreign-key constraints
hypertable / chunk / continuous_agg
```

Penjelasan:

- Ini **normal** pada TimescaleDB
- **BUKAN error**
- Backup tetap valid

---

## 12. Hal yang HARUS Dihindari

1. Backup Docker volume langsung
2. Restore tanpa test
3. AutoMigrate GORM saat restore
4. Simpan backup di disk yang sama
5. Backup custom format via STDOUT

---

## 13. Checklist

- [ ] Backup valid (`pg_restore --list`)
- [ ] File tersimpan di host (via `docker cp`)
- [ ] File terkirim ke server lain
- [ ] Retensi berjalan otomatis
- [ ] Restore pernah diuji ke database test

---

## 14. Conclusion

- Gunakan **pg_dump + pg_restore**
- Aman walau database IoT rame
- Jangan backup volume Docker
- Selalu validasi backup