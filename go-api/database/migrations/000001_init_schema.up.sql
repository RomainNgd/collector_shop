-- Baseline schema, generated from the GORM models as of the migration to
-- versioned migrations (previously applied via AutoMigrate in production).

CREATE TABLE categories (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    name VARCHAR(120) NOT NULL,
    description VARCHAR(1000) NOT NULL
);

CREATE INDEX idx_categories_deleted_at ON categories USING btree (deleted_at);
CREATE UNIQUE INDEX idx_categories_name ON categories USING btree (name);

CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    email TEXT NOT NULL,
    password TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'ROLE_USER'
);

CREATE INDEX idx_users_deleted_at ON users USING btree (deleted_at);
CREATE UNIQUE INDEX idx_users_email ON users USING btree (email);

CREATE TABLE products (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    name VARCHAR(120) NOT NULL,
    description VARCHAR(1000) NOT NULL,
    image VARCHAR(255),
    price NUMERIC NOT NULL,
    stock BIGINT NOT NULL DEFAULT 1,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    seller_id BIGINT,
    promotion_type VARCHAR(20),
    promotion_value NUMERIC,
    promotion_active BOOLEAN NOT NULL DEFAULT FALSE,
    category_id BIGINT NOT NULL,
    CONSTRAINT fk_products_seller FOREIGN KEY (seller_id) REFERENCES users(id) ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_products_category FOREIGN KEY (category_id) REFERENCES categories(id) ON UPDATE CASCADE ON DELETE RESTRICT
);

CREATE INDEX idx_products_deleted_at ON products USING btree (deleted_at);
CREATE INDEX idx_products_seller_id ON products USING btree (seller_id);
CREATE INDEX idx_products_category_id ON products USING btree (category_id);
CREATE INDEX idx_products_is_active ON products USING btree (is_active);

CREATE TABLE promotions (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    name VARCHAR(120) NOT NULL,
    description VARCHAR(1000),
    type VARCHAR(20) NOT NULL,
    value NUMERIC NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    applies_to_all BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX idx_promotions_deleted_at ON promotions USING btree (deleted_at);
CREATE INDEX idx_promotions_type ON promotions USING btree (type);

CREATE TABLE product_promotions (
    promotion_id BIGINT NOT NULL,
    product_id BIGINT NOT NULL,
    PRIMARY KEY (promotion_id, product_id),
    CONSTRAINT fk_product_promotions_promotion FOREIGN KEY (promotion_id) REFERENCES promotions(id) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_product_promotions_product FOREIGN KEY (product_id) REFERENCES products(id) ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE orders (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    user_id BIGINT NOT NULL,
    status VARCHAR(40) NOT NULL DEFAULT 'awaiting_payment',
    currency VARCHAR(3) NOT NULL DEFAULT 'EUR',
    item_count BIGINT NOT NULL,
    subtotal NUMERIC NOT NULL,
    discount_total NUMERIC NOT NULL,
    total NUMERIC NOT NULL,
    payment_provider VARCHAR(40),
    payment_status VARCHAR(40) DEFAULT 'pending',
    paid_at TIMESTAMPTZ,
    stripe_checkout_session_id VARCHAR(255),
    stripe_checkout_session_status VARCHAR(40),
    stripe_payment_intent_id VARCHAR(255),
    stripe_checkout_expires_at TIMESTAMPTZ,
    CONSTRAINT fk_orders_user FOREIGN KEY (user_id) REFERENCES users(id) ON UPDATE CASCADE ON DELETE RESTRICT
);

CREATE INDEX idx_orders_deleted_at ON orders USING btree (deleted_at);
CREATE INDEX idx_orders_user_id ON orders USING btree (user_id);
CREATE INDEX idx_orders_status ON orders USING btree (status);
CREATE INDEX idx_orders_payment_status ON orders USING btree (payment_status);
CREATE INDEX idx_orders_stripe_checkout_session_id ON orders USING btree (stripe_checkout_session_id);

CREATE TABLE order_items (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    order_id BIGINT NOT NULL,
    product_id BIGINT NOT NULL,
    seller_id BIGINT,
    seller_email VARCHAR(255),
    product_name VARCHAR(120) NOT NULL,
    product_description VARCHAR(1000) NOT NULL,
    product_image VARCHAR(255),
    category_name VARCHAR(120),
    quantity BIGINT NOT NULL,
    unit_base_price NUMERIC NOT NULL,
    unit_price NUMERIC NOT NULL,
    unit_discount NUMERIC NOT NULL,
    line_base_total NUMERIC NOT NULL,
    line_discount_total NUMERIC NOT NULL,
    line_total NUMERIC NOT NULL,
    promotion_id BIGINT,
    promotion_name VARCHAR(120),
    promotion_type VARCHAR(20),
    promotion_value NUMERIC,
    promotion_applies_all BOOLEAN,
    CONSTRAINT fk_orders_items FOREIGN KEY (order_id) REFERENCES orders(id) ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE INDEX idx_order_items_deleted_at ON order_items USING btree (deleted_at);
CREATE INDEX idx_order_items_order_id ON order_items USING btree (order_id);
CREATE INDEX idx_order_items_product_id ON order_items USING btree (product_id);
CREATE INDEX idx_order_items_seller_id ON order_items USING btree (seller_id);
