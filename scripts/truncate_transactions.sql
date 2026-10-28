-- ============================================================
-- TRUNCATE TRANSACTION DATA
-- Small ERP System
--
-- Script ini menghapus seluruh data transaksi tanpa menyentuh
-- data master (users, roles, products, customers, suppliers, dll.)
-- dan Chart of Accounts.
--
-- URUTAN PENTING: ikuti urutan CASCADE untuk menghindari
-- foreign key violations.
-- ============================================================

BEGIN;

-- 1. Payments (tidak ada FK ke tabel transaksi lain)
TRUNCATE TABLE payments RESTART IDENTITY CASCADE;

-- 2. Finance: journal lines → journal entries
TRUNCATE TABLE journal_entry_lines RESTART IDENTITY CASCADE;
TRUNCATE TABLE journal_entries     RESTART IDENTITY CASCADE;

-- 3. Finance: AP & AR
TRUNCATE TABLE accounts_payable    RESTART IDENTITY CASCADE;
TRUNCATE TABLE accounts_receivable RESTART IDENTITY CASCADE;

-- 4. Procurement: GR items → GR → PO items → PO
TRUNCATE TABLE goods_receipt_items  RESTART IDENTITY CASCADE;
TRUNCATE TABLE goods_receipts       RESTART IDENTITY CASCADE;
TRUNCATE TABLE purchase_order_items RESTART IDENTITY CASCADE;
TRUNCATE TABLE purchase_orders      RESTART IDENTITY CASCADE;

-- 5. Sales: SO items → SO
TRUNCATE TABLE sales_order_items RESTART IDENTITY CASCADE;
TRUNCATE TABLE sales_orders      RESTART IDENTITY CASCADE;

-- 6. Inventory: stock movements + reset quantity ke 0
TRUNCATE TABLE stock_movements RESTART IDENTITY CASCADE;

UPDATE inventory SET quantity = 0, last_updated = NOW();

COMMIT;

-- Verifikasi setelah truncate
SELECT 'payments'             AS tabel, COUNT(*) AS jumlah FROM payments
UNION ALL
SELECT 'journal_entries',      COUNT(*) FROM journal_entries
UNION ALL
SELECT 'journal_entry_lines',  COUNT(*) FROM journal_entry_lines
UNION ALL
SELECT 'accounts_payable',     COUNT(*) FROM accounts_payable
UNION ALL
SELECT 'accounts_receivable',  COUNT(*) FROM accounts_receivable
UNION ALL
SELECT 'goods_receipts',       COUNT(*) FROM goods_receipts
UNION ALL
SELECT 'purchase_orders',      COUNT(*) FROM purchase_orders
UNION ALL
SELECT 'sales_orders',         COUNT(*) FROM sales_orders
UNION ALL
SELECT 'stock_movements',      COUNT(*) FROM stock_movements
UNION ALL
SELECT 'inventory (sum qty)',   SUM(quantity) FROM inventory;
