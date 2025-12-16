package service

import (
	"context"
	"strconv"

	"achievement-backend/app/models"
	"achievement-backend/app/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type StudentLecturerService struct {
	studentRepo        repository.StudentRepository
	lecturerRepo       repository.LecturerRepository
	userRepo           repository.UserRepository
	roleRepo           repository.RoleRepository
	achievementRepo    repository.AchievementRepository
	achievementRefRepo repository.AchievementReferenceRepository
}

func NewStudentLecturerService(
	studentRepo repository.StudentRepository,
	lecturerRepo repository.LecturerRepository,
	userRepo repository.UserRepository,
	roleRepo	 repository.RoleRepository,
	achievementRepo repository.AchievementRepository,
	achievementRefRepo repository.AchievementReferenceRepository,
) *StudentLecturerService {
	return &StudentLecturerService{
		studentRepo:        studentRepo,
		lecturerRepo:       lecturerRepo,
		userRepo:           userRepo,
		achievementRepo:    achievementRepo,
		achievementRefRepo: achievementRefRepo,
		roleRepo: 			roleRepo,
	}
}

func (s *StudentLecturerService) GetAllStudents(c *fiber.Ctx) error {
	// Get query parameters
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	search := c.Query("search", "")
	programStudy := c.Query("program_study", "")
	academicYear := c.Query("academic_year", "")
	hasAdvisor := c.Query("has_advisor", "")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	var students []models.StudentResponse
	var total int
	var errQuery error

	if search != "" {
		students, total, errQuery = s.studentRepo.SearchByName(search, page, limit)
	} else if programStudy != "" {
		rawStudents, count, err := s.studentRepo.GetByProgramStudy(programStudy, page, limit)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to get students", "details": err.Error()})
		}

		// Convert to StudentResponse
		for _, rawStudent := range rawStudents {
			user, _ := s.userRepo.GetByID(rawStudent.UserID)
			if user != nil {
				students = append(students, models.StudentResponse{
					ID:           rawStudent.ID,
					UserID:       rawStudent.UserID,
					FullName:     user.FullName,
					Email:        user.Email,
					Username:     user.Username,
					StudentID:    rawStudent.StudentID,
					ProgramStudy: rawStudent.ProgramStudy,
					AcademicYear: rawStudent.AcademicYear,
					AdvisorID:    rawStudent.AdvisorID,
					CreatedAt:    rawStudent.CreatedAt,
				})
			}
		}
		total = count
	} else if academicYear != "" {
		rawStudents, count, err := s.studentRepo.GetByAcademicYear(academicYear, page, limit)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to get students", "details": err.Error()})
		}

		for _, rawStudent := range rawStudents {
			user, _ := s.userRepo.GetByID(rawStudent.UserID)
			if user != nil {
				students = append(students, models.StudentResponse{
					ID:           rawStudent.ID,
					UserID:       rawStudent.UserID,
					FullName:     user.FullName,
					Email:        user.Email,
					Username:     user.Username,
					StudentID:    rawStudent.StudentID,
					ProgramStudy: rawStudent.ProgramStudy,
					AcademicYear: rawStudent.AcademicYear,
					AdvisorID:    rawStudent.AdvisorID,
					CreatedAt:    rawStudent.CreatedAt,
				})
			}
		}
		total = count
	} else if hasAdvisor != "" {
		hasAdvisorBool, err := strconv.ParseBool(hasAdvisor)
		if err == nil {
			if !hasAdvisorBool {
				rawStudents, count, err := s.studentRepo.GetAdvisorless(page, limit)
				if err != nil {
					return c.Status(500).JSON(fiber.Map{"error": "Failed to get students", "details": err.Error()})
				}

				for _, rawStudent := range rawStudents {
					user, _ := s.userRepo.GetByID(rawStudent.UserID)
					if user != nil {
						students = append(students, models.StudentResponse{
							ID:           rawStudent.ID,
							UserID:       rawStudent.UserID,
							FullName:     user.FullName,
							Email:        user.Email,
							Username:     user.Username,
							StudentID:    rawStudent.StudentID,
							ProgramStudy: rawStudent.ProgramStudy,
							AcademicYear: rawStudent.AcademicYear,
							AdvisorID:    rawStudent.AdvisorID,
							CreatedAt:    rawStudent.CreatedAt,
						})
					}
				}
				total = count
			} else {
				students, total, errQuery = s.studentRepo.GetWithUserDetails(page, limit)
			}
		} else {
			students, total, errQuery = s.studentRepo.GetWithUserDetails(page, limit)
		}
	} else {
		students, total, errQuery = s.studentRepo.GetWithUserDetails(page, limit)
	}

	if errQuery != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to get students",
			"details": errQuery.Error(),
		})
	}

	// Tambahkan advisor name jika ada
	for i := range students {
		if students[i].AdvisorID != nil && *students[i].AdvisorID != uuid.Nil {
			lecturer, _ := s.lecturerRepo.GetByID(*students[i].AdvisorID)
			if lecturer != nil {
				advisorUser, _ := s.userRepo.GetByID(lecturer.UserID)
				if advisorUser != nil {
					students[i].AdvisorName = advisorUser.FullName
				}
			}
		}
	}

	totalPages := (total + limit - 1) / limit
	hasNext := page < totalPages
	hasPrev := page > 1

	return c.JSON(fiber.Map{
		"success": true,
		"data":    students,
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

func (s *StudentLecturerService) GetStudentByID(c *fiber.Ctx) error {
	// Get student ID from params
	studentIDStr := c.Params("id")
	studentID, err := uuid.Parse(studentIDStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid student ID"})
	}

	// Get current user ID
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	// Get current user object
	currentUser, ok := c.Locals("user").(*models.User)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	// Get student data
	student, err := s.studentRepo.GetByID(studentID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get student"})
	}
	if student == nil {
		return c.Status(404).JSON(fiber.Map{"error": "Student not found"})
	}

	// 1. Cek apakah user adalah mahasiswa ini sendiri
	if student.UserID == userID {
		// Mahasiswa akses data sendiri - ALLOWED
	} else {
		// 2. Cek apakah user adalah dosen wali mahasiswa ini
		lecturer, _ := s.lecturerRepo.GetByUserID(userID)
		isDosenWali := lecturer != nil && student.AdvisorID != nil && *student.AdvisorID == lecturer.ID
		
		if !isDosenWali {
			// 3. Cek apakah user adalah Admin
			role, err := s.roleRepo.GetByID(currentUser.RoleID)
			if err != nil || role == nil {
				return c.Status(500).JSON(fiber.Map{"error": "Failed to get user role"})
			}
			
			if role.Name != "Admin" {
				// Bukan Admin, bukan dosen wali, bukan mahasiswa sendiri
				return c.Status(403).JSON(fiber.Map{"error": "Access denied"})
			}
			// Admin - ALLOWED
		}
		// Dosen wali akses mahasiswa bimbingannya - ALLOWED
	}

	// Get user details
	user, err := s.userRepo.GetByID(student.UserID)
	if err != nil || user == nil {
		return c.Status(404).JSON(fiber.Map{"error": "User not found for this student"})
	}

	// Build response
	response := models.StudentResponse{
		ID:           student.ID,
		UserID:       student.UserID,
		FullName:     user.FullName,
		Username:     user.Username,
		Email:        user.Email,
		StudentID:    student.StudentID,
		ProgramStudy: student.ProgramStudy,
		AcademicYear: student.AcademicYear,
		AdvisorID:    student.AdvisorID,
		CreatedAt:    student.CreatedAt,
	}

	// Add advisor name if exists
	if student.AdvisorID != nil && *student.AdvisorID != uuid.Nil {
		lecturer, _ := s.lecturerRepo.GetByID(*student.AdvisorID)
		if lecturer != nil {
			advisorUser, _ := s.userRepo.GetByID(lecturer.UserID)
			if advisorUser != nil {
				response.AdvisorName = advisorUser.FullName
			}
		}
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    response,
	})
}

func (s *StudentLecturerService) GetStudentAchievements(c *fiber.Ctx) error {
	// Get student ID from params
	studentIDStr := c.Params("id")
	studentID, err := uuid.Parse(studentIDStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid student ID"})
	}

	// Get current user ID
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	// Get current user object
	currentUser, ok := c.Locals("user").(*models.User)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	// Get student data
	student, err := s.studentRepo.GetByID(studentID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get student"})
	}
	if student == nil {
		return c.Status(404).JSON(fiber.Map{"error": "Student not found"})
	}

	if student.UserID != userID {
		lecturer, _ := s.lecturerRepo.GetByUserID(userID)
		isDosenWali := lecturer != nil && student.AdvisorID != nil && *student.AdvisorID == lecturer.ID
		
		if !isDosenWali {
			role, err := s.roleRepo.GetByID(currentUser.RoleID)
			if err != nil || role == nil {
				return c.Status(500).JSON(fiber.Map{"error": "Failed to get user role"})
			}
			
			if role.Name != "Admin" {
				return c.Status(403).JSON(fiber.Map{"error": "Access denied"})
			}
		}
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
	ctx := c.UserContext()
	if ctx == nil {
		ctx = context.Background()
	}

	// Get achievements
	refs, total, err := s.achievementRefRepo.FindByStudentID(ctx, studentID, status, page, limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to get student achievements",
			"details": err.Error(),
		})
	}

	// Get achievement details
	var achievements []fiber.Map
	for _, ref := range refs {
		if ref.MongoAchievementID == "" {
			continue
		}

		mongoID, err := primitive.ObjectIDFromHex(ref.MongoAchievementID)
		if err != nil {
			continue
		}

		achievement, err := s.achievementRepo.FindByID(ctx, mongoID)
		if err != nil || achievement == nil {
			continue
		}

		var verifiedByName string
		if ref.VerifiedBy != nil {
			verifiedUser, _ := s.userRepo.GetByID(*ref.VerifiedBy)
			if verifiedUser != nil {
				verifiedByName = verifiedUser.FullName
			}
		}

		achievements = append(achievements, fiber.Map{
			"id":           ref.ID,
			"mongo_id":     mongoID.Hex(),
			"status":       ref.Status,
			"title":        achievement.Title,
			"type":         achievement.AchievementType,
			"points":       achievement.Points,
			"submitted_at": ref.SubmittedAt,
			"verified_at":  ref.VerifiedAt,
			"verified_by": fiber.Map{
				"id":   ref.VerifiedBy,
				"name": verifiedByName,
			},
			"rejection_note": ref.RejectionNote,
			"created_at":     ref.CreatedAt,
			"updated_at":     ref.UpdatedAt,
		})
	}

	totalPages := (total + limit - 1) / limit
	hasNext := page < totalPages
	hasPrev := page > 1

	user, _ := s.userRepo.GetByID(student.UserID)
	var studentName string
	if user != nil {
		studentName = user.FullName
	}

	studentInfo := fiber.Map{
		"id":            student.ID,
		"student_id":    student.StudentID,
		"name":          studentName,
		"program_study": student.ProgramStudy,
		"academic_year": student.AcademicYear,
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"student":      studentInfo,
			"achievements": achievements,
			"total":        total,
			"pagination": fiber.Map{
				"page":        page,
				"limit":       limit,
				"total_pages": totalPages,
				"has_next":    hasNext,
				"has_prev":    hasPrev,
			},
		},
	})
}

