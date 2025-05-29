package handler

import (
	"github.com/android-sms-gateway/core/validator"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type Base struct {
	Validator *validator.Validate
	Logger    *zap.Logger
}

func (c *Base) BodyParserValidator(ctx *fiber.Ctx, out interface{}) error {
	if err := ctx.BodyParser(out); err != nil {
		c.Logger.Error("failed to parse request", zap.Error(err))
		return fiber.NewError(fiber.StatusBadRequest, "failed to parse request")
	}

	if err := c.Validator.Struct(out); err != nil {
		c.Logger.Error("failed to validate request", zap.Error(err))
		return fiber.NewError(fiber.StatusBadRequest, "failed to validate request")
	}

	if v, ok := out.(validator.Validatable); ok {
		if err := v.Validate(); err != nil {
			c.Logger.Error("failed to validate request", zap.Error(err))
			return fiber.NewError(fiber.StatusBadRequest, "failed to validate request")
		}
	}

	return nil
}
