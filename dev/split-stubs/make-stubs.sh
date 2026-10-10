#!/usr/bin/env bash
# TEMPORARY - test data for the JAV scene split feature. Remove before the PR leaves draft.
#
# Makes tiny stand-in video files that ffprobe (and so XBVR) reports the same codec, width,
# height, frame rate and duration for as the real thing. Each "video" is a Matroska file with
# two black frames, the second one placed at the end of the duration. It keeps whatever name you
# give it (.mp4 included): ffprobe reads the contents, XBVR matches by name.
#
# Not copied: bit rate (comes out empty -> XBVR stores 0), file size, OsHash.
#
# Usage: make-stubs.sh <spec.tsv> <output folder>
#
# spec.tsv, one file per line, tab separated, as printed by the PowerShell snippet:
#   name  codec  width  height  avg_frame_rate  duration_seconds  bit_rate
#   84vrkm00139vrv1uhqe1.mp4  hevc  5760  2880  60/1  1390.433  50123456
# Lines with only a name (no codec) become small text files (.hsp, .srt, .funscript, ...).
# XBVR only records size and dates for those, the contents don't matter.

set -euo pipefail

spec="${1:?usage: make-stubs.sh <spec.tsv> <output folder>}"
out="${2:?usage: make-stubs.sh <spec.tsv> <output folder>}"
mkdir -p "$out"

command -v ffmpeg >/dev/null || { echo "ffmpeg missing: sudo apt-get update && sudo apt-get install -y ffmpeg"; exit 1; }

while IFS=$'\t' read -r name codec width height fps duration bitrate; do
	name="${name%$'\r'}"
	[ -z "$name" ] && continue
	case "$name" in \#*) continue ;; esac
	dest="$out/$name"

	if [ -z "${codec:-}" ]; then
		case "$name" in
		*.srt) printf '1\n00:00:01,000 --> 00:00:02,000\nstub\n' >"$dest" ;;
		*.funscript) printf '{"actions":[{"at":0,"pos":0}]}\n' >"$dest" ;;
		*) printf 'stub\n' >"$dest" ;;
		esac
		echo "text   $name"
		continue
	fi

	case "$codec" in
	hevc) enc=(-c:v libx265 -preset ultrafast -x265-params log-level=error) ;;
	h264) enc=(-c:v libx264 -preset ultrafast) ;;
	*) echo "skip   $name: codec $codec not handled"; continue ;;
	esac

	# second frame starts one frame before the end, so the file ends at $duration (rounded to a whole frame)
	# -nostdin: otherwise ffmpeg eats the next lines of the spec file
	ffmpeg -nostdin -y -hide_banner -loglevel error \
		-f lavfi -i "color=c=black:s=${width}x${height}:r=${fps}" -frames:v 2 \
		-vf "setpts=N*round((${duration}-1/(${fps}))/TB)" \
		"${enc[@]}" -pix_fmt yuv420p -fps_mode passthrough -f matroska "$dest"

	echo "video  $name  ($(stat -c %s "$dest") bytes)"
done <"$spec"

echo
echo "What ffprobe says now:"
for f in "$out"/*; do
	if ffprobe -v error -select_streams v:0 -show_entries stream=codec_name -of csv=p=0 "$f" 2>/dev/null | grep -q .; then
		printf '%s\t' "$(basename "$f")"
		ffprobe -v error -select_streams v:0 -show_entries stream=codec_name,width,height,avg_frame_rate,duration -of csv=p=0 "$f"
	fi
done
