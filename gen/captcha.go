package gen

import (
	"encoding/json"
	"fmt"
	"time"
)

func (s *Session) createCapSolverTask() error {
	s.Log.Info("creating capSolver task...")

	res, err := s.Client.R().
		SetHeaders(map[string]string{
			"content-type": "application/json",
		}).
		SetBody(map[string]interface{}{
			"clientKey": s.UserConfig.CapSolverKey,
			"task": map[string]interface{}{
				"type":        "ReCaptchaV2TaskProxyLess",
				"websiteKey":  "6LeN0vMZAAAAAIKVl68OAJQy3zl8mZ0ESbkeEk1m",
				"websiteURL":  "https://www.instacart.com/",
				"isInvisible": true,
			},
		}).
		Post("https://api.capsolver.com/createTask")
	if err != nil {
		return err
	}
	if res.StatusCode() != 200 {
		fmt.Println(res.String())
		return fmt.Errorf("invalid status code: %d", res.StatusCode())
	}

	var taskRes map[string]interface{}
	if err := json.Unmarshal(res.Body(), &taskRes); err != nil {
		return err
	}

	taskID, ok := taskRes["taskId"].(string)
	if !ok {
		errorDesc, _ := taskRes["errorDescription"].(string)
		return fmt.Errorf("failed to get task ID: %s", errorDesc)
	}

	s.state.CapTaskID = taskID

	return nil
}

type capSolverTaskRes struct {
	ErrorID  int `json:"errorId"`
	Solution struct {
		UserAgent          string `json:"userAgent"`
		GRecaptchaResponse string `json:"gRecaptchaResponse"`
	} `json:"solution"`
	Status string `json:"status"`
}

func (s *Session) getTaskResult() (*capSolverTaskRes, error) {

	res, err := s.Client.R().
		SetHeaders(map[string]string{
			"content-type": "application/json",
		}).
		SetBody(map[string]interface{}{
			"clientKey": s.UserConfig.CapSolverKey,
			"taskId":    s.state.CapTaskID,
		}).
		Post("https://api.capsolver.com/getTaskResult")
	if err != nil {
		return nil, err
	}
	if res.StatusCode() != 200 {
		fmt.Println(res.String())
		return nil, fmt.Errorf("invalid status code: %d", res.StatusCode())
	}

	var taskRes capSolverTaskRes
	if err := json.Unmarshal(res.Body(), &taskRes); err != nil {
		return nil, err
	}

	return &taskRes, nil
}

func (s *Session) getCaptchaSolution() error {
	s.Log.Info("waiting for captcha solution...")

	timeout := time.After(60 * time.Second)
	tick := time.Tick(5 * time.Second)

	for {
		select {
		case <-s.Ctx.Done():
			return fmt.Errorf("context canceled")
		case <-timeout:
			return fmt.Errorf("timed out waiting for captcha solution")
		case <-tick:
			taskRes, err := s.getTaskResult()
			if err != nil {
				return err
			}
			if taskRes.Status == "ready" {
				s.state.ReCapToken = taskRes.Solution.GRecaptchaResponse
				return nil
			}
		}
	}
}
