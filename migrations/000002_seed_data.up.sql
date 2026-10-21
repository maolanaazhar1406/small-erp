-- Seed Data for Small ERP System
-- Default roles, permissions, admin user, COA, and demo data

-- ============================================================
-- ROLES
-- ============================================================
INSERT INTO roles (name, description) VALUES
    ('admin', 'System Administrator with full access'),
    ('manager', 'Manager with read/write access to most modules'),
    ('staff', 'Staff with limited access')
ON CONFLICT (name) DO NOTHING;

-- ============================================================
-- PERMISSIONS
-- ============================================================
INSERT INTO permissions (name, description) VALUES
    ('users:read', 'View users'),
    ('users:create', 'Create users'),
    ('users:update', 'Update users'),
    ('users:delete', 'Delete users'),
    ('roles:read', 'View roles'),
    ('roles:create', 'Create roles'),
    ('roles:update', 'Update roles'),
    ('roles:delete', 'Delete roles'),
    ('products:read', 'View products'),
    ('products:create', 'Create products'),
    ('products:update', 'Update products'),
    ('products:delete', 'Delete products'),
    ('customers:read', 'View customers'),
    ('customers:create', 'Create customers'),
    ('customers:update', 'Update customers'),
    ('customers:delete', 'Delete customers'),
    ('suppliers:read', 'View suppliers'),
    ('suppliers:create', 'Create suppliers'),
    ('suppliers:update', 'Update suppliers'),
    ('suppliers:delete', 'Delete suppliers'),
    ('purchases:read', 'View purchase orders'),
    ('purchases:create', 'Create purchase orders'),
    ('purchases:update', 'Update purchase orders'),
    ('inventory:read', 'View inventory'),
    ('inventory:create', 'Create inventory adjustments'),
    ('sales:read', 'View sales orders'),
    ('sales:create', 'Create sales orders'),
    ('sales:update', 'Update sales orders'),
    ('finance:read', 'View financial data'),
    ('finance:create', 'Create financial entries'),
    ('finance:update', 'Update financial data'),
    ('reports:read', 'View reports')
ON CONFLICT (name) DO NOTHING;

-- ============================================================
-- ROLE-PERMISSION ASSIGNMENTS
-- ============================================================

-- Admin gets all permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r CROSS JOIN permissions p
WHERE r.name = 'admin'
ON CONFLICT DO NOTHING;

-- Manager gets most permissions except user/role management
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r CROSS JOIN permissions p
WHERE r.name = 'manager'
  AND p.name NOT IN ('users:create', 'users:update', 'users:delete', 'roles:create', 'roles:update', 'roles:delete')
ON CONFLICT DO NOTHING;

-- Staff gets read permissions + create sales/purchases
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r CROSS JOIN permissions p
WHERE r.name = 'staff'
  AND (p.name LIKE '%:read' OR p.name IN ('sales:create', 'purchases:create', 'inventory:create'))
ON CONFLICT DO NOTHING;

-- ============================================================
-- DEFAULT ADMIN USER (password: admin123)
-- ============================================================
INSERT INTO users (username, email, password_hash, full_name, role_id, is_active)
SELECT 'admin', 'admin@erp.local',
    '$2a$10$QYGrCZWxIJW6d84FXNe5Ae3us5phpCQi/4666XaavK0qP9aD7cfqO',
    'System Administrator',
    r.id, TRUE
FROM roles r WHERE r.name = 'admin'
ON CONFLICT (username) DO NOTHING;

-- Demo manager user (password: manager123)
INSERT INTO users (username, email, password_hash, full_name, role_id, is_active)
SELECT 'manager', 'manager@erp.local',
    '$2a$10$0NA.QTJDVgCmmPsd3pr14.fJoEDsgjOlAkaiBQisZhx8R4ijrKx2u',
    'Demo Manager',
    r.id, TRUE
FROM roles r WHERE r.name = 'manager'
ON CONFLICT (username) DO NOTHING;

-- Demo staff user (password: staff123)
INSERT INTO users (username, email, password_hash, full_name, role_id, is_active)
SELECT 'staff', 'staff@erp.local',
    '$2a$10$b4ITk2jpt2JnG6MrnzwcZuG0rdQYKWt0B6AlctYNZ4XBS7p7ltHZu',
    'Demo Staff',
    r.id, TRUE
FROM roles r WHERE r.name = 'staff'
ON CONFLICT (username) DO NOTHING;

-- ============================================================
-- CHART OF ACCOUNTS
-- ============================================================
INSERT INTO chart_of_accounts (code, name, account_type) VALUES
    -- Assets
    ('1000', 'Assets', 'asset'),
    ('1100', 'Cash', 'asset'),
    ('1200', 'Accounts Receivable', 'asset'),
    ('1300', 'Inventory', 'asset'),
    -- Liabilities
    ('2000', 'Liabilities', 'liability'),
    ('2100', 'Accounts Payable', 'liability'),
    -- Equity
    ('3000', 'Equity', 'equity'),
    ('3100', 'Owner Equity', 'equity'),
    ('3200', 'Retained Earnings', 'equity'),
    -- Revenue
    ('4000', 'Revenue', 'revenue'),
    ('4100', 'Sales Revenue', 'revenue'),
    -- Expenses
    ('5000', 'Expenses', 'expense'),
    ('5100', 'Cost of Goods Sold', 'expense'),
    ('5200', 'Operating Expenses', 'expense')
