package service

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
	roleRepo repository.RoleRepository, 
) *AchievementService {
	return &AchievementService{
		achievementRepo:    achievementRepo,
		achievementRefRepo: achievementRefRepo,
		studentRepo:        studentRepo,
		lecturerRepo:       lecturerRepo,
		userRepo:           userRepo,
		roleRepo:           roleRepo, 
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
			"id":               mongoID.Hex(),     
			"reference_id":     ref.ID,              
			"student_id":       studentID,           
			"achievement_type": req.AchievementType, 
			"title":            req.Title,           
			"description":      req.Description,     
			"status":           ref.Status,          
			"points":           req.Points,          
			"created_at":       ref.CreatedAt,       
			"created_by":       user.ID,             
		},
	})
}

func (s *AchievementService) GetAchievementByID(c *fiber.Ctx) error {
	ctx := c.UserContext()
	
	refID := c.Params("id") 
	
	refUUID, err := uuid.Parse(refID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid achievement ID"})
	}

	// Get reference dari PostgreSQL
	ref, err := s.achievementRefRepo.FindByID(ctx, refUUID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get achievement"})
	}
	if ref == nil {
		return c.Status(404).JSON(fiber.Map{"error": "Achievement not found"})
	}

	// Get achievement dari MongoDB
	mongoID, err := primitive.ObjectIDFromHex(ref.MongoAchievementID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Invalid MongoDB ID in reference"})
	}

	achievement, err := s.achievementRepo.FindByID(ctx, mongoID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get achievement details"})
	}
	if achievement == nil {
		return c.Status(404).JSON(fiber.Map{"error": "Achievement details not found"})
	}

	// Get user info untuk validasi access
	userID, _ := c.Locals("user_id").(uuid.UUID)
	user, _ := c.Locals("user").(*models.User)
	userRole, err := s.roleRepo.GetByID(user.RoleID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get user role"})
	}

	canAccess := false
	
	switch userRole.Name {
	case "Admin":
		// Admin bisa akses semua
		canAccess = true
		
	case "Mahasiswa":
		// Mahasiswa hanya bisa akses miliknya sendiri
		student, _ := s.studentRepo.GetByUserID(userID)
		if student != nil && student.ID == ref.StudentID {
			canAccess = true
		}
		
	case "Dosen Wali":
		// Dosen hanya bisa akses mahasiswa bimbingannya
		lecturer, _ := s.lecturerRepo.GetByUserID(userID)
		if lecturer != nil {
			// Cek apakah student ini adalah bimbingan dosen
			student, _ := s.studentRepo.GetByID(ref.StudentID)
			if student != nil && student.AdvisorID != nil && *student.AdvisorID == lecturer.ID {
				canAccess = true
			}
		}
	}
	
	if !canAccess {
		return c.Status(403).JSON(fiber.Map{"error": "Access denied"})
	}

	student, _ := s.studentRepo.GetByID(ref.StudentID)
	var studentInfo fiber.Map
	if student != nil {
		studentUser, _ := s.userRepo.GetByID(student.UserID)
		studentInfo = fiber.Map{
			"id":   student.ID,
			"nim":  student.StudentID,
			"name": studentUser.FullName,
			"program_study": student.ProgramStudy,
			"academic_year": student.AcademicYear,
		}
	}

	var verifiedByInfo fiber.Map
	if ref.VerifiedBy != nil {
		verifiedUser, _ := s.userRepo.GetByID(*ref.VerifiedBy)
		if verifiedUser != nil {
			verifiedByInfo = fiber.Map{
				"id":   verifiedUser.ID,
				"name": verifiedUser.FullName,
			}
		}
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			// IDs
			"id":       ref.ID,                 
			"mongo_id": mongoID.Hex(),          
			
			// Achievement data
			"achievement_type": achievement.AchievementType,
			"title":           achievement.Title,
			"description":     achievement.Description,
			"points":         achievement.Points,
			"tags":           achievement.Tags,
			"details":        achievement.Details,
			"attachments":    achievement.Attachments,
			
			// Status info
			"status":         ref.Status,
			"submitted_at":   ref.SubmittedAt,
			"verified_at":    ref.VerifiedAt,
			"verified_by":    verifiedByInfo,
			"rejection_note": ref.RejectionNote,
			
			// Student info
			"student":       studentInfo,
			"student_id":    ref.StudentID,
			
			// Timestamps
			"created_at":    ref.CreatedAt,
			"updated_at":    ref.UpdatedAt,
		},
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

	refs, total, err := s.achievementRefRepo.FindByStudentID(ctx, student.ID, status, page, limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to get achievements",
			"details": err.Error(),
		})
	}

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
	
	refID := c.Params("id") 
	
	// Parse sebagai UUID
	refUUID, err := uuid.Parse(refID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid achievement ID"})
	}

	// Get reference dari PostgreSQL
	ref, err := s.achievementRefRepo.FindByID(ctx, refUUID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get achievement"})
	}
	if ref == nil {
		return c.Status(404).JSON(fiber.Map{"error": "Achievement not found"})
	}

	// Get achievement dari MongoDB
	mongoID, err := primitive.ObjectIDFromHex(ref.MongoAchievementID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Invalid MongoDB ID in reference"})
	}

	achievement, err := s.achievementRepo.FindByID(ctx, mongoID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get achievement details"})
	}
	if achievement == nil {
		return c.Status(404).JSON(fiber.Map{"error": "Achievement details not found"})
	}

	// Get user info
	userID, _ := c.Locals("user_id").(uuid.UUID)
	user, _ := c.Locals("user").(*models.User)
	
	// Get user role
	userRole, err := s.roleRepo.GetByID(user.RoleID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get user role"})
	}

	// Ownership check berdasarkan role
	switch userRole.Name {
case "Mahasiswa":
		// Mahasiswa hanya bisa update miliknya sendiri
		student, _ := s.studentRepo.GetByUserID(userID)
		if student == nil {
			return c.Status(403).JSON(fiber.Map{"error": "Student not found"})
		}
		
		if student.ID != ref.StudentID {
			return c.Status(403).JSON(fiber.Map{"error": "Not your achievement"})
		}
	case "Admin":
		// Admin bisa update semua
		// Tidak perlu validasi ownership
	default:
		// Dosen & lainnya tidak boleh update
		return c.Status(403).JSON(fiber.Map{"error": "Unauthorized role"})
	}

	// Status check (hanya draft yang bisa diupdate)
	if ref.Status != "draft" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Only draft achievements can be updated",
			"current_status": ref.Status,
		})
	}

	// Parse request body sebagai map
	var req map[string]interface{}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Apply updates
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
		
		// Update event info
		if loc, ok := details["location"].(string); ok && loc != "" {
			achievement.Details.Location = loc
		}
		if org, ok := details["organizer"].(string); ok && org != "" {
			achievement.Details.Organizer = org
		}
		if score, ok := details["score"].(float64); ok {
			achievement.Details.Score = int(score)
		}
		
		// Update custom fields
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

	// Update di MongoDB
	if err := s.achievementRepo.Update(ctx, mongoID, achievement); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update achievement"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Achievement updated successfully",
		"data": fiber.Map{
			"id":         ref.ID,                 
			"mongo_id":   mongoID.Hex(),          
			"status":     ref.Status,
			"updated_at": achievement.UpdatedAt,
		},
	})
}

