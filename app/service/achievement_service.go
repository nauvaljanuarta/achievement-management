package service

import (
	"fmt"
	"time"

	"achievement-backend/app/models"
	"achievement-backend/app/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type AchievementService struct {
	achievementRepo    repository.AchievementRepository
	achievementRefRepo repository.AchievementReferenceRepository
	studentRepo        repository.StudentRepository
	lecturerRepo       repository.LecturerRepository
	userRepo           repository.UserRepository
	roleRepo           repository.RoleRepository
}

func NewAchievementService(
	achievementRepo repository.AchievementRepository,
	achievementRefRepo repository.AchievementReferenceRepository,
	studentRepo repository.StudentRepository,
	lecturerRepo repository.LecturerRepository,
	userRepo repository.UserRepository,
	roleRepo repository.RoleRepository, // TAMBAHKAN
) *AchievementService {
	return &AchievementService{
		achievementRepo:    achievementRepo,
		achievementRefRepo: achievementRefRepo,
		studentRepo:        studentRepo,
		lecturerRepo:       lecturerRepo,
		userRepo:           userRepo,
		roleRepo:           roleRepo, // TAMBAHKAN
	}
}

func (s *AchievementService) CreateAchievement(c *fiber.Ctx) error {
	ctx := c.UserContext()

	// Get user info from JWT
	user, ok := c.Locals("user").(*models.User)
	if !ok {
		return c.Status(401).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	// Get user's role
	role, err := s.roleRepo.GetByID(user.RoleID)
	if err != nil || role == nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to get user role",
		})
	}

	// Parse request body
	var req models.CreateAchievementRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
	}

	// Validate achievement type
	validTypes := []string{"academic", "competition", "organization", "publication", "certification", "other"}
	if !contains(validTypes, req.AchievementType) {
		return c.Status(400).JSON(fiber.Map{
			"error": fmt.Sprintf("Invalid achievement type. Valid types: %v", validTypes),
		})
	}

	var studentID uuid.UUID

	// Determine target student ID based on role
	switch role.Name {
	case "Mahasiswa":
		// Mahasiswa can only create achievements for themselves
		student, err := s.studentRepo.GetByUserID(user.ID)
		if err != nil || student == nil {
			return c.Status(403).JSON(fiber.Map{
				"error": "User is not a student or student profile not found",
			})
		}
		studentID = student.ID

	case "Admin":
		if req.StudentID == nil || *req.StudentID == uuid.Nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "student_id is required when creating achievement as admin",
			})
		}

		student, err := s.studentRepo.GetByID(*req.StudentID)
		if err != nil || student == nil {
			return c.Status(404).JSON(fiber.Map{
				"error": "Student not found",
			})
		}
		studentID = *req.StudentID

	case "Dosen Wali":
		return c.Status(403).JSON(fiber.Map{
			"error": "Dosen Wali cannot create achievements",
		})

	default:
		return c.Status(403).JSON(fiber.Map{
			"error": "Unauthorized role",
		})
	}

	achievement := &models.Achievement{
		StudentID:       studentID.String(),
		AchievementType: req.AchievementType,
		Title:           req.Title,
		Description:     req.Description,
		Details:         req.Details,
		Attachments:     req.Attachments,
		Tags:            req.Tags,
		Points:          req.Points,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	mongoID, err := s.achievementRepo.Create(ctx, achievement)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to create achievement",
			"details": err.Error(),
		})
	}

	ref := &models.AchievementReference{
		ID:                 uuid.New(),
		StudentID:          studentID,
		MongoAchievementID: mongoID.Hex(),
		Status:             "draft",
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	if err := s.achievementRefRepo.Create(ctx, ref); err != nil {
		s.achievementRepo.Delete(ctx, mongoID)
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to create achievement reference",
			"details": err.Error(),
		})
	}

	return c.Status(201).JSON(fiber.Map{
		"success": true,
		"message": "Achievement created successfully",
		"data": fiber.Map{
			"id":               mongoID.Hex(),       // MongoDB ID
			"reference_id":     ref.ID,              // PostgreSQL reference ID
			"student_id":       studentID,           // Target student
			"achievement_type": req.AchievementType, // Type: competition/publication/etc
			"title":            req.Title,           // Judul prestasi
			"description":      req.Description,     // Deskripsi
			"status":           ref.Status,          // Status (draft)
			"points":           req.Points,          // Poin
			"created_at":       ref.CreatedAt,       // Waktu dibuat
			"created_by":       user.ID,             // Siapa yang buat (optional)
		},
	})
}

