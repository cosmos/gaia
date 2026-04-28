#!/bin/bash

coverage_profile="$1"
filtered_coverage_profile="$2"
exclusion_file=".coverageignore"

cp "$coverage_profile" "$filtered_coverage_profile"

while read -r pattern; do
  files_to_exclude=$(find . -type f -regex ".*$pattern")
  for file in $files_to_exclude; do
    relative_path=$(realpath --relative-to="." "$file")
    grep -v "$relative_path" "$filtered_coverage_profile" > temp_coverage.out && mv temp_coverage.out "$filtered_coverage_profile"
  done
done < "$exclusion_file"