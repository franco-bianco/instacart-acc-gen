package gen

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	AccountQuantity int    `json:"account_quantity"` // number of accounts to gen
	GmailAddress    string `json:"gmail_address"`    // gmail address to use
	CapSolverKey    string `json:"cap_solver_key"`   // https://www.capsolver.com/ API key
}

func LoadConfig() (*Config, error) {

	file, err := os.Open("data/config.json")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return nil, err
	}

	// validate config
	if config.AccountQuantity <= 0 {
		return nil, fmt.Errorf("account_quantity must be greater than 0")
	}
	if len(config.GmailAddress) == 0 {
		return nil, fmt.Errorf("catchall_domain must be set")
	}
	config.GmailAddress = strings.ToLower(config.GmailAddress)
	config.GmailAddress = strings.Split(config.GmailAddress, "@")[0]
	if len(config.CapSolverKey) == 0 {
		return nil, fmt.Errorf("capSolver API key not provided")
	}

	return &config, nil
}

type NewAccount struct {
	Timestamp   string
	Email       string
	Password    string
	CouponValue string
}

func (n *NewAccount) AddEntry() error {

	file, err := os.OpenFile("data/accounts.csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open accounts.csv: %w", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if fileInfo.Size() == 0 {
		header := []string{"Timestamp", "Email", "Password", "CouponValue"}
		if err := writer.Write(header); err != nil {
			return err
		}
	}

	record := []string{n.Timestamp, n.Email, n.Password, n.CouponValue}
	if err := writer.Write(record); err != nil {
		return err
	}

	return nil
}

func LoadTxtFile(filepath string) ([]string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines, nil
}
