-- Set `back_button_support` feature flag as user updatable
UPDATE feature_flags SET user_updatable = true WHERE key = 'back_button_support';
