CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TYPE event_status AS ENUM ('DRAFT', 'PUBLISHED', 'CANCELLED', 'ENDED');
CREATE TYPE order_status AS ENUM ('PENDING', 'PAID', 'CANCELLED', 'TIMEOUT');
CREATE TYPE ticket_status AS ENUM ('UNUSED', 'USED');


CREATE TABLE IF NOT EXISTS users (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    username character varying(50) NOT NULL,
    email character varying(255) NOT NULL,
    password_hash text NOT NULL,
    role character varying(20) DEFAULT 'user'::character varying,
    profile_data jsonb DEFAULT '{}'::jsonb,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT users_pkey PRIMARY KEY (id),
    CONSTRAINT users_email_key UNIQUE (email),
    CONSTRAINT users_username_key UNIQUE (username),
    CONSTRAINT users_role_check CHECK (((role)::text = ANY ((ARRAY['user'::character varying, 'admin'::character varying])::text[])))
);


CREATE TABLE IF NOT EXISTS events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    location VARCHAR(255),
    banner_url VARCHAR(500),
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE NOT NULL,
    status event_status DEFAULT 'DRAFT',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT check_dates CHECK (end_time > start_time)
);


CREATE TABLE IF NOT EXISTS ticket_types (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    event_id UUID REFERENCES events(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL, -- Vd: VIP, GA, Early Bird
    price DECIMAL(10, 2) NOT NULL CHECK (price >= 0),
    initial_quantity INT NOT NULL CHECK (initial_quantity >= 0),
    remaining_quantity INT NOT NULL CHECK (remaining_quantity >= 0), -- Quan trọng: Không bao giờ được âm
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE IF NOT EXISTS orders (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id),
    total_amount DECIMAL(12, 2) NOT NULL DEFAULT 0,
    status order_status DEFAULT 'PENDING',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP, -- Dùng field này để quét đơn quá hạn (TTL)
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE IF NOT EXISTS order_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id UUID REFERENCES orders(id) ON DELETE CASCADE,
    ticket_type_id UUID REFERENCES ticket_types(id),
    quantity INT NOT NULL CHECK (quantity > 0),
    price DECIMAL(10, 2) NOT NULL -- Lưu giá tại thời điểm mua (Snapshot price)
);


CREATE TABLE IF NOT EXISTS tickets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id UUID REFERENCES orders(id),
    ticket_type_id UUID REFERENCES ticket_types(id),
    ticket_code VARCHAR(50) UNIQUE NOT NULL, -- Mã QR
    status ticket_status DEFAULT 'UNUSED',
    owner_name VARCHAR(100), -- Tên người đi xem
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';


CREATE TRIGGER update_users_modtime BEFORE UPDATE ON users FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
CREATE TRIGGER update_events_modtime BEFORE UPDATE ON events FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
CREATE TRIGGER update_ticket_types_modtime BEFORE UPDATE ON ticket_types FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
CREATE TRIGGER update_orders_modtime BEFORE UPDATE ON orders FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();


CREATE INDEX idx_events_slug ON events(slug);
CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_ticket_types_event_id ON ticket_types(event_id);

INSERT INTO users (username, email, password_hash, role) 
VALUES ('admin', 'admin@example.com', '$2a$10$WGkl8JLxQSRPXfnM8qxQi.XAJ4kX4p7N5nN5nN5nN5nN5nN5nN5nK', 'admin');

WITH new_event AS (
    INSERT INTO events (name, slug, location, start_time, end_time, status)
    VALUES ('Super Rock Concert 2026', 'rock-concert-2026', 'My Dinh Stadium', NOW() + INTERVAL '30 days', NOW() + INTERVAL '30 days 4 hours', 'PUBLISHED')
    RETURNING id
)

INSERT INTO ticket_types (event_id, name, price, initial_quantity, remaining_quantity)
SELECT id, 'VIP Ticket', 2000000, 100, 100 FROM new_event
UNION ALL
SELECT id, 'Standard Ticket', 500000, 1000, 1000 FROM new_event;