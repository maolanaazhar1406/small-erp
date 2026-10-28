# Search Feature — Walkthrough

## Ringkasan

Fitur pencarian `?q=` berhasil ditambahkan ke **13 endpoint List** pada seluruh menu ERP.

## Files yang Diubah (12 files)

| File | Perubahan |
|------|-----------|
| [repository/product_repo.go](file:///home/user/Documents/small-erp/internal/repository/product_repo.go) | [ListCategories](file:///home/user/Documents/small-erp/internal/repository/product_repo.go#21-51), [ListUnits](file:///home/user/Documents/small-erp/internal/repository/product_repo.go#91-121), [ListProducts](file:///home/user/Documents/small-erp/internal/repository/product_repo.go#161-195) → tambah param `q` |
| [repository/partner_repo.go](file:///home/user/Documents/small-erp/internal/repository/partner_repo.go) | [ListCustomers](file:///home/user/Documents/small-erp/internal/repository/partner_repo.go#21-53), [ListSuppliers](file:///home/user/Documents/small-erp/internal/repository/partner_repo.go#124-156) → tambah param `q` |
| [repository/purchase_repo.go](file:///home/user/Documents/small-erp/internal/repository/purchase_repo.go) | [List](file:///home/user/Documents/small-erp/internal/handler/purchase.go#119-129) (PO), [ListGoodsReceipts](file:///home/user/Documents/small-erp/internal/repository/purchase_repo.go#260-293) → tambah param `q` |
| [repository/sales_repo.go](file:///home/user/Documents/small-erp/internal/repository/sales_repo.go) | [List](file:///home/user/Documents/small-erp/internal/handler/purchase.go#119-129) (SO) → tambah param `q` |
| [repository/finance_repo.go](file:///home/user/Documents/small-erp/internal/repository/finance_repo.go) | [ListJournalEntries](file:///home/user/Documents/small-erp/internal/repository/finance_repo.go#102-136), [ListAP](file:///home/user/Documents/small-erp/internal/repository/finance_repo.go#220-256), [ListAR](file:///home/user/Documents/small-erp/internal/handler/finance.go#146-156), [ListPayments](file:///home/user/Documents/small-erp/internal/repository/finance_repo.go#307-347) → tambah param `q` |
| [repository/user_repo.go](file:///home/user/Documents/small-erp/internal/repository/user_repo.go) | [List](file:///home/user/Documents/small-erp/internal/handler/purchase.go#119-129) (Users) → tambah param `q` |
| [handler/product.go](file:///home/user/Documents/small-erp/internal/handler/product.go) | 3 handler List: baca `q` dari query string, teruskan ke repo |
| [handler/partner.go](file:///home/user/Documents/small-erp/internal/handler/partner.go) | 2 handler List: baca `q` |
| [handler/purchase.go](file:///home/user/Documents/small-erp/internal/handler/purchase.go) | 2 handler List: baca `q` |
| [handler/sales.go](file:///home/user/Documents/small-erp/internal/handler/sales.go) | 1 handler List: baca `q` |
| [handler/finance.go](file:///home/user/Documents/small-erp/internal/handler/finance.go) | 4 handler List: baca `q` |
| [handler/user.go](file:///home/user/Documents/small-erp/internal/handler/user.go) | 1 handler List: baca `q` |

## Cara Kerja

Setiap handler membaca parameter dari query string:
```go
q := strings.TrimSpace(r.URL.Query().Get("q"))
```

Query SQL menggunakan pola ILIKE yang sama pada `COUNT` dan `SELECT`:
```sql
WHERE ($1 = '' OR code  ILIKE '%' || $1 || '%'
               OR name  ILIKE '%' || $1 || '%')
```
- Jika `q` kosong → mengembalikan semua data (backward compatible)
- Jika `q` diisi → hanya data yang cocok yang dikembalikan
- Pagination (`total`) dihitung dengan filter yang sama → konsisten

## Kolom Pencarian per Menu

| Menu | Kolom yang dicari |
|------|------------------|
| Categories | [name](file:///home/user/Documents/small-erp/internal/repository/user_repo.go#20-33) |
| Units | [name](file:///home/user/Documents/small-erp/internal/repository/user_repo.go#20-33), `abbreviation` |
| Products | `code`, [name](file:///home/user/Documents/small-erp/internal/repository/user_repo.go#20-33) |
| Customers | `code`, [name](file:///home/user/Documents/small-erp/internal/repository/user_repo.go#20-33), `email`, `phone` |
| Suppliers | `code`, [name](file:///home/user/Documents/small-erp/internal/repository/user_repo.go#20-33), `email`, `phone` |
| Purchase Orders | `po_number`, `supplier name` |
| Goods Receipts | `gr_number`, `po_number` |
| Sales Orders | `so_number`, `customer name` |
| Journal Entries | `entry_number`, `description` |
| Accounts Payable | `supplier name`, `po_number` |
| Accounts Receivable | `customer name`, `so_number` |
| Payments | `payment_number`, `payment_type`, `reference number` |
| Users | `username`, `full_name`, `email` |

## Contoh Penggunaan

```http
GET /api/v1/products?q=pupuk
GET /api/v1/customers?q=toko
GET /api/v1/purchase-orders?q=PO-2026
GET /api/v1/payments?q=payable
GET /api/v1/products?q=pupuk&page=1&per_page=10
```

## Hasil Verifikasi

```
$ go build ./...
(tidak ada output — clean compile) ✅
```

Tidak ada regresi pada fungsionalitas yang sudah ada karena jika `q` tidak dikirim, `strings.TrimSpace("")` mengembalikan string kosong, dan kondisi `$1 = ''` pada SQL akan `TRUE` sehingga semua data dikembalikan.