func (s *StudentLecturerService) UpdateStudentAdvisor(c *fiber.Ctx) error {
	// Get student ID from params
	studentIDStr := c.Params("id")
	studentID, err := uuid.Parse(studentIDStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid student ID"})
	}

	// Parse request body
	var req struct {
		AdvisorID *string `json:"advisor_id,omitempty"` 
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Check if student exists
	student, err := s.studentRepo.GetByID(studentID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get student"})
	}
	if student == nil {
		return c.Status(404).JSON(fiber.Map{"error": "Student not found"})
	}

	var advisorID uuid.UUID
	var advisorName string

	// Handle advisor assignment/removal
	if req.AdvisorID == nil || *req.AdvisorID == "" {
		// Remove advisor
		if err := s.studentRepo.RemoveAdvisor(studentID); err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error":   "Failed to remove advisor",
				"details": err.Error(),
			})
		}
	} else {
		// Assign new advisor
		advisorID, err = uuid.Parse(*req.AdvisorID)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid advisor ID"})
		}

		// Check if advisor exists and is actually a lecturer
		lecturer, err := s.lecturerRepo.GetByID(advisorID)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to get advisor"})
		}
		if lecturer == nil {
			return c.Status(404).JSON(fiber.Map{"error": "Lecturer not found"})
		}

		// Get advisor user info for response
		advisorUser, _ := s.userRepo.GetByID(lecturer.UserID)
		if advisorUser != nil {
			advisorName = advisorUser.FullName
		}

		// Update advisor
		if err := s.studentRepo.UpdateAdvisor(studentID, advisorID); err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error":   "Failed to update advisor",
				"details": err.Error(),
			})
		}
	}

	// Get student user info for response
	studentUser, _ := s.userRepo.GetByID(student.UserID)

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Student advisor updated successfully",
		"data": fiber.Map{
			"student": fiber.Map{
				"id":            student.ID,
				"student_id":    student.StudentID,
				"name":          studentUser.FullName,
				"program_study": student.ProgramStudy,
				"academic_year": student.AcademicYear,
			},
			"advisor": fiber.Map{
				"id":   advisorID,
				"name": advisorName,
			},
			"action": func() string {
				if req.AdvisorID == nil || *req.AdvisorID == "" {
					return "removed"
				}
				return "assigned"
			}(),
		},
	})
}

