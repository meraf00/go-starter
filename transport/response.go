package transport

import (
	"encoding/json"
	"errors"
	"net/http"
)

func JSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if data != nil {
		return json.NewEncoder(w).Encode(data)
	}

	return nil
}

func Success(w http.ResponseWriter) error {
	return JSON(w, http.StatusOK, nil)
}

func SuccessWithData(w http.ResponseWriter, status int, data any) error {
	return JSON(w, status, data)
}

func Error(w http.ResponseWriter, err error) error {
	var appErr *AppError
	if errors.As(err, &appErr) {
		status := http.StatusInternalServerError
		switch appErr.Type {
		case BadRequestError:
			status = http.StatusBadRequest
		case ConflictError:
			status = http.StatusConflict
		case NotFoundError:
			status = http.StatusNotFound
		case ForbiddenError:
			status = http.StatusForbidden
		case UnauthorizedError:
			status = http.StatusUnauthorized
		}

		return JSON(w, status, err)
	}

	return JSON(w, http.StatusInternalServerError, map[string]any{
		"error":   "internal_server_error",
		"message": err.Error(),
	})
}
