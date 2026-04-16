package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/you/inkvault/internal/domain"
	"github.com/you/inkvault/internal/users"
)

func CanCreate(resource users.ResourceType) fiber.Handler { return permissionCheck(users.ActionCreate, resource) }
func CanEdit(resource users.ResourceType) fiber.Handler   { return permissionCheck(users.ActionEdit, resource) }
func CanDelete(resource users.ResourceType) fiber.Handler { return permissionCheck(users.ActionDelete, resource) }
func CanPublish(resource users.ResourceType) fiber.Handler { return permissionCheck(users.ActionPublish, resource) }

func permissionCheck(action users.Action, resource users.ResourceType) fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, _ := c.Locals("role").(string)
		userID, _ := c.Locals("userID").(string)
		u := &domain.User{ID: userID, Role: domain.Role(role), Status: domain.UserStatusActive}
		if !users.CanPerform(u, action, resource, "") {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":   "forbidden",
				"message": "You don't have permission to perform this action",
			})
		}
		return c.Next()
	}
}