// GetAchievementByID - Get detail achievement by ID
func (s *AchievementService) GetAchievementByID(c *fiber.Ctx) error {
	ctx := c.UserContext()

	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	currentUser, err := s.userRepo.GetByID(userID)
	if err != nil || currentUser == nil {
		return c.Status(401).JSON(fiber.Map{"error": "User not found"})
	}

	achievementID := c.Params("id")

	// Get reference first
	refUUID, err := uuid.Parse(achievementID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid achievement ID"})
	}

	ref, err := s.achievementRefRepo.FindByID(ctx, refUUID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get achievement reference"})
	}
	if ref == nil {
		return c.Status(404).JSON(fiber.Map{"error": "Achievement not found"})
	}

	// Check permissions based on role
	role, err := s.roleRepo.GetByID(currentUser.RoleID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get user role"})
	}
	if role == nil {
		return c.Status(403).JSON(fiber.Map{"error": "Role not found"})
	}

	hasAccess := false

	switch role.Name {
	case "Mahasiswa":
		// Student can only see their own achievements
		student, _ := s.studentRepo.GetByUserID(userID)
		if student != nil && student.ID == ref.StudentID {
			hasAccess = true
		}
	case "Dosen Wali":
		// Dosen can only see achievements of their advisees
		studentIDs, _ := s.achievementRefRepo.GetStudentIDsByAdvisor(ctx, userID)
		for _, sid := range studentIDs {
			if sid == ref.StudentID {
				hasAccess = true
				break
			}
		}
	case "Admin":
		// Admin can see all
		hasAccess = true
	default:
		hasAccess = false
	}

	if !hasAccess {
		return c.Status(403).JSON(fiber.Map{"error": "Access denied"})
	}

	// Get achievement detail from MongoDB
	mongoID, err := primitive.ObjectIDFromHex(ref.MongoAchievementID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Invalid MongoDB ID"})
	}

	achievement, err := s.achievementRepo.FindByID(ctx, mongoID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get achievement details"})
	}
	if achievement == nil {
		return c.Status(404).JSON(fiber.Map{"error": "Achievement details not found"})
	}

	// Get student info for response
	student, _ := s.studentRepo.GetByID(ref.StudentID)
	var studentName, studentNIM string
	if student != nil {
		studentUser, _ := s.userRepo.GetByID(student.UserID)
		if studentUser != nil {
			studentName = studentUser.FullName
		}
		studentNIM = student.StudentID
	}

	// Get verifier info if exists
	var verifierName *string
	if ref.VerifiedBy != nil {
		verifierUser, _ := s.userRepo.GetByID(*ref.VerifiedBy)
		if verifierUser != nil {
			name := verifierUser.FullName
			verifierName = &name
		}
	}

	// Combine data
	response := fiber.Map{
		"id": ref.ID,
		"student": fiber.Map{
			"id":   student.ID,
			"name": studentName,
			"nim":  studentNIM,
		},
		"status":         ref.Status,
		"submitted_at":   ref.SubmittedAt,
		"verified_at":    ref.VerifiedAt,
		"verified_by":    ref.VerifiedBy,
		"verifier_name":  verifierName,
		"rejection_note": ref.RejectionNote,
		"created_at":     ref.CreatedAt,
		"updated_at":     ref.UpdatedAt,
		"achievement":    achievement,
	}

	return c.JSON(fiber.Map{
		"data": response,
	})
}