func (s *AchievementService) DeleteAchievement(c *fiber.Ctx) error {
	ctx := c.UserContext()
	
	refID := c.Params("id") 
	
	refUUID, err := uuid.Parse(refID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid achievement ID"})
	}
	// Get reference dari PostgreSQL
	ref, err := s.achievementRefRepo.FindByID(ctx, refUUID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get achievement"})
	}
	if ref == nil {
		return c.Status(404).JSON(fiber.Map{"error": "Achievement not found"})
	}

	// Get user info
	userID, _ := c.Locals("user_id").(uuid.UUID)
	user, _ := c.Locals("user").(*models.User)
	userRole, err := s.roleRepo.GetByID(user.RoleID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get user role"})
	}

	// Ownership check berdasarkan role
	switch userRole.Name {
case "Mahasiswa":
		// Mahasiswa hanya bisa delete miliknya sendiri
		student, _ := s.studentRepo.GetByUserID(userID)
		if student == nil {
			return c.Status(403).JSON(fiber.Map{"error": "Student not found"})
		}
		
		if student.ID != ref.StudentID {
			return c.Status(403).JSON(fiber.Map{"error": "Not your achievement"})
		}
	case "Admin":
	default:
		return c.Status(403).JSON(fiber.Map{"error": "Unauthorized role"})
	}

	// Status check (hanya draft yang bisa di-delete)
	if ref.Status != "draft" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Only draft achievements can be deleted",
			"current_status": ref.Status,
		})
	}

	// SOFT DELETE: Update status ke 'deleted' di PostgreSQL
	if err := s.achievementRefRepo.SoftDelete(ctx, ref.ID); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to delete achievement",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Achievement deleted successfully",
		"data": fiber.Map{
			"id":             ref.ID,
			"mongo_id":       ref.MongoAchievementID,
			"previous_status": ref.Status,
			"new_status":     "deleted",
			"deleted_at":     time.Now(),
		},
	})
}

