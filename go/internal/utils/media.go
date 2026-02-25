package utils

import (
	"encoding/json"
	"fmt"
	"math"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

type ffprobeJSONResult struct {
	Streams []struct {
		Width        int    `json:"width"`
		Height       int    `json:"height"`
		RFrameRate   string `json:"r_frame_rate"`
		AvgFrameRate string `json:"avg_frame_rate"`
	} `json:"streams"`
	Format struct {
		Duration string `json:"duration"`
	} `json:"format"`
}

func GetFFprobePath() string {
	name := "ffprobe"
	if runtime.GOOS == "windows" {
		name = "ffprobe.exe"
	}
	return filepath.Join(GetExeDir(), "binary", name)
}

func ProbeVideoInfo(file string) (string, error) {
	out, err := exec.Command(
		GetFFprobePath(),
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=width,height,r_frame_rate,avg_frame_rate",
		"-show_entries", "format=duration",
		"-of", "json",
		file,
	).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("ffprobe failed: %w, output: %s", err, strings.TrimSpace(string(out)))
	}

	var result ffprobeJSONResult
	if err := json.Unmarshal(out, &result); err != nil {
		return "", err
	}

	info := map[string]any{}
	if len(result.Streams) > 0 {
		stream := result.Streams[0]
		if stream.Width > 0 {
			info["width"] = stream.Width
		}
		if stream.Height > 0 {
			info["height"] = stream.Height
		}
		fps := parseFFprobeFPS(stream.AvgFrameRate)
		if fps <= 0 {
			fps = parseFFprobeFPS(stream.RFrameRate)
		}
		if fps > 0 {
			info["fps"] = normalizeFloat(fps)
		}
	}

	if result.Format.Duration != "" {
		dur, err := strconv.ParseFloat(strings.TrimSpace(result.Format.Duration), 64)
		if err == nil && dur > 0 {
			info["duration"] = normalizeFloat(dur)
		}
	}

	raw, err := json.Marshal(info)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func parseFFprobeFPS(value string) float64 {
	value = strings.TrimSpace(value)
	if value == "" || value == "0/0" {
		return 0
	}
	parts := strings.Split(value, "/")
	if len(parts) != 2 {
		f, _ := strconv.ParseFloat(value, 64)
		return f
	}
	n, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0
	}
	d, err := strconv.ParseFloat(parts[1], 64)
	if err != nil || d == 0 {
		return 0
	}
	return n / d
}

func normalizeFloat(v float64) any {
	rounded := math.Round(v*100) / 100
	if math.Abs(rounded-math.Round(rounded)) < 1e-9 {
		return int64(math.Round(rounded))
	}
	return rounded
}
