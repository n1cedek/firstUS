package deleteurl

import (
	resp "awesomeProject1/iternal/lib/api/response"
	"awesomeProject1/iternal/lib/logger/sl"
	"awesomeProject1/iternal/storage"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
)

//go:generate go run github.com/vektra/mockery/v2@v2.50.4 --name=URLSaver

type URLDeleter interface {
	DeleteURL(alias string) error
}

func New(log *slog.Logger, urlDeleter URLDeleter) http.HandlerFunc {
	const op = "handler.url.delete.New"

	return func(w http.ResponseWriter, r *http.Request) {

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		// Извлекаем alias из URL параметров
		alias := chi.URLParam(r, "alias")

		// Если alias пустой, возвращаем ошибку
		if alias == "" {

			log.Info("alias is empty")

			render.JSON(w, r, resp.Error("invalid request"))

			return
		}

		// Пытаемся удалить(удаляем) URL
		err := urlDeleter.DeleteURL(alias)

		// Обработка ошибки: если URL не найден
		if errors.Is(err, storage.ErrURLNotFound) {
			log.Info("url not found", "alias", alias)
			render.JSON(w, r, resp.Error("not found"))
			return
		}
		// Обработка других ошибок
		if err != nil {
			log.Error("failed to delete url", sl.Err(err))
			render.JSON(w, r, resp.Error("internal error"))
			return
		}
		// Логируем успешное удаление
		log.Info("url deleted successfully", slog.String("alias", alias))

		// Возвращаем успешный ответ
		render.JSON(w, r, resp.Success("URL deleted successfully"))
	}
}