// GetMyAchievements - Mahasiswa melihat prestasi sendiri
func (s *AchievementService) GetMyAchievements(c *fiber.Ctx) error {
	ctx := c.UserContext()

	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	// Get student ID
	student, err := s.studentRepo.GetByUserID(userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get student profile"})
	}
	if student == nil {
		return c.Status(403).JSON(fiber.Map{"error": "User is not a student"})
	}

	// Get query parameters
	status := c.Query("status", "")
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// Get references from PostgreSQL
	refs, total, err := s.achievementRefRepo.FindByStudentID(ctx, student.ID, status, page, limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to get achievements",
			"details": err.Error(),
		})
	}

	// Get achievement details from MongoDB
	var achievements []fiber.Map
	for _, ref := range refs {
		mongoID, err := primitive.ObjectIDFromHex(ref.MongoAchievementID)
		if err != nil {
			continue // Skip invalid IDs
		}

		achievement, err := s.achievementRepo.FindByID(ctx, mongoID)
		if err != nil {
			continue // Skip if can't get details
		}

		achievements = append(achievements, fiber.Map{
			"id":           ref.ID,
			"status":       ref.Status,
			"title":        achievement.Title,
			"type":         achievement.AchievementType,
			"points":       achievement.Points,
			"submitted_at": ref.SubmittedAt,
			"verified_at":  ref.VerifiedAt,
			"created_at":   ref.CreatedAt,
		})
	}

	totalPages := (total + limit - 1) / limit
	hasNext := page < totalPages
	hasPrev := page > 1

	return c.JSON(fiber.Map{
		"data": achievements,
		"pagination": fiber.Map{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": totalPages,
			"has_next":    hasNext,
			"has_prev":    hasPrev,
		},
	})
}

func (s *AchievementService) UpdateAchievement(c *fiber.Ctx) error {
	ctx := c.UserContext()
	
	mongoIDHex := c.Params("id")
	
	mongoID, err := primitive.ObjectIDFromHex(mongoIDHex)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid achievement ID"})
	}

	// Get achievement dari MongoDB
	achievement, err := s.achievementRepo.FindByID(ctx, mongoID)
	if err != nil || achievement == nil {
		return c.Status(404).JSON(fiber.Map{"error": "Achievement not found"})
	}

	// Get reference
	ref, err := s.achievementRefRepo.FindByMongoID(ctx, mongoIDHex)
	if err != nil || ref == nil {
		return c.Status(404).JSON(fiber.Map{"error": "Achievement status not found"})
	}

	// Get user info
	userID, _ := c.Locals("user_id").(uuid.UUID)
	user, _ := c.Locals("user").(*models.User)
	userRole, err := s.roleRepo.GetByID(user.RoleID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get user role"})
	}

	// Ownership check
	if userRole.Name == "Mahasiswa" {
		student, _ := s.studentRepo.GetByUserID(userID)
		if student == nil {
			return c.Status(403).JSON(fiber.Map{"error": "Student not found"})
		}
		
		studentUUID, _ := uuid.Parse(achievement.StudentID)
		if student.ID != studentUUID {
			return c.Status(403).JSON(fiber.Map{"error": "Not your achievement"})
		}
	}

	// Status check
	if ref.Status != "draft" {
		return c.Status(400).JSON(fiber.Map{"error": "Only draft achievements can be updated"})
	}

	// Parse request body sebagai map
	var req map[string]interface{}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if title, ok := req["title"].(string); ok && title != "" {
		achievement.Title = title
	}
	if description, ok := req["description"].(string); ok {
		achievement.Description = description
	}
	if points, ok := req["points"].(float64); ok {
		achievement.Points = int(points)
	}
	if tags, ok := req["tags"].([]interface{}); ok && len(tags) > 0 {
		var newTags []string
		for _, tag := range tags {
			if str, ok := tag.(string); ok {
				newTags = append(newTags, str)
			}
		}
		if len(newTags) > 0 {
			achievement.Tags = newTags
		}
	}

	// Update details jika ada
	if details, ok := req["details"].(map[string]interface{}); ok {
		// Update competition fields
		if compName, ok := details["competition_name"].(string); ok && compName != "" {
			achievement.Details.CompetitionName = &compName
		}
		if compLevel, ok := details["competition_level"].(string); ok && compLevel != "" {
			achievement.Details.CompetitionLevel = &compLevel
		}
		
		// Update string fields
		if pubTitle, ok := details["publication_title"].(string); ok && pubTitle != "" {
			achievement.Details.PublicationTitle = pubTitle
		}
		if publisher, ok := details["publisher"].(string); ok && publisher != "" {
			achievement.Details.Publisher = publisher
		}
		
		// Update organization
		if orgName, ok := details["organization_name"].(string); ok && orgName != "" {
			achievement.Details.OrganizationName = orgName
		}
		if position, ok := details["position"].(string); ok && position != "" {
			achievement.Details.Position = position
		}
		
		if loc, ok := details["location"].(string); ok && loc != "" {
			achievement.Details.Location = loc
		}
		if org, ok := details["organizer"].(string); ok && org != "" {
			achievement.Details.Organizer = org
		}
		if score, ok := details["score"].(float64); ok {
			achievement.Details.Score = int(score)
		}
		
		if customFields, ok := details["custom_fields"].(map[string]interface{}); ok {
			if achievement.Details.CustomFields == nil {
				achievement.Details.CustomFields = make(map[string]interface{})
			}
			for key, value := range customFields {
				achievement.Details.CustomFields[key] = value
			}
		}
	}

	achievement.UpdatedAt = time.Now()

	if err := s.achievementRepo.Update(ctx, mongoID, achievement); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update achievement"})
	}
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Achievement updated successfully",
		"data": fiber.Map{
			"id":         mongoID.Hex(),
			"status":     ref.Status,
			"updated_at": achievement.UpdatedAt,
		},
	})
}

