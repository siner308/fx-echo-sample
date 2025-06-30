package middleware

import (
	"fxserver/pkg/dto"
	"net/http"

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
		if c.Response().Committed {
			return
		}

		var (
			code = http.StatusInternalServerError
			msg  = "Internal Server Error"
		)

		if he, ok := err.(*echo.HTTPError); ok {
			code = he.Code
			if he.Message != nil {
				if s, ok := he.Message.(string); ok {
					msg = s
				}
			}
		}

		em.logger.Error("HTTP error occurred",
			zap.Error(err),
			zap.Int("status_code", code),
			zap.String("method", c.Request().Method),
			zap.String("path", c.Request().URL.Path),
			zap.String("remote_ip", c.RealIP()),
		)

		errorResponse := dto.ErrorResponse{
			Error: msg,
		}

		if !c.Response().Committed {
			if err := c.JSON(code, errorResponse); err != nil {
				em.logger.Error("Failed to send error response", zap.Error(err))
			}
		}
	}
}