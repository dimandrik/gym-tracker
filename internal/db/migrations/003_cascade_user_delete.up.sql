ALTER TABLE machines DROP CONSTRAINT machines_user_id_fkey;
ALTER TABLE machines ADD CONSTRAINT machines_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE workouts DROP CONSTRAINT workouts_user_id_fkey;
ALTER TABLE workouts ADD CONSTRAINT workouts_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE workout_items DROP CONSTRAINT workout_items_workout_id_fkey;
ALTER TABLE workout_items ADD CONSTRAINT workout_items_workout_id_fkey
    FOREIGN KEY (workout_id) REFERENCES workouts(id) ON DELETE CASCADE;