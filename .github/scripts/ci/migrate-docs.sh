#!/bin/bash
# shellcheck disable=SC2016 # the backticks in the sed below are Markdown, not command substitution
set -e

# Ensure base directory exists (reference matches Sidebar.astro)
mkdir -p webpage/src/content/docs/reference

# Process all markdown files in docs/ recursively, and the guides in examples/guides/, which
# tfplugindocs leaves out (it renders guides only from templates/guides/)
find docs examples/guides -type f -name "*.md" | while read -r file; do
  # Determine relative path components
  rel_path="${file#docs/}"
  dir_part=$(dirname "$rel_path")
  base_part=$(basename "$rel_path")
  
  # Prefix with 'ignition_' if not index.md and not already prefixed
  target_name="$base_part"
  if [[ "$base_part" != "index.md" && "$base_part" != ignition_* ]]; then
    target_name="ignition_$base_part"
  fi
  
  # Determine final destination path in the reference/ directory; a guide joins the hand-written ones
  if [[ "$file" == examples/guides/* ]]; then
    target_name="$base_part"
    dest_path="webpage/src/content/docs/guides/$target_name"
  elif [[ "$dir_part" == "." ]]; then
    dest_path="webpage/src/content/docs/reference/$target_name"
  else
    dest_path="webpage/src/content/docs/reference/$dir_part/$target_name"
  fi
  
  # Create target subdirectory if it doesn't exist
  mkdir -p "$(dirname "$dest_path")"
  
  # Special title handling for index.md
  default_title="Sources"
  if [[ "$rel_path" == "index.md" ]]; then
    default_title="Ignition Provider"
  fi
  
  echo "Migrating $file to $dest_path"
  
  # Extract title and description using awk
  is_guide=0
  [[ "$file" == examples/guides/* ]] && is_guide=1
  awk -v def_title="$default_title" -v is_guide="$is_guide" ' 
    BEGIN { in_fm = 0; print_fm = 0; title = ""; desc = ""; }
    /^---$/ {
      if (in_fm == 0) { in_fm = 1; next; }
      if (in_fm == 1) { in_fm = 0; print_fm = 1; next; }
    }
    in_fm == 1 {
      if ($0 ~ /^page_title:/) {
        gsub(/^page_title: "/, "", $0);
        gsub(/"$/, "", $0);
        gsub(/ - ignition$/, "", $0);
        title = $0;
      }
      if ($0 ~ /^description: |-/) {
        next;
      }
      if ($0 ~ /^  /) {
        gsub(/^  /, "", $0);
        desc = (desc ? desc " " : "") $0;
      }
    }
    print_fm == 1 {
      print "---";
      print "title: \"" (title ? title : def_title) "\"";
      print "description: \"" desc "\"";
      if (def_title == "Ignition Provider") print "order: 0";
      if (is_guide) print "order: 4";
      print "---";
      print_fm = 2;
    }
    # the layout renders the title as the page h1, so the "# name (Resource)" heading tfplugindocs writes
    # and the blank line after it would repeat it
    print_fm == 2 && !dropped && /^# / { dropped = 1; skip_blank = 1; next; }
    print_fm == 2 && skip_blank && /^$/ { skip_blank = 0; next; }
    print_fm == 2 { skip_blank = 0; print; }
  ' "$file" |
    # tfplugindocs' import text names Terraform's command and docs; the examples run tofu import
    sed -e 's|^The \[`terraform import` command\](https://developer.hashicorp.com/terraform/cli/commands/import) can be used, for example:$|The [`tofu import` command](https://opentofu.org/docs/cli/commands/import/) (or `terraform import`) can be used, for example:|' \
      -e 's|^title: "ignition Provider"$|title: "Provider"|' > "$dest_path"
done
