package middleware

import (
	"net/http"

	"fxserver/pkg/dto"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type ErrorMiddleware struct {
	logger *zap.Logger
}

func NewErrorMiddleware(logger *zap.Logger) *ErrorMiddleware {
	return &ErrorMiddleware{
		logger: logger,
	}
}

func (em *ErrorMiddleware) ErrorHandler() echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		code := http.StatusInternalServerError
		message := "Internal server error"

		if he, ok := err.(*echo.HTTPError); ok {
			code = he.Code
			if msg, ok := he.Message.(string); ok {
				message = msg
			}
		}

		// Log the error
		em.logger.Error("HTTP error occurred",
			zap.Error(err),
			zap.Int("status_code", code),
			zap.String("method", c.Request().Method),
			zap.String("path", c.Request().URL.Path),
			zap.String("remote_ip", c.RealIP()),
		)

		// Send error response
		if !c.Response().Committed {
			if c.Request().Method == http.MethodHead {
				err = c.NoContent(code)
			} else {
				err = c.JSON(code, dto.ErrorResponse{
					Error: message,
				})
			}
			if err != nil {
				em.logger.Error("Failed to send error response", zap.Error(err))
			}
		}
	}
}