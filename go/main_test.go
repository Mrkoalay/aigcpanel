package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"xiacutai-server/internal/service"
)

func Test_AddUserAuthor(t *testing.T) {
	arr := []string{"C:\\Users\\kaola\\Downloads\\sound_replace_1770887225140_silence_0.wav",
		"C:\\Users\\kaola\\Downloads\\sound_replace_1770887225140_seg_0.wav",
		"C:\\Users\\kaola\\Downloads\\sound_replace_1770887225140_silence_1.wav",
		"C:\\Users\\kaola\\Downloads\\sound_replace_1770887225140_seg_1.wav",
		"C:\\Users\\kaola\\Downloads\\sound_replace_1770887225140_seg_2.wav",
		"C:\\Users\\kaola\\Downloads\\sound_replace_1770887225140_seg_3.wav",
		"C:\\Users\\kaola\\Downloads\\sound_replace_1770887225140_seg_4.wav",
		"C:\\Users\\kaola\\Downloads\\sound_replace_1770887225140_silence_5.wav",
		"C:\\Users\\kaola\\Downloads\\sound_replace_1770887225140_seg_5.wav",
		"C:\\Users\\kaola\\Downloads\\sound_replace_1770887225140_seg_6.wav",
		"C:\\Users\\kaola\\Downloads\\sound_replace_1770887225140_silence_7.wav",
		"C:\\Users\\kaola\\Downloads\\sound_replace_1770887225140_seg_7.wav",
		"C:\\Users\\kaola\\Downloads\\sound_replace_1770887225140_silence_8.wav",
		"C:\\Users\\kaola\\Downloads\\sound_replace_1770887225140_seg_8.wav",
		"C:\\Users\\kaola\\Downloads\\sound_replace_1770887225140_seg_9.wav",
		"C:\\Users\\kaola\\Downloads\\sound_replace_1770887225140_silence_10.wav",
		"C:\\Users\\kaola\\Downloads\\sound_replace_1770887225140_seg_10.wav",
		"C:\\Users\\kaola\\Downloads\\sound_replace_1770887225140_silence_11.wav",
		"C:\\Users\\kaola\\Downloads\\sound_replace_1770887225140_seg_11.wav",
		"C:\\Users\\kaola\\Downloads\\sound_replace_1770887225140_seg_12.wav",
		"C:\\Users\\kaola\\Downloads\\sound_replace_1770887225140_seg_13.wav",
		"C:\\Users\\kaola\\Downloads\\sound_replace_1770887225140_silence_14.wav",
		"C:\\Users\\kaola\\Downloads\\sound_replace_1770887225140_seg_14.wav",
		"C:\\Users\\kaola\\Downloads\\sound_replace_1770887225140_seg_15.wav",
		"C:\\Users\\kaola\\Downloads\\sound_replace_1770887225140_seg_16.wav",
		"C:\\Users\\kaola\\Downloads\\sound_replace_1770887225140_silence_17.wav",
		"C:\\Users\\kaola\\Downloads\\sound_replace_1770887225140_seg_17.wav",
		"C:\\Users\\kaola\\Downloads\\sound_replace_1770887225140_seg_18.wav",
		"C:\\Users\\kaola\\Downloads\\sound_replace_1770887225140_seg_19.wav",
		"C:\\Users\\kaola\\Downloads\\sound_replace_1770887225140_seg_20.wav",
		"C:\\Users\\kaola\\Downloads\\sound_replace_1770887225140_seg_21.wav",
		"C:\\Users\\kaola\\Downloads\\sound_replace_1770887225140_seg_22.wav",
		"C:\\Users\\kaola\\Downloads\\sound_replace_1770887225140_seg_23.wav",
		"C:\\Users\\kaola\\Downloads\\sound_replace_1770887225140_seg_24.wav",
		"C:\\Users\\kaola\\Downloads\\sound_replace_1770887225140_seg_25.wav",
		"C:\\Users\\kaola\\Downloads\\sound_replace_1770887225140_seg_26.wav",
		"C:\\Users\\kaola\\Downloads\\sound_replace_1770887225140_silence_27.wav",
		"C:\\Users\\kaola\\Downloads\\sound_replace_1770887225140_seg_27.wav",
	}

	service.FfmpegConcatAudio(arr, "out.wav")
}

func Test_CalcConcatDuration(t *testing.T) {
	dur, _ := calcConcatDuration("concat.txt")
	fmt.Printf("\nTOTAL = %.3f sec (%.2f min)\n", dur, dur/60)

}
func probeDuration(path string) (float64, error) {
	out, err := exec.Command(
		"ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		path,
	).CombinedOutput()

	if err != nil {
		return 0, fmt.Errorf("ffprobe failed: %v, %s", err, string(out))
	}

	return strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
}

func calcConcatDuration(txt string) (float64, error) {
	f, err := os.Open(txt)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	total := 0.0

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// 解析 file 'xxx.wav'
		if !strings.HasPrefix(line, "file") {
			continue
		}

		start := strings.Index(line, "'")
		end := strings.LastIndex(line, "'")
		if start == -1 || end == -1 || start >= end {
			continue
		}

		path := line[start+1 : end]

		dur, err := probeDuration(path)
		if err != nil {
			return 0, err
		}

		fmt.Printf("%s = %.3fs\n", filepath.Base(path), dur)
		total += dur
	}

	return total, scanner.Err()
}
