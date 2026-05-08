-- +goose Up
ALTER TABLE datapipe_status 
ADD COLUMN last_scheduled_analysis_run_at 

-- +goose Down
SELECT 'down SQL query';
