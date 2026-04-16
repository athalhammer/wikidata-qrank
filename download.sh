#!/usr/bin/env bash
set -euo pipefail

# This script downloads Wikimedia pageview dumps for the previous 12 months (ending with the month before the current month),
# skipping files that are already present in the dumps/other/pageview_complete/ directory structure.

# Get current date
TODAY=$(date +%Y-%m-%d)
CUR_YEAR=$(date +%Y)
CUR_MONTH=$(date +%m)

# Calculate the first month to download (one year ago, next month)
START_YEAR=$(date -d "$CUR_YEAR-$CUR_MONTH-01 -1 year" +%Y)
START_MONTH=$(date -d "$CUR_YEAR-$CUR_MONTH-01 -1 year" +%m)

# Calculate the last month to download (previous month)
END_YEAR=$(date -d "$CUR_YEAR-$CUR_MONTH-01 -1 month" +%Y)
END_MONTH=$(date -d "$CUR_YEAR-$CUR_MONTH-01 -1 month" +%m)

# Function to compare year/month (returns 0 if $1/$2 <= $3/$4)
month_le() {
  [[ $1 -lt $3 ]] && return 0
  [[ $1 -eq $3 && $2 -le $4 ]] && return 0
  return 1
}

# Iterate from START_YEAR/START_MONTH to END_YEAR/END_MONTH
YEAR=$START_YEAR
MONTH=$((10#$START_MONTH))  # Always integer, never padded
while month_le $YEAR $MONTH $END_YEAR $END_MONTH; do
  MONTH_PAD=$(printf "%02d" "$MONTH")
  YM=$(printf "%04d-%02d" "$YEAR" "$MONTH")
  echo $YM
  DEST="dumps/other/pageview_complete/$YEAR/$YM"
  mkdir -p "$DEST"

  # Iterate to 31 days and wast some download attempts.
  for DAY_NUM in $(seq 1 31); do
    DAY_PAD=$(printf "%02d" "$DAY_NUM")
    YMD="${YEAR}${MONTH_PAD}${DAY_PAD}"
    URL="https://dumps.wikimedia.org/other/pageview_complete/${YEAR}/${YM}/pageviews-${YMD}-user.bz2"
    FILE="$DEST/pageviews-${YMD}-user.bz2"
    if [[ -f "$FILE" ]]; then
      echo "Already downloaded: $FILE"
      continue
    fi
    echo "Downloading $URL"
    wget -c -O "$FILE" "$URL" || echo "Warning: failed to download $URL"
    test -s "$FILE" || { echo "deleting empty $FILE"; rm -f "$FILE"; }
  done

  # Increment month
  if [[ $MONTH -eq 12 ]]; then
    MONTH=1
    YEAR=$((YEAR+1))
  else
    MONTH=$((MONTH+1))
  fi
done
