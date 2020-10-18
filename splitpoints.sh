#!/bin/bash

# Format: HH:MM:SS:[frame], where frame is between 0 and 24 (25fps)

video_segment_timecodes=(
	"00:00:00:00-00:07:03:23"
	"00:07:03:24-00:11:30:08"
	"00:11:30:10-00:21:05:06"
	"00:21:08:08-00:24:20:18"
	"00:24:22:16-00:33:20:11"
	"00:33:20:12-00:49:01:01"
	"00:49:04:02-01:07:55:18"
	"01:07:58:20-01:16:34:11"
)

timecode_to_frames() {
	rest=$1
	h=${rest%%:*}
	rest=${rest#*:}
	m=${rest%%:*}
	rest=${rest#*:}
	sec=${rest%%:*}
	rest=${rest#*:}
	frame=${rest%%:*}

	total_secs=$((10#$sec + 10#$m*60 + 10#$h*3600))
	total_frames=$((10#$total_secs*25 + 10#$frame))
	echo -n $total_frames
}

frames_to_timecode() {
	rest=$1
	frame=$((rest % 25))
	rest=$((rest / 25))
	sec=$((rest % 60))
	rest=$((rest / 60))
	m=$((rest % 60))
	rest=$((rest / 60))
	h=$((rest % 60))
	rest=$((rest / 60))
	printf "%02d:%02d:%02d:%02d" $h $m $sec $frame
}

timecode_to_decimal() {
	frame=${1##*:}
	if (( 10#$frame == 0 )); then
		echo ${1%:*}.000
		return
	fi
	frame_dec=$(echo "scale=3; ${frame} / 25" | bc -l)
	echo "${1%:*}${frame_dec}"
}

i=0
for segment in "${video_segment_timecodes[@]}";
do
	seg_start=${segment%-*}
	seg_end=${segment#*-}
	seg_frames=$(($(timecode_to_frames $seg_end) - 
			$(timecode_to_frames $seg_start)))
	seg_len=$(frames_to_timecode $seg_frames)

	seg_start_dec=$(timecode_to_decimal $seg_start)
	seg_len_dec=$(timecode_to_decimal $seg_len)

	echo ffmpeg -ss ${seg_start_dec} -i audio_corrected.mov \
		-to ${seg_len_dec} -c copy segment_${i}.mov
	let i++
done
