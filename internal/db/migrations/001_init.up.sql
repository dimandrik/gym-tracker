CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL
);

CREATE TABLE machines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    name TEXT NOT NULL,
    photo_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE workouts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    workout_date DATE NOT NULL
);

CREATE TABLE workout_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workout_id UUID NOT NULL REFERENCES workouts(id),
    machine_id UUID NOT NULL REFERENCES machines(id)
);

CREATE TABLE sets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workout_item_id UUID NOT NULL REFERENCES workout_items(id),
    set_number INT NOT NULL,
    weight_kg NUMERIC(5,1) NOT NULL,
    reps INT NOT NULL
);