func (s *AchievementService) SubmitAchievement(c *fiber.Ctx) error {
	ctx := c.UserContext()
	
	refID := c.Params("id")
	refUUID, _ := uuid.Parse(refID)

	userID, _ := c.Locals("user_id").(uuid.UUID)
	user, _ := c.Locals("user").(*models.User)

	// Dapatkan role user dari repository
	userRole, err := s.roleRepo.GetByID(user.RoleID)
	if err != nil || userRole == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get user role"})
	}

	// Get reference
	ref, err := s.achievementRefRepo.FindByID(ctx, refUUID)
	if err != nil || ref == nil {
		return c.Status(404).JSON(fiber.Map{"error": "Achievement not found"})
	}

	// Cek status
	if ref.Status != "draft" {
		return c.Status(400).JSON(fiber.Map{
			"error": fmt.Sprintf("Only draft achievements can be submitted. Current: %s", ref.Status),
		})
	}

	// Jika Mahasiswa, cek ownership
	if userRole.Name == "Mahasiswa" {
		student, _ := s.studentRepo.GetByUserID(userID)
		if student == nil || student.ID != ref.StudentID {
			return c.Status(403).JSON(fiber.Map{"error": "Not your achievement"})
		}
	}
	// Admin tidak perlu validasi ownership

	// Submit
	if err := s.achievementRefRepo.SubmitForVerification(ctx, refUUID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to submit achievement"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Achievement submitted",
		"data": fiber.Map{
			"id":         ref.ID,
			"new_status": "submitted",
			"submitted_at": time.Now(),
		},
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

func (s *AchievementService) VerifyAchievement(c *fiber.Ctx) error {
	ctx := c.UserContext()
	
	refID := c.Params("id")
	refUUID, _ := uuid.Parse(refID)

	userID, _ := c.Locals("user_id").(uuid.UUID)
	user, _ := c.Locals("user").(*models.User)

	// Dapatkan role user
	userRole, err := s.roleRepo.GetByID(user.RoleID)
	if err != nil || userRole == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get user role"})
	}

	// Get reference
	ref, err := s.achievementRefRepo.FindByID(ctx, refUUID)
	if err != nil || ref == nil {
		return c.Status(404).JSON(fiber.Map{"error": "Achievement not found"})
	}

	// Cek status
	if ref.Status != "submitted" {
		return c.Status(400).JSON(fiber.Map{
			"error": fmt.Sprintf("Only submitted achievements can be verified. Current: %s", ref.Status),
		})
	}

	// Jika user adalah Dosen, cek apakah mahasiswa bimbingannya
	if userRole.Name == "Dosen Wali" {
		lecturer, _ := s.lecturerRepo.GetByUserID(userID)
		student, _ := s.studentRepo.GetByID(ref.StudentID)
		
		if lecturer == nil || student == nil || student.AdvisorID == nil || *student.AdvisorID != lecturer.ID {
			return c.Status(403).JSON(fiber.Map{
				"error": "You can only verify achievements of your advisees",
			})
		}
	}

	// Verify
	if err := s.achievementRefRepo.VerifyAchievement(ctx, refUUID, userID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to verify achievement"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Achievement verified",
		"data": fiber.Map{
			"id":         ref.ID,
			"new_status": "verified",
			"verified_by": userID,
			"verified_at": time.Now(),
		},
	})
}

