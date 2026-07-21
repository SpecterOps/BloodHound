#!/usr/bin/env bash
# DORA Metrics Trend Analysis Helper
# Generates reports for multiple time periods to analyze trends

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
STBERNARD="${STBERNARD_BIN:-stbernard}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Print usage
usage() {
    cat << EOF
DORA Metrics Trend Analysis

Generate DORA metrics reports for multiple time periods to analyze trends.

Usage:
  $(basename "$0") <command> [options]

Commands:
  quarters <year>           Generate reports for all quarters of a year
  months <year> <month>     Generate reports for each month in a range
  compare <start> <end>     Generate reports for two periods and compare

Examples:
  # Generate quarterly reports for 2024
  $(basename "$0") quarters 2024

  # Generate monthly reports for 2024
  $(basename "$0") months 2024

  # Compare Q1 vs Q2 2024
  $(basename "$0") compare 2024-01-01:2024-03-31 2024-04-01:2024-06-30

Output:
  Reports are saved to ./dora-reports/ as JSON files
  Summary is displayed in the terminal

EOF
    exit 1
}

# Generate quarterly reports for a year
generate_quarters() {
    local year=$1
    local output_dir="./dora-reports/quarters-${year}"
    
    echo -e "${BLUE}Generating quarterly reports for ${year}...${NC}"
    mkdir -p "$output_dir"
    
    # Q1
    echo -e "${YELLOW}Q1 ${year}${NC}"
    "$STBERNARD" dora report \
        -start "${year}-01-01" -end "${year}-03-31" \
        -format json -output "${output_dir}/q1.json"
    
    # Q2
    echo -e "${YELLOW}Q2 ${year}${NC}"
    "$STBERNARD" dora report \
        -start "${year}-04-01" -end "${year}-06-30" \
        -format json -output "${output_dir}/q2.json"
    
    # Q3
    echo -e "${YELLOW}Q3 ${year}${NC}"
    "$STBERNARD" dora report \
        -start "${year}-07-01" -end "${year}-09-30" \
        -format json -output "${output_dir}/q3.json"
    
    # Q4
    echo -e "${YELLOW}Q4 ${year}${NC}"
    "$STBERNARD" dora report \
        -start "${year}-10-01" -end "${year}-12-31" \
        -format json -output "${output_dir}/q4.json"
    
    echo -e "${GREEN}✅ Quarterly reports saved to ${output_dir}/${NC}"
    echo ""
    echo "To analyze trends, you can:"
    echo "  1. Compare deployment frequency across quarters"
    echo "  2. Track change failure rate improvements"
    echo "  3. See MTTR trends"
    echo "  4. Import JSON files into dashboards"
}

# Generate monthly reports for a year
generate_months() {
    local year=$1
    local output_dir="./dora-reports/months-${year}"
    
    echo -e "${BLUE}Generating monthly reports for ${year}...${NC}"
    mkdir -p "$output_dir"
    
    local months=("01" "02" "03" "04" "05" "06" "07" "08" "09" "10" "11" "12")
    local month_names=("Jan" "Feb" "Mar" "Apr" "May" "Jun" "Jul" "Aug" "Sep" "Oct" "Nov" "Dec")
    local days_in_month=("31" "28" "31" "30" "31" "30" "31" "31" "30" "31" "30" "31")
    
    # Check for leap year
    if (( year % 4 == 0 && ( year % 100 != 0 || year % 400 == 0 ) )); then
        days_in_month[1]="29"
    fi
    
    for i in {0..11}; do
        local month="${months[$i]}"
        local month_name="${month_names[$i]}"
        local last_day="${days_in_month[$i]}"
        
        echo -e "${YELLOW}${month_name} ${year}${NC}"
        "$STBERNARD" dora report \
            -start "${year}-${month}-01" -end "${year}-${month}-${last_day}" \
            -format json -output "${output_dir}/${month}.json"
    done
    
    echo -e "${GREEN}✅ Monthly reports saved to ${output_dir}/${NC}"
}

# Main command dispatcher
main() {
    if [[ $# -eq 0 ]]; then
        usage
    fi
    
    local command=$1
    shift
    
    case "$command" in
        quarters)
            if [[ $# -ne 1 ]]; then
                echo "Error: quarters command requires year argument"
                usage
            fi
            generate_quarters "$1"
            ;;
        months)
            if [[ $# -ne 1 ]]; then
                echo "Error: months command requires year argument"
                usage
            fi
            generate_months "$1"
            ;;
        *)
            echo "Error: Unknown command: $command"
            usage
            ;;
    esac
}

main "$@"
