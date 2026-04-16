package media

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/you/inkvault/internal/apierr"
	"github.com/you/inkvault/internal/domain"
	"github.com/you/inkvault/internal/repository"
)

// Handler exposes media upload/delete endpoints.
type Handler struct {
	svc   *Service
	store repository.Store
}

func NewHandler(svc *Service, store repository.Store) *Handler {
	return &Handler{svc: svc, store: store}
}

// Upload — POST /api/v1/media (multipart/form-data, field: "file")
func (h *Handler) Upload(c *fiber.Ctx) error {
	uploaderID, _ := c.Locals("userID").(string)

	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "no_file", "message": "Provide a file in the 'file' field",
		})
	}
	if file.Size > 10<<20 {
		return c.Status(fiber.StatusRequestEntityTooLarge).JSON(fiber.Map{
			"error": "too_large", "message": "File must be under 10 MB",
		})
	}

	f, err := file.Open()
	if err != nil {
		return apierr.ErrInternal.FiberResponse(c)
	}
	defer f.Close()

	data := make([]byte, file.Size)
	if _, err := f.Read(data); err != nil {
		return apierr.ErrInternal.FiberResponse(c)
	}

	result, err := h.svc.Upload(c.Context(), data, file.Filename, uploaderID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "upload_failed", "message": err.Error(),
		})
	}

	blogID := c.FormValue("blog_id")
	m := &domain.Media{
		ID:         uuid.New().String(),
		UploaderID: uploaderID,
		BlogID:     blogID,
		Filename:   file.Filename,
		MimeType:   result.MimeType,
		SizeBytes:  result.SizeBytes,
		StorageKey: result.Key,
		PublicURL:  result.PublicURL,
		CreatedAt:  time.Now(),
	}
	_ = h.store.Media().CreateMedia(c.Context(), m)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"url":        result.PublicURL,
		"key":        result.Key,
		"mime_type":  result.MimeType,
		"size_bytes": result.SizeBytes,
		"width":      result.Width,
		"height":     result.Height,
	})
}

// Delete — DELETE /api/v1/media/*key
func (h *Handler) Delete(c *fiber.Ctx) error {
	key := c.Params("*")
	uploaderID, _ := c.Locals("userID").(string)

	m, err := h.store.Media().GetMediaByKey(c.Context(), key)
	if err != nil || m.UploaderID != uploaderID {
		return apierr.ErrForbiddenEdit.FiberResponse(c)
	}
	if err := h.svc.Delete(c.Context(), key); err != nil {
		return apierr.ErrInternal.FiberResponse(c)
	}
	_ = h.store.Media().DeleteMedia(c.Context(), key)
	return c.Status(fiber.StatusNoContent).Send(nil)
}