func (s *AchievementService) RejectAchievement(c *fiber.Ctx) error {
	ctx := c.UserContext()
	
	refID := c.Params("id")
	refUUID, _ := uuid.Parse(refID)

	userID, _ := c.Locals("user_id").(uuid.UUID)
	user, _ := c.Locals("user").(*models.User)

	// Dapatkan role user
	userRole, err := s.roleRepo.GetByID(user.RoleID)
	if err != nil || userRole == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get user role"})
	}

	// Parse rejection note
	var req struct {
		RejectionNote string `json:"rejection_note"`
	}
	if err := c.BodyParser(&req); err != nil || req.RejectionNote == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Rejection note is required"})
	}

	// Get reference
	ref, err := s.achievementRefRepo.FindByID(ctx, refUUID)
	if err != nil || ref == nil {
		return c.Status(404).JSON(fiber.Map{"error": "Achievement not found"})
	}

	// Cek status
	if ref.Status != "submitted" {
		return c.Status(400).JSON(fiber.Map{
			"error": fmt.Sprintf("Only submitted achievements can be rejected. Current: %s", ref.Status),
		})
	}

	// Jika Dosen, cek advisor
	if userRole.Name == "Dosen Wali" {
		lecturer, _ := s.lecturerRepo.GetByUserID(userID)
		student, _ := s.studentRepo.GetByID(ref.StudentID)
		
		if lecturer == nil || student == nil || student.AdvisorID == nil || *student.AdvisorID != lecturer.ID {
			return c.Status(403).JSON(fiber.Map{
				"error": "You can only reject achievements of your advisees",
			})
		}
	}

	// Reject
	if err := s.achievementRefRepo.RejectAchievement(ctx, refUUID, userID, req.RejectionNote); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to reject achievement"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Achievement rejected",
		"data": fiber.Map{
			"id":             ref.ID,
			"new_status":     "rejected",
			"rejection_note": req.RejectionNote,
			"rejected_by":    userID,
			"rejected_at":    time.Now(),
		},
	})
}

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
	
	// Ambil Reference ID dari parameter (UUID)
	refID := c.Params("id") // "550e8400-e29b-41d3-a456-426614174000"
	
	// Parse sebagai UUID
	refUUID, err := uuid.Parse(refID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid achievement ID"})
	}

	// Get reference dari PostgreSQL
	ref, err := s.achievementRefRepo.FindByID(ctx, refUUID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get achievement"})
	}
	if ref == nil {
		return c.Status(404).JSON(fiber.Map{"error": "Achievement not found"})
	}

	// Get user info
	userID, _ := c.Locals("user_id").(uuid.UUID)
	user, _ := c.Locals("user").(*models.User)
	
	// Get user role
	userRole, err := s.roleRepo.GetByID(user.RoleID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get user role"})
	}

	// Access control berdasarkan role
	hasAccess := false
	
	switch userRole.Name {
case "Admin":
		// Admin bisa akses semua
		hasAccess = true
	case "Mahasiswa":
		// Mahasiswa hanya bisa akses miliknya sendiri
		student, _ := s.studentRepo.GetByUserID(userID)
		if student != nil && student.ID == ref.StudentID {
			hasAccess = true
		}
	case "Dosen Wali":
		// Dosen hanya bisa akses mahasiswa bimbingannya
		lecturer, _ := s.lecturerRepo.GetByUserID(userID)
		if lecturer != nil {
			student, _ := s.studentRepo.GetByID(ref.StudentID)
			if student != nil && student.AdvisorID != nil && *student.AdvisorID == lecturer.ID {
				hasAccess = true
			}
		}
	}

	if !hasAccess {
		return c.Status(403).JSON(fiber.Map{"error": "Access denied"})
	}

	// Create history dari reference data
	history := []fiber.Map{
		{
			"status":    "draft",
			"timestamp": ref.CreatedAt,
			"note":      "Achievement created",
		},
	}

	// Add submitted event jika ada
	if ref.SubmittedAt != nil {
		history = append(history, fiber.Map{
			"status":    "submitted",
			"timestamp": *ref.SubmittedAt,
			"note":      "Submitted for verification",
		})
	}

	// Add verified/rejected event jika ada
	if ref.VerifiedAt != nil {
		action := "verified"
		note := "Achievement verified"
		
		if ref.Status == "rejected" {
			action = "rejected"
			note = fmt.Sprintf("Achievement rejected: %s", *ref.RejectionNote)
		}

		// Get verified by user info
		var verifiedByName string
		if ref.VerifiedBy != nil {
			verifiedUser, _ := s.userRepo.GetByID(*ref.VerifiedBy)
			if verifiedUser != nil {
				verifiedByName = verifiedUser.FullName
			}
		}

		history = append(history, fiber.Map{
			"status":           action,
			"timestamp":        *ref.VerifiedAt,
			"verified_by":      ref.VerifiedBy,
			"verified_by_name": verifiedByName,
			"note":             note,
		})
	}

	// Add deleted event jika status deleted
	if ref.Status == "deleted" {
		history = append(history, fiber.Map{
			"status":    "deleted",
			"timestamp": ref.UpdatedAt, // updated_at saat di-delete
			"note":      "Achievement deleted",
		})
	}

	// Get achievement data untuk response
	mongoID, _ := primitive.ObjectIDFromHex(ref.MongoAchievementID)
	achievement, _ := s.achievementRepo.FindByID(ctx, mongoID)
	
	var title string
	if achievement != nil {
		title = achievement.Title
	}

	// Get student info
	student, _ := s.studentRepo.GetByID(ref.StudentID)
	var studentInfo fiber.Map
	if student != nil {
		studentUser, _ := s.userRepo.GetByID(student.UserID)
		studentInfo = fiber.Map{
			"id":   student.ID,
			"name": studentUser.FullName,
			"nim":  student.StudentID,
		}
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"achievement_id":   ref.ID,
			"mongo_id":         ref.MongoAchievementID,
			"title":            title,
			"current_status":   ref.Status,
			"student":          studentInfo,
			"total_history":    len(history),
			"history":          history,
		},
	})
}

