package fails

import (
	"encoding/json"
	"html/template"
	"net/http"
)

var errors = map[int]string{
	400: "There seems to be a problem with your request. Please check your input and try again.",
	401: "You need to be logged in to access this resource. Please sign in or check your credentials.",
	403: "You don't have permission to access this resource. Please contact an administrator.",
	404: "We couldn't find what you're looking for. The page or resource may have been moved or deleted.",
	408: "The request took too long to complete. Please try again.",
	500: "Something went wrong on our end. Our team has been notified and we're working on it.",
	502: "We're having trouble connecting to our servers. Please try again in a few moments.",
	503: "Our service is temporarily unavailable. We're working to restore it as quickly as possible.",
	504: "The server took too long to respond. Please try again later.",
}

var errorTitles = map[int]string{
	400: "Bad Request",
	401: "Unauthorized",
	403: "Forbidden",
	404: "Not Found",
	405: "Method Not Allowed",
	408: "Request Timeout",
	500: "Internal Server Error",
	502: "Bad Gateway",
	503: "Service Unavailable",
	504: "Gateway Timeout",

}
func ErrorPageHandler(w http.ResponseWriter, r *http.Request, status int) {
	w.WriteHeader(status)
	tmpl, err := template.ParseFiles("templates/error.html")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data:=struct{
		Code int
		Message string
		Title string
	}{
		Code: status,
		Message: errors[status],
		Title: errorTitles[status],
	}
	tmpl.Execute(w, data)
}

func JSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error":   http.StatusText(status),
		"message": message,
	})
}
