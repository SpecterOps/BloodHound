-- Set `back_button_support` feature flag as user updatable
UPDATE feature_flags SET user_updatable = true WHERE key = 'back_button_support';

-- Specify the `back_button_support` feature flag is currently only for BHCE users
UPDATE feature_flags SET description = 'Enable users to quickly navigate between views in a wider range of scenarios by utilizing the browser navigation buttons. Currently for BloodHound Community Edition users only.' WHERE key = 'back_button_support';