func (s *AchievementService) UploadAttachments(c *fiber.Ctx) error {
	ctx := c.UserContext()
	
	refID := c.Params("id")
	refUUID, _ := uuid.Parse(refID)

	userID, _ := c.Locals("user_id").(uuid.UUID)
	user, _ := c.Locals("user").(*models.User)

	// Dapatkan role user dari repository
	userRole, err := s.roleRepo.GetByID(user.RoleID)
	if err != nil || userRole == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get user role"})
	}

	// Get reference
	ref, err := s.achievementRefRepo.FindByID(ctx, refUUID)
	if err != nil || ref == nil {
		return c.Status(404).JSON(fiber.Map{"error": "Achievement not found"})
	}

	// Cek status
	if ref.Status != "draft" {
		return c.Status(400).JSON(fiber.Map{
			"error": fmt.Sprintf("Only draft achievements can have attachments uploaded. Current: %s", ref.Status),
		})
	}

	// Jika Mahasiswa, cek ownership
	if userRole.Name == "Mahasiswa" {
		student, _ := s.studentRepo.GetByUserID(userID)
		if student == nil || student.ID != ref.StudentID {
			return c.Status(403).JSON(fiber.Map{"error": "Not your achievement"})
		}
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
		cleanFileName := filepath.Base(file.Filename)
		var safeFileNameBuilder strings.Builder
		for _, r := range cleanFileName {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || 
			   (r >= '0' && r <= '9') || r == '.' || r == '-' || r == '_' {
				safeFileNameBuilder.WriteRune(r)
			} else if r == ' ' {
				safeFileNameBuilder.WriteRune('_') // ganti spasi dengan underscore
			} else {
				safeFileNameBuilder.WriteRune('_') // ganti karakter lain dengan underscore
			}
		}
		safeFileName := safeFileNameBuilder.String()
		
		if safeFileName == "" {
			safeFileName = "file_" + uuid.New().String()[:8]
		}

		// 3. Generate unique filename
		uniqueFilename := fmt.Sprintf("%s_%s_%s",
			time.Now().Format("20060102_150405"),
			uuid.New().String()[:8],
			safeFileName,
		)

		// Define file path
		filePath := fmt.Sprintf("uploads/achievements/%s/%s",
			ref.ID.String(),
			uniqueFilename,
		)

		// Create directory if not exists
		uploadDir := filepath.Dir(filePath)
		if err := os.MkdirAll(uploadDir, 0755); err != nil {
			continue // Skip file yang gagal
		}

		// Save file to filesystem
		if err := c.SaveFile(file, filePath); err != nil {
			continue // Skip file yang gagal
		}

		contentType := file.Header.Get("Content-Type")
		var cleanContentType strings.Builder
		for _, r := range contentType {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || 
			   (r >= '0' && r <= '9') || r == '/' || r == '-' || r == '+' || r == '.' {
				cleanContentType.WriteRune(r)
			}
		}
		if cleanContentType.Len() == 0 {
			cleanContentType.WriteString("application/octet-stream")
		}

		// Create attachment object
		attachment := models.Attachment{
			ID:         uuid.New(),
			FileName:   safeFileName, 
			FileURL:    "/" + filePath,
			FileType:   cleanContentType.String(), 
			FileSize:   file.Size,
			UploadedAt: time.Now(),
		}

		newAttachments = append(newAttachments, attachment)

		uploadedAttachments = append(uploadedAttachments, fiber.Map{
			"id":         attachment.ID,
			"file_name":  attachment.FileName,
			"file_url":   attachment.FileURL,
			"file_size":  attachment.FileSize,
			"file_type":  attachment.FileType,
			"uploaded_at": attachment.UploadedAt,
		})
	}

	if len(newAttachments) == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "No files were successfully uploaded"})
	}

	// Update achievement with new attachments
	achievement.Attachments = append(achievement.Attachments, newAttachments...)

	if err := s.achievementRepo.Update(ctx, mongoID, achievement); err != nil {
		// Cek apakah error karena UTF-8
		if strings.Contains(err.Error(), "UTF-8") {
			// Coba cleanup data sebelum save ulang
			for i := range achievement.Attachments {
				// Pastikan semua string field bersih
				achievement.Attachments[i].FileName = cleanString(achievement.Attachments[i].FileName)
				achievement.Attachments[i].FileType = cleanString(achievement.Attachments[i].FileType)
				achievement.Attachments[i].FileURL = cleanString(achievement.Attachments[i].FileURL)
			}
			
			// Coba save lagi
			if err := s.achievementRepo.Update(ctx, mongoID, achievement); err != nil {
				return c.Status(500).JSON(fiber.Map{
					"error": "Failed to update achievement attachments after cleanup",
				})
			}
		} else {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to update achievement attachments",
			})
		}
	}

	// Update reference timestamp
	if err := s.achievementRefRepo.UpdateStatus(ctx, ref.ID, "draft", nil, nil); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update achievement reference"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": fmt.Sprintf("%d file(s) uploaded successfully", len(uploadedAttachments)),
		"data": fiber.Map{
			"id":              ref.ID,
			"new_attachments": uploadedAttachments,
			"total_files":     len(uploadedAttachments),
			"uploaded_at":     time.Now(),
		},
	})
}

// Simple clean string function (tetap dalam scope yang sama)
func cleanString(s string) string {
	var result strings.Builder
	for _, r := range s {
		// Hanya izinkan karakter ASCII yang aman
		if r >= 32 && r <= 126 { // Printable ASCII characters
			result.WriteRune(r)
		} else {
			result.WriteRune('_')
		}
	}
	return result.String()
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
