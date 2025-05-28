package collector

import (
	"encoding/json"
	"net/http"
	"strings"
)

type DockerContainerInfo struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Image  string `json:"image"`
	Status string `json:"status"`
}

func CollectDockerContainers() ([]DockerContainerInfo, error) {
	resp, err := http.Get("http://localhost:2375/containers/json?all=true")
	if err != nil {
		return []DockerContainerInfo{
			{ID: "-", Name: "-", Image: "-", Status: "Docker API not reachable"},
		}, nil
	}
	defer resp.Body.Close()

	var rawContainers []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&rawContainers); err != nil {
		return []DockerContainerInfo{
			{ID: "-", Name: "-", Image: "-", Status: "Invalid JSON from Docker API"},
		}, nil
	}

	var result []DockerContainerInfo
	for _, c := range rawContainers {
		id, _ := c["Id"].(string)
		image, _ := c["Image"].(string)
		status, _ := c["Status"].(string)

		name := "-"
		if names, ok := c["Names"].([]interface{}); ok && len(names) > 0 {
			name, _ = names[0].(string)
			name = strings.TrimPrefix(name, "/")
		}

		result = append(result, DockerContainerInfo{
			ID:     trimString(id, 12),
			Name:   name,
			Image:  image,
			Status: status,
		})
	}
	return result, nil
}

func trimString(s string, max int) string {
	if len(s) > max {
		return s[:max]
	}
	return s
}
