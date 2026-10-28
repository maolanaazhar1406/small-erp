Saya ingin menambahkan fitur **search atau pencarian data** pada seluruh menu aplikasi ERP yang membutuhkan fitur tersebut.

### Ketentuan Implementasi

1. **Identifikasi seluruh menu yang membutuhkan fitur pencarian**, seperti Products, Customers, Purchase Orders, Sales Orders, Suppliers, Payments, dan menu lainnya yang memiliki daftar data.

2. **Tambahkan parameter pencarian pada setiap handler list di backend.** Ambil nilai parameter `q` dari query string menggunakan kode berikut:

   ```go
   q := strings.TrimSpace(r.URL.Query().Get("q"))
   ```

   Pastikan package `strings` sudah diimpor.

3. **Terapkan filter pencarian pada query SQL.** Untuk setiap handler list, tambahkan filter yang sesuai dengan kolom pada tabel terkait. Contohnya:

   ```sql
   WHERE (
       $1 = ''
       OR code ILIKE '%' || $1 || '%'
       OR name ILIKE '%' || $1 || '%'
   )
   ```

   Sesuaikan kolom pencarian dengan struktur data masing-masing menu. Jangan menggunakan kolom `code` atau `name` jika kolom tersebut tidak tersedia pada tabel terkait.

4. **Samakan filter pada query data dan query count.** Pastikan kedua query menggunakan parameter pencarian dan kondisi filter yang sama agar jumlah data, hasil pencarian, serta pagination tetap konsisten.

5. **Integrasikan parameter pencarian dengan frontend.** Frontend harus mengirimkan kata kunci pencarian melalui query parameter `q`, misalnya:

   ```http
   GET /api/v1/products?q=pupuk
   ```

   Backend harus membaca dan memproses parameter tersebut agar data yang dikembalikan benar-benar tersaring sesuai kata kunci.

6. **Pertahankan kompatibilitas dengan fitur yang sudah ada.** Jika parameter `q` kosong atau tidak dikirimkan, endpoint harus tetap mengembalikan seluruh data sesuai aturan pagination dan filter lain yang sudah berlaku.

### Ketentuan Penting

- Jangan hanya mengimplementasikan pencarian di frontend. Filter harus diterapkan di backend melalui query SQL.
- Pastikan semua handler list yang membutuhkan pencarian sudah mendukung parameter `q`.
- Pertahankan fungsi pagination, sorting, filtering, serta format response API yang sudah ada.
- Gunakan parameterized query untuk menjaga keamanan dari SQL injection.
- Hindari duplikasi kondisi pencarian yang tidak konsisten antara query data dan query count.
- Jangan mengubah atau menghapus fungsi lain yang tidak berkaitan dengan implementasi fitur pencarian.
- Lakukan pengujian pada setiap endpoint yang telah diperbarui untuk memastikan pencarian, pagination, dan pengambilan data tetap berjalan dengan benar.

### Hasil yang Diharapkan

Seluruh menu yang membutuhkan fitur pencarian dapat melakukan pencarian data berdasarkan kata kunci yang dimasukkan pengguna. Frontend mengirimkan parameter `q`, backend memprosesnya, dan API hanya mengembalikan data yang sesuai dengan kata kunci tersebut.

Pastikan implementasi dilakukan secara menyeluruh, bukan hanya pada satu atau dua menu, serta tidak menimbulkan regresi pada fitur yang sudah berjalan.