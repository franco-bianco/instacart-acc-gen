
# Instacart Account Generator Script

This script automates the process of creating new Instacart accounts and applying coupons to them.

## Prerequisites

- Make sure you have Golang installed on your system.
- This script requires the following files in the `data` folder:
  - `accounts.csv`
  - `codes.txt` (with Instacart coupon codes)
  - `proxies.txt` (optional, not necessary for script execution)
  - `config.json`

## Configuration

- Fill in the `config.json` file in the following format:

    ```json
    {
      "account_quantity": [Number of accounts to generate],
      "gmail_address": "[Your Gmail Address]",
      "cap_solver_key": "[Your CAP Solver Key]"
    }
    ```

## Running the Script

- Execute the script using the following command:

    ```bash
    go run ./cmd
    ```

## Output

- If the account is successfully generated with a coupon value greater than $10, it will be recorded in the `accounts.csv` file.
- The `accounts.csv` file follows this format:

    ```text
    Timestamp,Email,Password,CouponValue
    2023-12-03 18:54:35,example+7105@gmail.com,Password1,50
    2023-12-03 19:01:49,example+7466@gmail.com,Password2,40
    2023-12-03 19:01:57,example+7563@gmail.com,Password3,40
    ```

## Notes

- Ensure that you have valid coupon codes in the `codes.txt` file for the script to apply to new accounts.
