package error_handler

import (
	"fmt"
	"net/http"
)

func LogError(err error, user_message string, w http.ResponseWriter) {
	if err != nil {
		// If any error logging it
		// fmt.Println("Error Occured Reason  : ", user_message)
		// fmt.Println("Error description : ", err)

		if w != nil {
			fmt.Fprintf(w, user_message)
		}
	}
}
