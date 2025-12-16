package service

import (
	"context"
	"time"

	"achievement-backend/app/models"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

/* =====================================================
   1. MOCK ACHIEVEMENT REFERENCE REPOSITORY
   ===================================================== */
type MockAchievementReferenceRepository struct {
	CreateFn                   func(ctx context.Context, ref *models.AchievementReference) error
	FindByIDFn                 func(ctx context.Context, id uuid.UUID) (*models.AchievementReference, error)
	FindByStudentIDFn          func(ctx context.Context, studentID uuid.UUID, status string, page, limit int) ([]*models.AchievementReference, int, error)
	FindByMongoIDFn            func(ctx context.Context, mongoID string) (*models.AchievementReference, error)
	UpdateStatusFn             func(ctx context.Context, id uuid.UUID, status string, verifiedBy *uuid.UUID, rejectionNote *string) error
	DeleteFn                   func(ctx context.Context, id uuid.UUID) error
	SoftDeleteFn               func(ctx context.Context, id uuid.UUID) error
	FindByAdvisorIDFn          func(ctx context.Context, advisorID uuid.UUID, status string, page, limit int) ([]*models.AchievementReference, int, error)
	FindAllFn                  func(ctx context.Context, status string, page, limit int) ([]*models.AchievementReference, int, error)
	SubmitForVerificationFn    func(ctx context.Context, id uuid.UUID) error
	VerifyAchievementFn        func(ctx context.Context, id uuid.UUID, verifiedBy uuid.UUID) error
	RejectAchievementFn        func(ctx context.Context, id uuid.UUID, verifiedBy uuid.UUID, rejectionNote string) error
	CountByStatusFn            func(ctx context.Context, studentID uuid.UUID) (map[string]int, error)
	CountByStudentAndStatusFn  func(ctx context.Context, studentID uuid.UUID, status string) (int, error)
	GetStudentIDsByAdvisorFn   func(ctx context.Context, advisorID uuid.UUID) ([]uuid.UUID, error)
}

func (m *MockAchievementReferenceRepository) Create(ctx context.Context, ref *models.AchievementReference) error {
	if m.CreateFn != nil { return m.CreateFn(ctx, ref) }
	return nil
}
func (m *MockAchievementReferenceRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.AchievementReference, error) {
	if m.FindByIDFn != nil { return m.FindByIDFn(ctx, id) }
	return nil, nil // Default nil
}
func (m *MockAchievementReferenceRepository) FindByStudentID(ctx context.Context, studentID uuid.UUID, status string, page, limit int) ([]*models.AchievementReference, int, error) {
	if m.FindByStudentIDFn != nil { return m.FindByStudentIDFn(ctx, studentID, status, page, limit) }
	return []*models.AchievementReference{}, 0, nil
}
func (m *MockAchievementReferenceRepository) FindByMongoID(ctx context.Context, mongoID string) (*models.AchievementReference, error) {
	if m.FindByMongoIDFn != nil { return m.FindByMongoIDFn(ctx, mongoID) }
	return nil, nil
}
func (m *MockAchievementReferenceRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string, verifiedBy *uuid.UUID, rejectionNote *string) error {
	if m.UpdateStatusFn != nil { return m.UpdateStatusFn(ctx, id, status, verifiedBy, rejectionNote) }
	return nil
}
func (m *MockAchievementReferenceRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if m.DeleteFn != nil { return m.DeleteFn(ctx, id) }
	return nil
}
func (m *MockAchievementReferenceRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	if m.SoftDeleteFn != nil { return m.SoftDeleteFn(ctx, id) }
	return nil
}
func (m *MockAchievementReferenceRepository) FindByAdvisorID(ctx context.Context, advisorID uuid.UUID, status string, page, limit int) ([]*models.AchievementReference, int, error) {
	if m.FindByAdvisorIDFn != nil { return m.FindByAdvisorIDFn(ctx, advisorID, status, page, limit) }
	return []*models.AchievementReference{}, 0, nil
}
func (m *MockAchievementReferenceRepository) FindAll(ctx context.Context, status string, page, limit int) ([]*models.AchievementReference, int, error) {
	if m.FindAllFn != nil { return m.FindAllFn(ctx, status, page, limit) }
	return []*models.AchievementReference{}, 0, nil
}
func (m *MockAchievementReferenceRepository) SubmitForVerification(ctx context.Context, id uuid.UUID) error {
	if m.SubmitForVerificationFn != nil { return m.SubmitForVerificationFn(ctx, id) }
	return nil
}
func (m *MockAchievementReferenceRepository) VerifyAchievement(ctx context.Context, id uuid.UUID, verifiedBy uuid.UUID) error {
	if m.VerifyAchievementFn != nil { return m.VerifyAchievementFn(ctx, id, verifiedBy) }
	return nil
}
func (m *MockAchievementReferenceRepository) RejectAchievement(ctx context.Context, id uuid.UUID, verifiedBy uuid.UUID, rejectionNote string) error {
	if m.RejectAchievementFn != nil { return m.RejectAchievementFn(ctx, id, verifiedBy, rejectionNote) }
	return nil
}
func (m *MockAchievementReferenceRepository) CountByStatus(ctx context.Context, studentID uuid.UUID) (map[string]int, error) {
	if m.CountByStatusFn != nil { return m.CountByStatusFn(ctx, studentID) }
	return map[string]int{}, nil
}
func (m *MockAchievementReferenceRepository) CountByStudentAndStatus(ctx context.Context, studentID uuid.UUID, status string) (int, error) {
	if m.CountByStudentAndStatusFn != nil { return m.CountByStudentAndStatusFn(ctx, studentID, status) }
	return 0, nil
}
func (m *MockAchievementReferenceRepository) GetStudentIDsByAdvisor(ctx context.Context, advisorID uuid.UUID) ([]uuid.UUID, error) {
	if m.GetStudentIDsByAdvisorFn != nil { return m.GetStudentIDsByAdvisorFn(ctx, advisorID) }
	return []uuid.UUID{}, nil
}

/* =====================================================
   2. MOCK ACHIEVEMENT REPOSITORY
   ===================================================== */
type MockAchievementRepository struct {
	CreateFn                 func(ctx context.Context, achievement *models.Achievement) (primitive.ObjectID, error)
	FindByIDFn               func(ctx context.Context, id primitive.ObjectID) (*models.Achievement, error)
	FindByStudentIDFn        func(ctx context.Context, studentID uuid.UUID, page, limit int) ([]*models.Achievement, int, error)
	UpdateFn                 func(ctx context.Context, id primitive.ObjectID, achievement *models.Achievement) error
	DeleteFn                 func(ctx context.Context, id primitive.ObjectID) error
	FindWithFilterFn         func(ctx context.Context, filter bson.M, page, limit int) ([]*models.Achievement, int, error)
	FindByStudentIDsFn       func(ctx context.Context, studentIDs []uuid.UUID, page, limit int) ([]*models.Achievement, int, error)
	FindByStatusFn           func(ctx context.Context, studentID uuid.UUID, status string, page, limit int) ([]*models.Achievement, int, error)
	GetStatisticsByStudentFn func(ctx context.Context, studentID uuid.UUID) (*models.AchievementStatistics, error)
	GetStatisticsByAdvisorFn func(ctx context.Context, advisorID uuid.UUID) (*models.AchievementStatistics, error)
	CountByTypeFn            func(ctx context.Context, studentID uuid.UUID) (map[string]int, error)
	CountByPeriodFn          func(ctx context.Context, studentID uuid.UUID, startDate, endDate time.Time) (map[string]int, error)
}

func (m *MockAchievementRepository) Create(ctx context.Context, achievement *models.Achievement) (primitive.ObjectID, error) {
	if m.CreateFn != nil { return m.CreateFn(ctx, achievement) }
	return primitive.NilObjectID, nil
}
func (m *MockAchievementRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Achievement, error) {
	if m.FindByIDFn != nil { return m.FindByIDFn(ctx, id) }
	return nil, nil
}
func (m *MockAchievementRepository) FindByStudentID(ctx context.Context, studentID uuid.UUID, page, limit int) ([]*models.Achievement, int, error) {
	if m.FindByStudentIDFn != nil { return m.FindByStudentIDFn(ctx, studentID, page, limit) }
	return []*models.Achievement{}, 0, nil
}
func (m *MockAchievementRepository) Update(ctx context.Context, id primitive.ObjectID, achievement *models.Achievement) error {
	if m.UpdateFn != nil { return m.UpdateFn(ctx, id, achievement) }
	return nil
}
func (m *MockAchievementRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	if m.DeleteFn != nil { return m.DeleteFn(ctx, id) }
	return nil
}
func (m *MockAchievementRepository) FindWithFilter(ctx context.Context, filter bson.M, page, limit int) ([]*models.Achievement, int, error) {
	if m.FindWithFilterFn != nil { return m.FindWithFilterFn(ctx, filter, page, limit) }
	return []*models.Achievement{}, 0, nil
}
func (m *MockAchievementRepository) FindByStudentIDs(ctx context.Context, studentIDs []uuid.UUID, page, limit int) ([]*models.Achievement, int, error) {
	if m.FindByStudentIDsFn != nil { return m.FindByStudentIDsFn(ctx, studentIDs, page, limit) }
	return []*models.Achievement{}, 0, nil
}
func (m *MockAchievementRepository) FindByStatus(ctx context.Context, studentID uuid.UUID, status string, page, limit int) ([]*models.Achievement, int, error) {
	if m.FindByStatusFn != nil { return m.FindByStatusFn(ctx, studentID, status, page, limit) }
	return []*models.Achievement{}, 0, nil
}
func (m *MockAchievementRepository) GetStatisticsByStudent(ctx context.Context, studentID uuid.UUID) (*models.AchievementStatistics, error) {
	if m.GetStatisticsByStudentFn != nil { return m.GetStatisticsByStudentFn(ctx, studentID) }
	return nil, nil
}
func (m *MockAchievementRepository) GetStatisticsByAdvisor(ctx context.Context, advisorID uuid.UUID) (*models.AchievementStatistics, error) {
	if m.GetStatisticsByAdvisorFn != nil { return m.GetStatisticsByAdvisorFn(ctx, advisorID) }
	return nil, nil
}
func (m *MockAchievementRepository) CountByType(ctx context.Context, studentID uuid.UUID) (map[string]int, error) {
	if m.CountByTypeFn != nil { return m.CountByTypeFn(ctx, studentID) }
	return map[string]int{}, nil
}
func (m *MockAchievementRepository) CountByPeriod(ctx context.Context, studentID uuid.UUID, startDate, endDate time.Time) (map[string]int, error) {
	if m.CountByPeriodFn != nil { return m.CountByPeriodFn(ctx, studentID, startDate, endDate) }
	return map[string]int{}, nil
}

/* =====================================================
   3. MOCK LECTURER REPOSITORY
   ===================================================== */
type MockLecturerRepository struct {
	GetByIDFn            func(id uuid.UUID) (*models.Lecturer, error)
	GetByUserIDFn        func(userID uuid.UUID) (*models.Lecturer, error)
	GetByLecturerIDFn    func(lecturerID string) (*models.Lecturer, error)
	CreateFn             func(lecturer models.Lecturer) (uuid.UUID, error)
	UpdateFn             func(id uuid.UUID, req *models.UpdateLecturerRequest) error
	GetAllFn             func(page, limit int) ([]models.Lecturer, int, error)
	GetTotalCountFn      func() (int, error)
	GetWithUserDetailsFn func(page, limit int) ([]models.LecturerResponse, int, error)
	GetAdviseesCountFn   func(lecturerID uuid.UUID) (int, error)
	GetAdviseesFn        func(lecturerID uuid.UUID, page, limit int) ([]models.Student, int, error)
	SearchByNameFn       func(name string, page, limit int) ([]models.LecturerResponse, int, error)
	GetByDepartmentFn    func(department string, page, limit int) ([]models.Lecturer, int, error)
}

func (m *MockLecturerRepository) GetByID(id uuid.UUID) (*models.Lecturer, error) {
	if m.GetByIDFn != nil { return m.GetByIDFn(id) }
	return nil, nil
}
func (m *MockLecturerRepository) GetByUserID(userID uuid.UUID) (*models.Lecturer, error) {
	if m.GetByUserIDFn != nil { return m.GetByUserIDFn(userID) }
	return nil, nil
}
func (m *MockLecturerRepository) GetByLecturerID(lecturerID string) (*models.Lecturer, error) {
	if m.GetByLecturerIDFn != nil { return m.GetByLecturerIDFn(lecturerID) }
	return nil, nil
}
func (m *MockLecturerRepository) Create(lecturer models.Lecturer) (uuid.UUID, error) {
	if m.CreateFn != nil { return m.CreateFn(lecturer) }
	return uuid.Nil, nil
}
func (m *MockLecturerRepository) Update(id uuid.UUID, req *models.UpdateLecturerRequest) error {
	if m.UpdateFn != nil { return m.UpdateFn(id, req) }
	return nil
}
func (m *MockLecturerRepository) GetAll(page, limit int) ([]models.Lecturer, int, error) {
	if m.GetAllFn != nil { return m.GetAllFn(page, limit) }
	return []models.Lecturer{}, 0, nil
}
func (m *MockLecturerRepository) GetTotalCount() (int, error) {
	if m.GetTotalCountFn != nil { return m.GetTotalCountFn() }
	return 0, nil
}
func (m *MockLecturerRepository) GetWithUserDetails(page, limit int) ([]models.LecturerResponse, int, error) {
	if m.GetWithUserDetailsFn != nil { return m.GetWithUserDetailsFn(page, limit) }
	return []models.LecturerResponse{}, 0, nil
}
func (m *MockLecturerRepository) GetAdviseesCount(lecturerID uuid.UUID) (int, error) {
	if m.GetAdviseesCountFn != nil { return m.GetAdviseesCountFn(lecturerID) }
	return 0, nil
}
func (m *MockLecturerRepository) GetAdvisees(lecturerID uuid.UUID, page, limit int) ([]models.Student, int, error) {
	if m.GetAdviseesFn != nil { return m.GetAdviseesFn(lecturerID, page, limit) }
	return []models.Student{}, 0, nil
}
func (m *MockLecturerRepository) SearchByName(name string, page, limit int) ([]models.LecturerResponse, int, error) {
	if m.SearchByNameFn != nil { return m.SearchByNameFn(name, page, limit) }
	return []models.LecturerResponse{}, 0, nil
}
func (m *MockLecturerRepository) GetByDepartment(department string, page, limit int) ([]models.Lecturer, int, error) {
	if m.GetByDepartmentFn != nil { return m.GetByDepartmentFn(department, page, limit) }
	return []models.Lecturer{}, 0, nil
}

/* =====================================================
   4. MOCK ROLE REPOSITORY
   ===================================================== */
type MockRoleRepository struct {
	GetByIDFn                  func(id uuid.UUID) (*models.Role, error)
	GetByNameFn                func(name string) (*models.Role, error)
	GetAllFn                   func(page, limit int) ([]models.Role, int, error)
	GetTotalCountFn            func() (int, error)
	GetPermissionsByRoleIDFn   func(roleID uuid.UUID) ([]models.Permission, error)
	GetPermissionNamesByRoleIDFn func(roleID uuid.UUID) ([]string, error)
	AssignPermissionFn         func(roleID, permissionID uuid.UUID) error
	RemovePermissionFn         func(roleID, permissionID uuid.UUID) error
}

func (m *MockRoleRepository) GetByID(id uuid.UUID) (*models.Role, error) {
	if m.GetByIDFn != nil { return m.GetByIDFn(id) }
	return nil, nil
}
func (m *MockRoleRepository) GetByName(name string) (*models.Role, error) {
	if m.GetByNameFn != nil { return m.GetByNameFn(name) }
	return nil, nil
}
func (m *MockRoleRepository) GetAll(page, limit int) ([]models.Role, int, error) {
	if m.GetAllFn != nil { return m.GetAllFn(page, limit) }
	return []models.Role{}, 0, nil
}
func (m *MockRoleRepository) GetTotalCount() (int, error) {
	if m.GetTotalCountFn != nil { return m.GetTotalCountFn() }
	return 0, nil
}
func (m *MockRoleRepository) GetPermissionsByRoleID(roleID uuid.UUID) ([]models.Permission, error) {
	if m.GetPermissionsByRoleIDFn != nil { return m.GetPermissionsByRoleIDFn(roleID) }
	return []models.Permission{}, nil
}
func (m *MockRoleRepository) GetPermissionNamesByRoleID(roleID uuid.UUID) ([]string, error) {
	if m.GetPermissionNamesByRoleIDFn != nil { return m.GetPermissionNamesByRoleIDFn(roleID) }
	return []string{}, nil
}
func (m *MockRoleRepository) AssignPermission(roleID, permissionID uuid.UUID) error {
	if m.AssignPermissionFn != nil { return m.AssignPermissionFn(roleID, permissionID) }
	return nil
}
func (m *MockRoleRepository) RemovePermission(roleID, permissionID uuid.UUID) error {
	if m.RemovePermissionFn != nil { return m.RemovePermissionFn(roleID, permissionID) }
	return nil
}

/* =====================================================
   5. MOCK STUDENT REPOSITORY
   ===================================================== */
type MockStudentRepository struct {
	GetByIDFn                     func(id uuid.UUID) (*models.Student, error)
	GetByUserIDFn                 func(userID uuid.UUID) (*models.Student, error)
	GetByStudentIDFn              func(studentID string) (*models.Student, error)
	CreateFn                      func(student models.Student) (uuid.UUID, error)
	UpdateFn                      func(id uuid.UUID, req *models.UpdateStudentRequest) error
	UpdateAdvisorFn               func(studentID, advisorID uuid.UUID) error
	RemoveAdvisorFn               func(studentID uuid.UUID) error
	GetAllFn                      func(page, limit int) ([]models.Student, int, error)
	GetTotalCountFn               func() (int, error)
	GetWithUserDetailsFn          func(page, limit int) ([]models.StudentResponse, int, error)
	GetWithAdvisorDetailsFn       func(page, limit int) ([]models.StudentResponse, int, error)
	GetAllByAdvisorIDFn           func(advisorID uuid.UUID, page, limit int) ([]models.Student, int, error)
	GetAdvisorlessFn              func(page, limit int) ([]models.Student, int, error)
	SearchByNameFn                func(name string, page, limit int) ([]models.StudentResponse, int, error)
	GetByProgramStudyFn           func(programStudy string, page, limit int) ([]models.Student, int, error)
	GetByAcademicYearFn           func(academicYear string, page, limit int) ([]models.Student, int, error)
	GetStudentsCountByAdvisorFn   func(advisorID uuid.UUID) (int, error)
	GetStudentsCountByProgramStudyFn func() (map[string]int, error)
}

func (m *MockStudentRepository) GetByID(id uuid.UUID) (*models.Student, error) {
	if m.GetByIDFn != nil { return m.GetByIDFn(id) }
	return nil, nil
}
func (m *MockStudentRepository) GetByUserID(userID uuid.UUID) (*models.Student, error) {
	if m.GetByUserIDFn != nil { return m.GetByUserIDFn(userID) }
	return nil, nil
}
func (m *MockStudentRepository) GetByStudentID(studentID string) (*models.Student, error) {
	if m.GetByStudentIDFn != nil { return m.GetByStudentIDFn(studentID) }
	return nil, nil
}
func (m *MockStudentRepository) Create(student models.Student) (uuid.UUID, error) {
	if m.CreateFn != nil { return m.CreateFn(student) }
	return uuid.Nil, nil
}
func (m *MockStudentRepository) Update(id uuid.UUID, req *models.UpdateStudentRequest) error {
	if m.UpdateFn != nil { return m.UpdateFn(id, req) }
	return nil
}
func (m *MockStudentRepository) UpdateAdvisor(studentID, advisorID uuid.UUID) error {
	if m.UpdateAdvisorFn != nil { return m.UpdateAdvisorFn(studentID, advisorID) }
	return nil
}
func (m *MockStudentRepository) RemoveAdvisor(studentID uuid.UUID) error {
	if m.RemoveAdvisorFn != nil { return m.RemoveAdvisorFn(studentID) }
	return nil
}
func (m *MockStudentRepository) GetAll(page, limit int) ([]models.Student, int, error) {
	if m.GetAllFn != nil { return m.GetAllFn(page, limit) }
	return []models.Student{}, 0, nil
}
func (m *MockStudentRepository) GetTotalCount() (int, error) {
	if m.GetTotalCountFn != nil { return m.GetTotalCountFn() }
	return 0, nil
}
func (m *MockStudentRepository) GetWithUserDetails(page, limit int) ([]models.StudentResponse, int, error) {
	if m.GetWithUserDetailsFn != nil { return m.GetWithUserDetailsFn(page, limit) }
	return []models.StudentResponse{}, 0, nil
}
func (m *MockStudentRepository) GetWithAdvisorDetails(page, limit int) ([]models.StudentResponse, int, error) {
	if m.GetWithAdvisorDetailsFn != nil { return m.GetWithAdvisorDetailsFn(page, limit) }
	return []models.StudentResponse{}, 0, nil
}
func (m *MockStudentRepository) GetAllByAdvisorID(advisorID uuid.UUID, page, limit int) ([]models.Student, int, error) {
	if m.GetAllByAdvisorIDFn != nil { return m.GetAllByAdvisorIDFn(advisorID, page, limit) }
	return []models.Student{}, 0, nil
}
func (m *MockStudentRepository) GetAdvisorless(page, limit int) ([]models.Student, int, error) {
	if m.GetAdvisorlessFn != nil { return m.GetAdvisorlessFn(page, limit) }
	return []models.Student{}, 0, nil
}
func (m *MockStudentRepository) SearchByName(name string, page, limit int) ([]models.StudentResponse, int, error) {
	if m.SearchByNameFn != nil { return m.SearchByNameFn(name, page, limit) }
	return []models.StudentResponse{}, 0, nil
}
func (m *MockStudentRepository) GetByProgramStudy(programStudy string, page, limit int) ([]models.Student, int, error) {
	if m.GetByProgramStudyFn != nil { return m.GetByProgramStudyFn(programStudy, page, limit) }
	return []models.Student{}, 0, nil
}
func (m *MockStudentRepository) GetByAcademicYear(academicYear string, page, limit int) ([]models.Student, int, error) {
	if m.GetByAcademicYearFn != nil { return m.GetByAcademicYearFn(academicYear, page, limit) }
	return []models.Student{}, 0, nil
}
func (m *MockStudentRepository) GetStudentsCountByAdvisor(advisorID uuid.UUID) (int, error) {
	if m.GetStudentsCountByAdvisorFn != nil { return m.GetStudentsCountByAdvisorFn(advisorID) }
	return 0, nil
}
func (m *MockStudentRepository) GetStudentsCountByProgramStudy() (map[string]int, error) {
	if m.GetStudentsCountByProgramStudyFn != nil { return m.GetStudentsCountByProgramStudyFn() }
	return map[string]int{}, nil
}

/* =====================================================
   6. MOCK USER REPOSITORY
   ===================================================== */
type MockUserRepository struct {
	GetByIDFn              func(id uuid.UUID) (*models.User, error)
	GetByEmailFn           func(email string) (*models.User, error)
	GetByUsernameFn        func(username string) (*models.User, error)
	GetByUsernameOrEmailFn func(identifier string) (*models.User, error)
	CreateFn               func(user *models.User) (uuid.UUID, error)
	UpdateFn               func(id uuid.UUID, req *models.UpdateUserRequest) error
	UpdatePasswordFn       func(id uuid.UUID, hashedPassword string) error
	SoftDeleteFn           func(id uuid.UUID) error
	HardDeleteFn           func(id uuid.UUID) error
	GetAllFn               func(page, limit int) ([]models.User, int, error)
	GetInactiveUsersFn     func(page, limit int) ([]models.User, int, error)
	GetAllWithInactiveFn   func(page, limit int) ([]models.User, int, error)
	SearchByNameFn         func(name string, page, limit int) ([]models.User, int, error)
	GetByRoleFn            func(roleID uuid.UUID, page, limit int) ([]models.User, int, error)
	GetUsersCountByRoleFn  func() (map[uuid.UUID]int, error)
	GetTotalActiveCountFn  func() (int, error)
	GetTotalInactiveCountFn func() (int, error)
}

func (m *MockUserRepository) GetByID(id uuid.UUID) (*models.User, error) {
	if m.GetByIDFn != nil { return m.GetByIDFn(id) }
	return nil, nil
}
func (m *MockUserRepository) GetByEmail(email string) (*models.User, error) {
	if m.GetByEmailFn != nil { return m.GetByEmailFn(email) }
	return nil, nil
}
func (m *MockUserRepository) GetByUsername(username string) (*models.User, error) {
	if m.GetByUsernameFn != nil { return m.GetByUsernameFn(username) }
	return nil, nil
}
func (m *MockUserRepository) GetByUsernameOrEmail(identifier string) (*models.User, error) {
	if m.GetByUsernameOrEmailFn != nil { return m.GetByUsernameOrEmailFn(identifier) }
	return nil, nil
}
func (m *MockUserRepository) Create(user *models.User) (uuid.UUID, error) {
	if m.CreateFn != nil { return m.CreateFn(user) }
	return uuid.Nil, nil
}
func (m *MockUserRepository) Update(id uuid.UUID, req *models.UpdateUserRequest) error {
	if m.UpdateFn != nil { return m.UpdateFn(id, req) }
	return nil
}
func (m *MockUserRepository) UpdatePassword(id uuid.UUID, hashedPassword string) error {
	if m.UpdatePasswordFn != nil { return m.UpdatePasswordFn(id, hashedPassword) }
	return nil
}
func (m *MockUserRepository) SoftDelete(id uuid.UUID) error {
	if m.SoftDeleteFn != nil { return m.SoftDeleteFn(id) }
	return nil
}
func (m *MockUserRepository) HardDelete(id uuid.UUID) error {
	if m.HardDeleteFn != nil { return m.HardDeleteFn(id) }
	return nil
}
func (m *MockUserRepository) GetAll(page, limit int) ([]models.User, int, error) {
	if m.GetAllFn != nil { return m.GetAllFn(page, limit) }
	return []models.User{}, 0, nil
}
func (m *MockUserRepository) GetInactiveUsers(page, limit int) ([]models.User, int, error) {
	if m.GetInactiveUsersFn != nil { return m.GetInactiveUsersFn(page, limit) }
	return []models.User{}, 0, nil
}
func (m *MockUserRepository) GetAllWithInactive(page, limit int) ([]models.User, int, error) {
	if m.GetAllWithInactiveFn != nil { return m.GetAllWithInactiveFn(page, limit) }
	return []models.User{}, 0, nil
}
func (m *MockUserRepository) SearchByName(name string, page, limit int) ([]models.User, int, error) {
	if m.SearchByNameFn != nil { return m.SearchByNameFn(name, page, limit) }
	return []models.User{}, 0, nil
}
func (m *MockUserRepository) GetByRole(roleID uuid.UUID, page, limit int) ([]models.User, int, error) {
	if m.GetByRoleFn != nil { return m.GetByRoleFn(roleID, page, limit) }
	return []models.User{}, 0, nil
}
func (m *MockUserRepository) GetUsersCountByRole() (map[uuid.UUID]int, error) {
	if m.GetUsersCountByRoleFn != nil { return m.GetUsersCountByRoleFn() }
	return map[uuid.UUID]int{}, nil
}
func (m *MockUserRepository) GetTotalActiveCount() (int, error) {
	if m.GetTotalActiveCountFn != nil { return m.GetTotalActiveCountFn() }
	return 0, nil
}
func (m *MockUserRepository) GetTotalInactiveCount() (int, error) {
	if m.GetTotalInactiveCountFn != nil { return m.GetTotalInactiveCountFn() }
	return 0, nil
}