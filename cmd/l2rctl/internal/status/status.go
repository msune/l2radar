package status

import (
	"encoding/json"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/msune/l2rctl/internal/docker"
)

const (
	ProbeContainer = "l2radar"
	UIContainer    = "l2radar-ui"
)

type containerInfo struct {
	State struct {
		Status    string `json:"Status"`
		StartedAt string `json:"StartedAt"`
	} `json:"State"`
}

type row struct {
	name    string
	status  string
	started string
}

func inspectContainer(r docker.Runner, name string) row {
	stdout, _, err := r.Run("inspect", "--type", "container", name)
	if err != nil {
		return row{name: name, status: "not found", started: "-"}
	}
	var infos []containerInfo
	if err := json.Unmarshal([]byte(stdout), &infos); err != nil || len(infos) == 0 {
		return row{name: name, status: "not found", started: "-"}
	}
	info := infos[0]
	started := info.State.StartedAt
	if started == "" {
		started = "-"
	}
	return row{name: name, status: info.State.Status, started: started}
}

// Status returns a formatted table of container statuses.
func Status(r docker.Runner) (string, error) {
	rows := []row{
		inspectContainer(r, ProbeContainer),
		inspectContainer(r, UIContainer),
	}

	var sb strings.Builder
	w := tabwriter.NewWriter(&sb, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "CONTAINER\tSTATUS\tSTARTED")
	for _, r := range rows {
		fmt.Fprintf(w, "%s\t%s\t%s\n", r.name, r.status, r.started)
	}
	w.Flush()
	return sb.String(), nil
}