func (s *StudentLecturerService) GetAllLecturers(c *fiber.Ctx) error {
	// Get query parameters
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	search := c.Query("search", "")
	department := c.Query("department", "")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// NOTE: Permission sudah di-check di middleware

	var lecturers []models.LecturerResponse
	var total int
	var errQuery error

	// Apply filters
	if search != "" {
		lecturers, total, errQuery = s.lecturerRepo.SearchByName(search, page, limit)
	} else if department != "" {
		rawLecturers, count, err := s.lecturerRepo.GetByDepartment(department, page, limit)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to get lecturers", "details": err.Error()})
		}

		for _, rawLecturer := range rawLecturers {
			user, _ := s.userRepo.GetByID(rawLecturer.UserID)
			if user != nil {
				lecturers = append(lecturers, models.LecturerResponse{
					ID:         rawLecturer.ID,
					UserID:     rawLecturer.UserID,
					FullName:   user.FullName,
					Username:   user.Username,
					Email:      user.Email,
					LecturerID: rawLecturer.LecturerID,
					Department: rawLecturer.Department,
					CreatedAt:  rawLecturer.CreatedAt,
				})
			}
		}
		total = count
	} else {
		lecturers, total, errQuery = s.lecturerRepo.GetWithUserDetails(page, limit)
	}

	if errQuery != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to get lecturers",
			"details": errQuery.Error(),
		})
	}

	var lecturersWithCount []map[string]interface{}

	for _, lecturer := range lecturers {
		studentCount, _ := s.lecturerRepo.GetAdviseesCount(lecturer.ID)

		lecturerMap := map[string]interface{}{
			"id":            lecturer.ID,
			"user_id":       lecturer.UserID,
			"full_name":     lecturer.FullName,
			"username":      lecturer.Username,
			"email":         lecturer.Email,
			"lecturer_id":   lecturer.LecturerID,
			"department":    lecturer.Department,
			"created_at":    lecturer.CreatedAt,
			"student_count": studentCount,
		}

		lecturersWithCount = append(lecturersWithCount, lecturerMap)
	}

	totalPages := (total + limit - 1) / limit
	hasNext := page < totalPages
	hasPrev := page > 1

	return c.JSON(fiber.Map{
		"success": true,
		"data":    lecturersWithCount,
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

func (s *StudentLecturerService) GetLecturerAdvisees(c *fiber.Ctx) error {
	// Get lecturer ID from params
	lecturerIDStr := c.Params("id")
	lecturerID, err := uuid.Parse(lecturerIDStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid lecturer ID"})
	}

	// Get current user ID
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	// Check if lecturer exists
	lecturer, err := s.lecturerRepo.GetByID(lecturerID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to get lecturer",
			"details": err.Error(),
		})
	}
	if lecturer == nil {
		return c.Status(404).JSON(fiber.Map{"error": "Lecturer not found"})
	}

	// Check authorization
	// Jika middleware RBAC sudah mengecek permission, cukup cek ownership
	currentUser, ok := c.Locals("user").(*models.User)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	role, err := s.roleRepo.GetByID(currentUser.RoleID)
	if err != nil || role == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to verify user role"})
	}

	// Admin bisa akses semua, Dosen Wali hanya akses data sendiri
	if role.Name == "Dosen Wali" && lecturer.UserID != userID {
		return c.Status(403).JSON(fiber.Map{
			"error":   "Access denied",
			"details": "You can only access your own advisees",
		})
	}

	// Get query parameters
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// Get advisees for this lecturer
	students, total, err := s.lecturerRepo.GetAdvisees(lecturerID, page, limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to get advisees",
			"details": err.Error(),
		})
	}

	// Build response with user details
	var advisees []fiber.Map
	for _, student := range students {
		user, _ := s.userRepo.GetByID(student.UserID)
		if user != nil {
			// Get achievement statistics for this student
			var statsTotal int
			var verifiedCount int
			
			if stats, err := s.achievementRepo.GetStatisticsByStudent(c.UserContext(), student.ID); err == nil {
				statsTotal = stats.TotalAchievements
			}
			
			if count, err := s.achievementRefRepo.CountByStudentAndStatus(
				c.UserContext(), student.ID, "verified"); err == nil {
				verifiedCount = count
			}

			advisees = append(advisees, fiber.Map{
				"id":            student.ID,
				"student_id":    student.StudentID,
				"name":          user.FullName,
				"email":         user.Email,
				"program_study": student.ProgramStudy,
				"academic_year": student.AcademicYear,
				"created_at":    student.CreatedAt,
				"achievement_stats": fiber.Map{
					"total":    statsTotal,
					"verified": verifiedCount,
				},
			})
		}
	}

	// Get lecturer user info
	lecturerUser, _ := s.userRepo.GetByID(lecturer.UserID)

	totalPages := (total + limit - 1) / limit
	hasNext := page < totalPages
	hasPrev := page > 1

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"lecturer": fiber.Map{
				"id":          lecturer.ID,
				"lecturer_id": lecturer.LecturerID,
				"name":        lecturerUser.FullName,
				"department":  lecturer.Department,
			},
			"advisees":       advisees,
			"total_students": total,
			"pagination": fiber.Map{
				"page":        page,
				"limit":       limit,
				"total_pages": totalPages,
				"has_next":    hasNext,
				"has_prev":    hasPrev,
			},
		},
	})
}