ON CONFLICT (code) DO NOTHING;

-- Set parent relationships
UPDATE chart_of_accounts SET parent_id = (SELECT id FROM chart_of_accounts WHERE code = '1000') WHERE code IN ('1100', '1200', '1300');
UPDATE chart_of_accounts SET parent_id = (SELECT id FROM chart_of_accounts WHERE code = '2000') WHERE code = '2100';
UPDATE chart_of_accounts SET parent_id = (SELECT id FROM chart_of_accounts WHERE code = '3000') WHERE code IN ('3100', '3200');
UPDATE chart_of_accounts SET parent_id = (SELECT id FROM chart_of_accounts WHERE code = '4000') WHERE code = '4100';
UPDATE chart_of_accounts SET parent_id = (SELECT id FROM chart_of_accounts WHERE code = '5000') WHERE code IN ('5100', '5200');

-- ============================================================
-- DEMO MASTER DATA
-- ============================================================

-- Units
INSERT INTO units (name, abbreviation) VALUES
    ('Piece', 'pcs'),
    ('Kilogram', 'kg'),
    ('Liter', 'ltr'),
    ('Meter', 'm'),
    ('Box', 'box')
ON CONFLICT (name) DO NOTHING;

-- Product Categories
INSERT INTO product_categories (name, description) VALUES
    ('Electronics', 'Electronic devices and components'),
    ('Office Supplies', 'Office equipment and stationery'),
    ('Raw Materials', 'Raw production materials'),
    ('Finished Goods', 'Completed products for sale')
ON CONFLICT (name) DO NOTHING;

-- Products
INSERT INTO products (code, name, description, category_id, unit_id, price, cost) VALUES
    ('PRD-001', 'Laptop Computer', 'Business laptop 14 inch',
        (SELECT id FROM product_categories WHERE name = 'Electronics'),
        (SELECT id FROM units WHERE abbreviation = 'pcs'),
        15000000, 12000000),
    ('PRD-002', 'Wireless Mouse', 'Ergonomic wireless mouse',
        (SELECT id FROM product_categories WHERE name = 'Electronics'),
        (SELECT id FROM units WHERE abbreviation = 'pcs'),
        250000, 150000),
    ('PRD-003', 'A4 Paper', 'A4 copy paper 80gsm',
        (SELECT id FROM product_categories WHERE name = 'Office Supplies'),
        (SELECT id FROM units WHERE abbreviation = 'box'),
        55000, 40000),
    ('PRD-004', 'USB Cable', 'USB-C cable 1 meter',
        (SELECT id FROM product_categories WHERE name = 'Electronics'),
        (SELECT id FROM units WHERE abbreviation = 'pcs'),
        75000, 35000),
    ('PRD-005', 'Printer Ink', 'Black printer ink cartridge',
        (SELECT id FROM product_categories WHERE name = 'Office Supplies'),
        (SELECT id FROM units WHERE abbreviation = 'pcs'),
        350000, 200000)
ON CONFLICT (code) DO NOTHING;

-- Initialize inventory for all products
INSERT INTO inventory (product_id, quantity)
SELECT id, 0 FROM products
ON CONFLICT (product_id) DO NOTHING;

-- Customers
INSERT INTO customers (code, name, email, phone, address) VALUES
    ('CUS-001', 'PT Maju Bersama', 'info@majubersama.co.id', '021-12345678', 'Jl. Sudirman No. 100, Jakarta'),
    ('CUS-002', 'CV Teknologi Nusantara', 'order@teknusa.co.id', '021-87654321', 'Jl. Gatot Subroto No. 50, Jakarta'),
    ('CUS-003', 'UD Sejahtera', 'sales@sejahtera.com', '031-11223344', 'Jl. Basuki Rahmat No. 25, Surabaya')
ON CONFLICT (code) DO NOTHING;

-- Suppliers
INSERT INTO suppliers (code, name, email, phone, address) VALUES
    ('SUP-001', 'PT Distributor Utama', 'procurement@distutama.co.id', '021-99887766', 'Jl. Industri No. 10, Tangerang'),
    ('SUP-002', 'CV Sumber Makmur', 'order@sumbermakmur.co.id', '022-55443322', 'Jl. Asia Afrika No. 75, Bandung'),
    ('SUP-003', 'PT Global Supply', 'info@globalsupply.co.id', '021-66778899', 'Jl. Raya Bogor KM 30, Depok')
ON CONFLICT (code) DO NOTHING;
