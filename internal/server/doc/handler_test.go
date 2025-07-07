// Методы для работы с метриками

// Поддерживает два типа метрик gauge и counter

//

// gauge чисто с плавающей точной

// counter целое положительное число

package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/megaded/metrictmr/internal/data"
)

func Example() {
	//Сохранение метрика
	metric := make([]data.Metric, 1)
	data, _ := json.Marshal(metric)
	method := "update"
	if len(metric) > 1 {
		method = "updates"
	}

	url := fmt.Sprintf("%s/%s", "update", method)

	req, _ := http.NewRequestWithContext(context.TODO(), http.MethodPost, url, bytes.NewBuffer(data))
	client := http.Client{}
	req.Header.Set("Content-Type", "application/json")
	res, err := client.Do(req)
	if err == nil {
		defer res.Body.Close()
	}
}
