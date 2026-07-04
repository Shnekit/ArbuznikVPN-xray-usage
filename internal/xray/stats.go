package xray

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/yourusername/xray-usage-collector/internal/model"
)

const apiServer = "127.0.0.1:10085"

type statsResponse struct {
	Stat []statEntry `json:"stat"`
}

type statEntry struct {
	Name  string `json:"name"`
	Value uint64 `json:"value"`
}

func ReadStats() ([]model.UserStats, error) {

	cmd := exec.Command(
		"xray",
		"api",
		"statsquery",
		"--server="+apiServer,
	)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute xray statsquery: %w", err)
	}

	var response statsResponse

	if err := json.Unmarshal(output, &response); err != nil {
		return nil, fmt.Errorf("failed to parse xray JSON: %w", err)
	}

	users := make(map[string]*model.UserStats)

	for _, stat := range response.Stat {

		parts := strings.Split(stat.Name, ">>>")

		// Expected:
		//
		// user>>>Dad1>>>traffic>>>uplink
		//
		if len(parts) != 4 {
			continue
		}

		if parts[0] != "user" {
			continue
		}

		if parts[2] != "traffic" {
			continue
		}

		email := parts[1]
		direction := parts[3]

		user, exists := users[email]
		if !exists {

			user = &model.UserStats{
				Email: email,
			}

			users[email] = user
		}

		switch direction {

		case "uplink":
			user.UploadBytes = stat.Value

		case "downlink":
			user.DownloadBytes = stat.Value
		}
	}

	result := make([]model.UserStats, 0, len(users))

	for _, user := range users {
		result = append(result, *user)
	}

	return result, nil
}
