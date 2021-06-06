package plug_controller

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

type plugController struct {
	sensorIDToPlugIPMap map[string]string
}

func New(
	sensors []struct {
		ID                  string `yaml:"id"`
		CorrespondingPlugIP string `yaml:"corresponding_plug_ip"`
	},
) *plugController {
	sensorIDToPlugIPMap := make(map[string]string)
	for _, sensor := range sensors {
		sensorIDToPlugIPMap[sensor.ID] = sensor.CorrespondingPlugIP
	}
	return &plugController{
		sensorIDToPlugIPMap,
	}
}

func (p *plugController) TurnOnCorrespondingPlug(sensorID string) error {
	return p.modifyState("on", sensorID)
}

func (p *plugController) TurnOffCorrespondingPlug(sensorID string) error {
	return p.modifyState("off", sensorID)
}

func (p *plugController) modifyState(state string, sensorID string) error {
	// Get the plug ID.
	plugIP := p.sensorIDToPlugIPMap[sensorID]
	if plugIP == "" {
		message := fmt.Sprintf("Could not get plug ID for sensor ID %s when attempting to turn %s the plug.", sensorID, state)
		return errors.New(message)
	}

	// Set the state of the plug.
	command := fmt.Sprintf("kasa --host %s --plug %s", plugIP, state)
	err, stdout, stderr := runCommand(command)
	if err != nil {
		message := fmt.Sprintf("Failed to turn %s for plug at IP %s for sensor %s with error %s and stderr %s", state, plugIP, sensorID, err.Error(), stderr)
		return errors.New(message)
	}
	if stderr != "" {
		message := fmt.Sprintf("Failed to turn %s for plug at IP %s for sensor %s with stderr %s", state, plugIP, sensorID, stderr)
		return errors.New(message)
	}

	// Validate the state was changed.
	command = fmt.Sprintf("kasa --host %s --plug state", plugIP)
	err, stdout, stderr = runCommand(command)
	if err != nil {
		message := fmt.Sprintf("Failed to validate state %s for plug at IP %s for sensor %s with error %s and stderr %s", state, plugIP, sensorID, err.Error(), stderr)
		return errors.New(message)
	}
	if stderr != "" {
		message := fmt.Sprintf("Failed to validate %s plug at IP %s for sensor %s with stderr %s", state, plugIP, sensorID, stderr)
		return errors.New(message)
	}
	expectedState := fmt.Sprintf("Device state: %s", strings.ToUpper(state))
	stateMatches := strings.Contains(stdout, expectedState)
	if !stateMatches {
		message := fmt.Sprintf("Unexpected state when turning plug %s at IP %s for sensor %s with stdout %s", state, plugIP, sensorID, stdout)
		return errors.New(message)
	}

	// All validated!
	return nil
}

func runCommand(command string) (error, string, string) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command("sh", "-c", command)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return err, stdout.String(), stderr.String()
}