func (s *AchievementService) DeleteAchievement(c *fiber.Ctx) error {
	ctx := c.UserContext()
	
	mongoIDHex := c.Params("id")
	
	mongoID, err := primitive.ObjectIDFromHex(mongoIDHex)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid achievement ID"})
	}

	achievement, err := s.achievementRepo.FindByID(ctx, mongoID)
	if err != nil || achievement == nil {
		return c.Status(404).JSON(fiber.Map{"error": "Achievement not found"})
	}

	ref, err := s.achievementRefRepo.FindByMongoID(ctx, mongoIDHex)
	if err != nil || ref == nil {
		return c.Status(404).JSON(fiber.Map{"error": "Achievement status not found"})
	}

	userID, _ := c.Locals("user_id").(uuid.UUID)
	user, _ := c.Locals("user").(*models.User)
	userRole, err := s.roleRepo.GetByID(user.RoleID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get user role"})
	}

	if userRole.Name == "Mahasiswa" {
		student, _ := s.studentRepo.GetByUserID(userID)
		if student == nil {
			return c.Status(403).JSON(fiber.Map{"error": "Student not found"})
		}
		
		studentUUID, _ := uuid.Parse(achievement.StudentID)
		if student.ID != studentUUID {
			return c.Status(403).JSON(fiber.Map{"error": "Not your achievement"})
		}
	}

	if ref.Status != "draft" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Only draft achievements can be deleted",
			"current_status": ref.Status,
		})
	}

	if err := s.achievementRefRepo.SoftDelete(ctx, ref.ID); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to delete achievement",
		})
	}

	// 9. Response
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Achievement deleted successfully",
		"data": fiber.Map{
			"id":             mongoID.Hex(),
			"reference_id":   ref.ID,
			"previous_status": ref.Status,
			"new_status":     "deleted",
			"deleted_at":     time.Now(),
		},
	})
}

// SubmitAchievement - Mahasiswa submit prestasi untuk verifikasi
func (s *AchievementService) SubmitAchievement(c *fiber.Ctx) error {
	ctx := c.UserContext()
	achievementID := c.Params("id")

	refUUID, err := uuid.Parse(achievementID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid achievement ID"})
	}

	ref, err := s.achievementRefRepo.FindByID(ctx, refUUID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get achievement"})
	}
	if ref == nil {
		return c.Status(404).JSON(fiber.Map{"error": "Achievement not found"})
	}

	// Check if user is the owner
	userID, _ := c.Locals("user_id").(uuid.UUID)
	student, _ := s.studentRepo.GetByUserID(userID)
	if student == nil || student.ID != ref.StudentID {
		return c.Status(403).JSON(fiber.Map{"error": "Access denied"})
	}

	// Check if already submitted/verified
	if ref.Status != "draft" {
		return c.Status(400).JSON(fiber.Map{
			"error": fmt.Sprintf("Achievement cannot be submitted. Current status: %s", ref.Status),
		})
	}

	// Update status to submitted
	if err := s.achievementRefRepo.SubmitForVerification(ctx, ref.ID); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to submit achievement",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Achievement submitted for verification",
		"status":  "submitted",
	})
}

