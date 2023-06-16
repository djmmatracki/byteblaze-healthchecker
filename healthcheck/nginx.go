package healthcheck

import (
	"fmt"
	"net/http"
)

func indexHealthy() error {
	client := &http.Client{}

	req, err := http.NewRequest("GET", "http://nginx-1", nil)
	if err != nil {
		fmt.Println("Failed to create request:", err)
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Failed to send request:", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("error while sending request")
	}

	fmt.Println("Proper response successful")
	return nil
}
