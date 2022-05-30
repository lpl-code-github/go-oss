package logs

import "net/http"

func Handler(w http.ResponseWriter, r *http.Request) {
	m := r.Method

	if m == http.MethodPost {
		post(w, r)
		return
	}

	w.WriteHeader(http.StatusMethodNotAllowed)
}