// GetAdviseeAchievements - Dosen melihat prestasi mahasiswa bimbingan
func (s *AchievementService) GetAdviseeAchievements(c *fiber.Ctx) error {
	ctx := c.UserContext()

	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	// Get lecturer ID
	lecturer, err := s.lecturerRepo.GetByUserID(userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get lecturer profile"})
	}
	if lecturer == nil {
		return c.Status(403).JSON(fiber.Map{"error": "User is not a lecturer"})
	}

	// Get query parameters
	status := c.Query("status", "")
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// Get references from PostgreSQL
	refs, total, err := s.achievementRefRepo.FindByAdvisorID(ctx, lecturer.ID, status, page, limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to get achievements",
			"details": err.Error(),
		})
	}

	// Get achievement details from MongoDB
	var achievements []fiber.Map
	for _, ref := range refs {
		mongoID, err := primitive.ObjectIDFromHex(ref.MongoAchievementID)
		if err != nil {
			continue
		}

		achievement, err := s.achievementRepo.FindByID(ctx, mongoID)
		if err != nil {
			continue
		}

		// Get student info
		student, _ := s.studentRepo.GetByID(ref.StudentID)
		var studentName string
		if student != nil {
			studentUser, _ := s.userRepo.GetByID(student.UserID)
			if studentUser != nil {
				studentName = studentUser.FullName
			}
		}

		achievements = append(achievements, fiber.Map{
			"id":           ref.ID,
			"status":       ref.Status,
			"title":        achievement.Title,
			"type":         achievement.AchievementType,
			"points":       achievement.Points,
			"submitted_at": ref.SubmittedAt,
			"verified_at":  ref.VerifiedAt,
			"created_at":   ref.CreatedAt,
			"student": fiber.Map{
				"id":   student.ID,
				"name": studentName,
				"nim":  student.StudentID,
			},
		})
	}

	totalPages := (total + limit - 1) / limit
	hasNext := page < totalPages
	hasPrev := page > 1

	return c.JSON(fiber.Map{
		"data": achievements,
		"pagination": fiber.Map{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": totalPages,
			"has_next":    hasNext,
			"has_prev":    hasPrev,
		},
	})
}

