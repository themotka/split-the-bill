-- Пользователи приложения
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    avatar_url TEXT,
    email TEXT UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);в

-- События (мероприятия, поездки и т.д.)
CREATE TABLE events (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    created_by INTEGER REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Участники событий
CREATE TABLE event_participants (
    id SERIAL PRIMARY KEY,
    event_id INTEGER REFERENCES events(id) ON DELETE CASCADE,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE
);

-- Траты (чеки), сделанные в рамках события
CREATE TABLE expenses (
    id SERIAL PRIMARY KEY,
    event_id INTEGER REFERENCES events(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    amount NUMERIC(10, 2) NOT NULL,
    paid_by INTEGER REFERENCES users(id),
    paid_at DATE DEFAULT CURRENT_DATE
);

-- Распределение долей в каждой трате
CREATE TABLE expense_shares (
    id SERIAL PRIMARY KEY,
    expense_id INTEGER REFERENCES expenses(id) ON DELETE CASCADE,
    user_id INTEGER REFERENCES users(id),
    share_amount NUMERIC(10, 2) NOT NULL
);

-- Учет задолженности между участниками
CREATE TABLE debts (
    id SERIAL PRIMARY KEY,
    event_id INTEGER REFERENCES events(id) ON DELETE CASCADE,
    from_user INTEGER REFERENCES users(id),
    to_user INTEGER REFERENCES users(id),
    amount NUMERIC(10, 2) NOT NULL,
    is_settled BOOLEAN DEFAULT FALSE
);

-- Запись фактических оплат долгов
CREATE TABLE payments (
    id SERIAL PRIMARY KEY,
    from_user INTEGER REFERENCES users(id),
    to_user INTEGER REFERENCES users(id),
    amount NUMERIC(10, 2) NOT NULL,
    paid_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    event_id INTEGER REFERENCES events(id)
);
