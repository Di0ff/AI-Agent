package server

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"aiAgent/internal/config"
	"aiAgent/internal/database"
	"aiAgent/internal/logger"
)

type Server struct {
	cfg  *config.Cfg
	log  *logger.Zap
	db   *database.Database
	repo *database.TaskRepository
}

func New(cfg *config.Cfg, log *logger.Zap, db *database.Database) *Server {
	return &Server{
		cfg:  cfg,
		log:  log,
		db:   db,
		repo: database.NewTaskRepository(db.DB),
	}
}

func (s *Server) Run(ctx context.Context) error {
	r := gin.New()
	r.Use(gin.Recovery())

	// Простейший лог-мидлвар
	r.Use(func(c *gin.Context) {
		s.log.Info("HTTP",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
		)
		c.Next()
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Создать задачу
	r.POST("/api/task", func(c *gin.Context) {
		var req struct {
			UserInput string `json:"user_input" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		task := database.Task{
			UserInput: req.UserInput,
			Status:    "pending",
		}
		if err := s.repo.CreateTask(&task); err != nil {
			s.log.Error("db create task", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"task_id": task.ID})
	})

	// Получить задачу
	r.GET("/api/task/:id", func(c *gin.Context) {
		id64, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
			return
		}
		task, err := s.repo.GetTaskByID(uint(id64))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusOK, task)
	})

	// Список задач
	r.GET("/api/tasks", func(c *gin.Context) {
		tasks, err := s.repo.ListTasks(50, 0)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
			return
		}
		c.JSON(http.StatusOK, tasks)
	})

	addr := fmt.Sprintf("%s:%s", s.cfg.App.Host, s.cfg.App.Port)
	s.log.Info("Сервер запущен", zap.String("addr", addr))
	return r.Run(addr)
}
