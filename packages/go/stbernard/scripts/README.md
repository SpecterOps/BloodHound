# DORA Metrics Analysis Scripts

Helper scripts for analyzing DORA metrics trends over time.

## Trend Analysis Helper

The `dora-trend-analysis.sh` script makes it easy to generate reports for multiple time periods.

### Prerequisites

1. Collect historical data once:
   ```bash
   # Collect 3 years of data
   stbernard dora collect -days 1095
   ```

2. Run the script to generate trend reports

### Usage

**Generate Quarterly Reports:**
```bash
./scripts/dora-trend-analysis.sh quarters 2024
```

Output:
```
./dora-reports/quarters-2024/
  ├── q1.json  # Jan-Mar 2024
  ├── q2.json  # Apr-Jun 2024
  ├── q3.json  # Jul-Sep 2024
  └── q4.json  # Oct-Dec 2024
```

**Generate Monthly Reports:**
```bash
./scripts/dora-trend-analysis.sh months 2024
```

Output:
```
./dora-reports/months-2024/
  ├── 01.json  # January
  ├── 02.json  # February
  ├── ...
  └── 12.json  # December
```

### Manual Report Generation

You can also generate reports manually for any custom period:

```bash
# Specific quarter
stbernard dora report -start 2024-01-01 -end 2024-03-31

# Specific month
stbernard dora report -start 2024-05-01 -end 2024-05-31

# Custom 90-day period
stbernard dora report -start 2024-01-15 -end 2024-04-15

# Last 90 days of 2023
stbernard dora report -start 2023-10-03 -end 2023-12-31

# Export as JSON
stbernard dora report -start 2024-01-01 -end 2024-03-31 \
  -format json -output q1-2024.json
```

## Analyzing Trends

Once you have JSON reports for multiple periods, you can:

### 1. Compare Deployment Frequency

```bash
# Extract deployment frequency from each quarter
jq '.dora_metrics.deployment_frequency' dora-reports/quarters-2024/q*.json
```

### 2. Track Change Failure Rate

```bash
# Show failure rate trend
jq '.dora_metrics.change_failure_rate' dora-reports/quarters-2024/q*.json
```

### 3. MTTR Improvements

```bash
# Show median restore time by quarter
jq '.dora_metrics.median_restore_time_hours' dora-reports/quarters-2024/q*.json
```

### 4. Import to Dashboard

The JSON files can be imported into:
- Grafana (via JSON API data source)
- Tableau
- Power BI
- Google Sheets (via script)
- Custom dashboards

## Example Workflow

**Goal:** Analyze if new practices adopted in Q2 improved metrics

```bash
# 1. Generate quarterly reports
./scripts/dora-trend-analysis.sh quarters 2024

# 2. Compare Q1 (before) vs Q3 (after)
echo "Q1 2024:"
jq '.dora_metrics | {
  deployments: .deployment_count,
  frequency: .deployment_frequency_per_day,
  failure_rate: .change_failure_rate,
  mttr_hours: .median_restore_time_hours
}' dora-reports/quarters-2024/q1.json

echo "Q3 2024:"
jq '.dora_metrics | {
  deployments: .deployment_count,
  frequency: .deployment_frequency_per_day,
  failure_rate: .change_failure_rate,
  mttr_hours: .median_restore_time_hours
}' dora-reports/quarters-2024/q3.json

# 3. Look for improvements:
# - Did deployment frequency increase?
# - Did failure rate decrease?
# - Did MTTR improve?
```

## Tips

**Keep Data Fresh:**
```bash
# Add to weekly cron job
0 0 * * 0 cd /path/to/repo && stbernard dora collect -days 1095
```

**Generate Executive Report:**
```bash
# Create year-over-year comparison
./scripts/dora-trend-analysis.sh quarters 2023
./scripts/dora-trend-analysis.sh quarters 2024

# Compare same quarter across years
diff <(jq '.dora_metrics' dora-reports/quarters-2023/q1.json) \
     <(jq '.dora_metrics' dora-reports/quarters-2024/q1.json)
```

**Track Process Changes:**
```bash
# Before adopting new CI/CD pipeline (Jan-Mar)
stbernard dora report -start 2024-01-01 -end 2024-03-31 -output before.json

# After adoption (Jul-Sep)
stbernard dora report -start 2024-07-01 -end 2024-09-30 -output after.json

# Compare metrics to measure impact
```