// VerifyAchievement - Dosen memverifikasi prestasi
func (s *AchievementService) VerifyAchievement(c *fiber.Ctx) error {
	ctx := c.UserContext()
	achievementID := c.Params("id")

	refUUID, err := uuid.Parse(achievementID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid achievement ID"})
	}

	// Get lecturer ID
	userID, _ := c.Locals("user_id").(uuid.UUID)
	lecturer, err := s.lecturerRepo.GetByUserID(userID)
	if err != nil || lecturer == nil {
		return c.Status(403).JSON(fiber.Map{"error": "User is not a lecturer"})
	}

	// Verify achievement
	if err := s.achievementRefRepo.VerifyAchievement(ctx, refUUID, lecturer.ID); err != nil {
		statusCode := 500
		errorMsg := "Failed to verify achievement"

		if err.Error() == "achievement not found or not in submitted status" {
			statusCode = 400
			errorMsg = err.Error()
		}

		return c.Status(statusCode).JSON(fiber.Map{
			"error":   errorMsg,
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Achievement verified successfully",
		"status":  "verified",
	})
}

// RejectAchievement - Dosen menolak prestasi
func (s *AchievementService) RejectAchievement(c *fiber.Ctx) error {
	ctx := c.UserContext()
	achievementID := c.Params("id")

	refUUID, err := uuid.Parse(achievementID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid achievement ID"})
	}

	var req struct {
		RejectionNote string `json:"rejection_note"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.RejectionNote == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Rejection note is required"})
	}

	// Get lecturer ID
	userID, _ := c.Locals("user_id").(uuid.UUID)
	lecturer, err := s.lecturerRepo.GetByUserID(userID)
	if err != nil || lecturer == nil {
		return c.Status(403).JSON(fiber.Map{"error": "User is not a lecturer"})
	}

	// Reject achievement
	if err := s.achievementRefRepo.RejectAchievement(ctx, refUUID, lecturer.ID, req.RejectionNote); err != nil {
		statusCode := 500
		errorMsg := "Failed to reject achievement"

		if err.Error() == "achievement not found or not in submitted status" {
			statusCode = 400
			errorMsg = err.Error()
		}

		return c.Status(statusCode).JSON(fiber.Map{
			"error":   errorMsg,
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Achievement rejected",
		"status":  "rejected",
	})
}

// GetAchievementStatistics - Get statistics for achievements
func (s *AchievementService) GetAchievementStatistics(c *fiber.Ctx) error {
	ctx := c.UserContext()

	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	currentUser, err := s.userRepo.GetByID(userID)
	if err != nil || currentUser == nil {
		return c.Status(401).JSON(fiber.Map{"error": "User not found"})
	}

	// Get role
	role, err := s.roleRepo.GetByID(currentUser.RoleID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get user role"})
	}
	if role == nil {
		return c.Status(403).JSON(fiber.Map{"error": "Role not found"})
	}

	var statistics *models.AchievementStatistics
	var errStat error

	switch role.Name {
	case "Mahasiswa":
		student, _ := s.studentRepo.GetByUserID(userID)
		if student != nil {
			statistics, errStat = s.achievementRepo.GetStatisticsByStudent(ctx, student.ID)
		} else {
			return c.Status(403).JSON(fiber.Map{"error": "User is not a student"})
		}
	case "Dosen Wali":
		lecturer, _ := s.lecturerRepo.GetByUserID(userID)
		if lecturer != nil {
			statistics, errStat = s.achievementRepo.GetStatisticsByAdvisor(ctx, lecturer.ID)
		} else {
			return c.Status(403).JSON(fiber.Map{"error": "User is not a lecturer"})
		}
	case "Admin":
		// For admin, return empty or implement admin-specific stats
		statistics = &models.AchievementStatistics{
			TotalAchievements: 0,
			ByType:            make(map[string]int),
			ByPeriod:          make(map[string]int),
			TopTags:           []string{},
			AveragePoints:     0,
		}
	default:
		return c.Status(403).JSON(fiber.Map{"error": "Access denied"})
	}

	if errStat != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to get statistics",
			"details": errStat.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data": statistics,
	})
}

// GetAchievementsByRole - List achievements filtered by role (untuk GET /achievements)
func (s *AchievementService) GetAchievementsByRole(c *fiber.Ctx) error {
	ctx := c.UserContext()

	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	currentUser, err := s.userRepo.GetByID(userID)
	if err != nil || currentUser == nil {
		return c.Status(401).JSON(fiber.Map{"error": "User not found"})
	}

	// Get role
	role, err := s.roleRepo.GetByID(currentUser.RoleID)
	if err != nil || role == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get user role"})
	}

	// Get query parameters
	status := c.Query("status", "")
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// Kecuali jika secara eksplisit meminta status "deleted"
	var filterStatus string
	if status == "deleted" {
		// Hanya tampilkan deleted jika diminta khusus
		filterStatus = "deleted"
	} else if status != "" {
		// Status tertentu (draft, submitted, etc) - exclude deleted
		filterStatus = status
	} else {
		// Tanpa filter status - exclude deleted secara default
		filterStatus = "" // Repository akan handle exclude deleted
	}

	var refs []*models.AchievementReference
	var total int
	var errQuery error

	switch role.Name {
	case "Mahasiswa":
		student, _ := s.studentRepo.GetByUserID(userID)
		if student == nil {
			return c.Status(403).JSON(fiber.Map{"error": "User is not a student"})
		}
		refs, total, errQuery = s.achievementRefRepo.FindByStudentID(ctx, student.ID, filterStatus, page, limit)

	case "Dosen Wali":
		// Dosen hanya bisa lihat prestasi mahasiswa bimbingannya
		lecturer, _ := s.lecturerRepo.GetByUserID(userID)
		if lecturer == nil {
			return c.Status(403).JSON(fiber.Map{"error": "User is not a lecturer"})
		}
		refs, total, errQuery = s.achievementRefRepo.FindByAdvisorID(ctx, lecturer.ID, filterStatus, page, limit)

	case "Admin":
		// Admin bisa lihat semua prestasi (termasuk deleted jika diminta)
		refs, total, errQuery = s.achievementRefRepo.FindAll(ctx, filterStatus, page, limit)

	default:
		return c.Status(403).JSON(fiber.Map{"error": "Access denied"})
	}

	if errQuery != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to get achievements",
			"details": errQuery.Error(),
		})
	}

	// Get achievement details from MongoDB
	var achievements []fiber.Map
	for _, ref := range refs {
		mongoID, err := primitive.ObjectIDFromHex(ref.MongoAchievementID)
		if err != nil {
			continue
		}

		achievement, err := s.achievementRepo.FindByID(ctx, mongoID)
		if err != nil {
			continue
		}

		// Get student info
		student, _ := s.studentRepo.GetByID(ref.StudentID)
		var studentName, studentNIM string
		if student != nil {
			studentUser, _ := s.userRepo.GetByID(student.UserID)
			if studentUser != nil {
				studentName = studentUser.FullName
			}
			studentNIM = student.StudentID
		}

		achievements = append(achievements, fiber.Map{
			"id":           ref.ID,
			"status":       ref.Status,
			"title":        achievement.Title,
			"type":         achievement.AchievementType,
			"points":       achievement.Points,
			"submitted_at": ref.SubmittedAt,
			"verified_at":  ref.VerifiedAt,
			"created_at":   ref.CreatedAt,
			"student": fiber.Map{
				"id":   student.ID,
				"name": studentName,
				"nim":  studentNIM,
			},
		})
	}

	totalPages := (total + limit - 1) / limit
	hasNext := page < totalPages
	hasPrev := page > 1

	return c.JSON(fiber.Map{
		"success": true,
		"data": achievements,
		"pagination": fiber.Map{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": totalPages,
			"has_next":    hasNext,
			"has_prev":    hasPrev,
		},
	})
}
func (s *AchievementService) GetAchievementHistory(c *fiber.Ctx) error {
	ctx := c.UserContext()
	achievementID := c.Params("id")

	refUUID, err := uuid.Parse(achievementID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid achievement ID"})
	}

	// Get reference
	ref, err := s.achievementRefRepo.FindByID(ctx, refUUID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get achievement"})
	}
	if ref == nil {
		return c.Status(404).JSON(fiber.Map{"error": "Achievement not found"})
	}

	userID, _ := c.Locals("user_id").(uuid.UUID)
	currentUser, _ := s.userRepo.GetByID(userID)
	if currentUser == nil {
		return c.Status(401).JSON(fiber.Map{"error": "User not found"})
	}

	role, _ := s.roleRepo.GetByID(currentUser.RoleID)
	hasAccess := false

	if role != nil {
		switch role.Name {
		case "Mahasiswa":
			student, _ := s.studentRepo.GetByUserID(userID)
			if student != nil && student.ID == ref.StudentID {
				hasAccess = true
			}
		case "Dosen Wali":
			studentIDs, _ := s.achievementRefRepo.GetStudentIDsByAdvisor(ctx, userID)
			for _, sid := range studentIDs {
				if sid == ref.StudentID {
					hasAccess = true
					break
				}
			}
		case "Admin":
			hasAccess = true
		}
	}

	if !hasAccess {
		return c.Status(403).JSON(fiber.Map{"error": "Access denied"})
	}

	// Create history from reference data
	history := []fiber.Map{
		{
			"status":    "draft",
			"timestamp": ref.CreatedAt,
			"note":      "Achievement created",
		},
	}

	if ref.SubmittedAt != nil {
		history = append(history, fiber.Map{
			"status":    "submitted",
			"timestamp": *ref.SubmittedAt,
			"note":      "Submitted for verification",
		})
	}

	if ref.VerifiedAt != nil {
		action := "verified"
		note := "Achievement verified"
		if ref.Status == "rejected" {
			action = "rejected"
			note = fmt.Sprintf("Achievement rejected: %s", *ref.RejectionNote)
		}

		history = append(history, fiber.Map{
			"status":      action,
			"timestamp":   *ref.VerifiedAt,
			"verified_by": ref.VerifiedBy,
			"note":        note,
		})
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"achievement_id": ref.ID,
			"current_status": ref.Status,
			"history":        history,
		},
	})
}

func (s *AchievementService) UploadAttachments(c *fiber.Ctx) error {
	ctx := c.UserContext()
	achievementID := c.Params("id")

	refUUID, err := uuid.Parse(achievementID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid achievement ID"})
	}

	// Get reference
	ref, err := s.achievementRefRepo.FindByID(ctx, refUUID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get achievement"})
	}
	if ref == nil {
		return c.Status(404).JSON(fiber.Map{"error": "Achievement not found"})
	}

	// Check permissions - hanya owner yang bisa upload
	userID, _ := c.Locals("user_id").(uuid.UUID)
	student, _ := s.studentRepo.GetByUserID(userID)
	if student == nil || student.ID != ref.StudentID {
		return c.Status(403).JSON(fiber.Map{"error": "Access denied"})
	}

	// Check if achievement is still draft
	if ref.Status != "draft" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Attachments can only be added to draft achievements",
		})
	}

	// Parse multipart form
	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid form data"})
	}

	files := form.File["attachments"]
	if len(files) == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "No files uploaded"})
	}

	// Get achievement from MongoDB
	mongoID, err := primitive.ObjectIDFromHex(ref.MongoAchievementID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Invalid MongoDB ID"})
	}

	achievement, err := s.achievementRepo.FindByID(ctx, mongoID)
	if err != nil || achievement == nil {
		return c.Status(404).JSON(fiber.Map{"error": "Achievement details not found"})
	}

	// Process uploaded files
	var uploadedAttachments []fiber.Map
	var newAttachments []models.Attachment

	for _, file := range files {
		// Simpan file (implementasi actual storage)
		// Ini contoh sederhana - sesuaikan dengan storage system kamu
		filePath := fmt.Sprintf("uploads/achievements/%s/%s", ref.ID.String(), file.Filename)

		// Simpan file ke filesystem
		if err := c.SaveFile(file, filePath); err != nil {
			continue // Skip file yang gagal
		}

		// Create attachment object
		attachment := models.Attachment{
			ID:         uuid.New(),
			FileName:   file.Filename,
			FileURL:    "/" + filePath, // atau URL public
			FileType:   file.Header.Get("Content-Type"),
			FileSize:   file.Size,
			UploadedAt: time.Now(),
		}

		newAttachments = append(newAttachments, attachment)

		uploadedAttachments = append(uploadedAttachments, fiber.Map{
			"id":        attachment.ID,
			"file_name": attachment.FileName,
			"file_url":  attachment.FileURL,
			"file_size": attachment.FileSize,
			"file_type": attachment.FileType,
		})
	}

	// Update achievement with new attachments
	achievement.Attachments = append(achievement.Attachments, newAttachments...)

	if err := s.achievementRepo.Update(ctx, mongoID, achievement); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to update achievement attachments",
			"details": err.Error(),
		})
	}

	// Update reference timestamp
	if err := s.achievementRefRepo.UpdateStatus(ctx, ref.ID, "draft", nil, nil); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to update achievement reference",
		})
	}

	return c.JSON(fiber.Map{
		"message":     "Files uploaded successfully",
		"attachments": uploadedAttachments,
		"total_files": len(uploadedAttachments),
	})
}

// Helper function to check if a string exists in a slice
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
