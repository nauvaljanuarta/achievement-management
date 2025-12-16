package service

import (
	"context"
	"errors"
	"time"

	"achievement-backend/app/models"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

/* =====================================================
   1. MOCK ACHIEVEMENT REPOSITORY (Updated)
   ===================================================== */
type MockAchievementRepository struct {
	CreateFn                 func(ctx context.Context, a *models.Achievement) (primitive.ObjectID, error)
	FindByIDFn               func(ctx context.Context, id primitive.ObjectID) (*models.Achievement, error)
	UpdateFn                 func(ctx context.Context, id primitive.ObjectID, a *models.Achievement) error
	DeleteFn                 func(ctx context.Context, id primitive.ObjectID) error
	GetStatisticsByStudentFn func(ctx context.Context, studentID uuid.UUID) (*models.AchievementStatistics, error)
	GetStatisticsByAdvisorFn func(ctx context.Context, advisorID uuid.UUID) (*models.AchievementStatistics, error)
	
    // Method tambahan yang sebelumnya error
	CountByPeriodFn          func(ctx context.Context, id uuid.UUID, start, end time.Time) (map[string]int, error)
	CountByTypeFn            func(ctx context.Context, studentID uuid.UUID) (map[string]int, error) // <--- INI DITAMBAHKAN
}

// ... method Create, FindByID, Update, Delete, Stats ... (sama seperti sebelumnya)

func (m *MockAchievementRepository) Create(ctx context.Context, a *models.Achievement) (primitive.ObjectID, error) {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, a)
	}
	return primitive.NilObjectID, errors.New("not implemented")
}
func (m *MockAchievementRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Achievement, error) {
	if m.FindByIDFn != nil {
		return m.FindByIDFn(ctx, id)
	}
	return nil, errors.New("not implemented")
}
func (m *MockAchievementRepository) Update(ctx context.Context, id primitive.ObjectID, a *models.Achievement) error {
	if m.UpdateFn != nil {
		return m.UpdateFn(ctx, id, a)
	}
	return errors.New("not implemented")
}
func (m *MockAchievementRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	if m.DeleteFn != nil {
		return m.DeleteFn(ctx, id)
	}
	return errors.New("not implemented")
}
func (m *MockAchievementRepository) GetStatisticsByStudent(ctx context.Context, studentID uuid.UUID) (*models.AchievementStatistics, error) {
	if m.GetStatisticsByStudentFn != nil {
		return m.GetStatisticsByStudentFn(ctx, studentID)
	}
	return nil, errors.New("not implemented")
}
func (m *MockAchievementRepository) GetStatisticsByAdvisor(ctx context.Context, advisorID uuid.UUID) (*models.AchievementStatistics, error) {
	if m.GetStatisticsByAdvisorFn != nil {
		return m.GetStatisticsByAdvisorFn(ctx, advisorID)
	}
	return nil, errors.New("not implemented")
}
func (m *MockAchievementRepository) CountByPeriod(ctx context.Context, id uuid.UUID, start, end time.Time) (map[string]int, error) {
	if m.CountByPeriodFn != nil {
		return m.CountByPeriodFn(ctx, id, start, end)
	}
	return map[string]int{}, nil
}

// IMPLEMENTASI BARU: CountByType
func (m *MockAchievementRepository) CountByType(ctx context.Context, studentID uuid.UUID) (map[string]int, error) {
	if m.CountByTypeFn != nil {
		return m.CountByTypeFn(ctx, studentID)
	}
	return map[string]int{}, nil // Default return map kosong
}


/* =====================================================
   2. MOCK STUDENT REPOSITORY (Updated)
   ===================================================== */
type MockStudentRepository struct {
	GetByUserFn func(userID uuid.UUID) (*models.Student, error)
	GetByIDFn   func(id uuid.UUID) (*models.Student, error)
	CreateFn    func(ctx context.Context, s *models.Student) error // <--- INI DITAMBAHKAN
}

func (m *MockStudentRepository) GetByUserID(userID uuid.UUID) (*models.Student, error) {
	if m.GetByUserFn != nil {
		return m.GetByUserFn(userID)
	}
	return nil, errors.New("not implemented")
}

func (m *MockStudentRepository) GetByID(id uuid.UUID) (*models.Student, error) {
	if m.GetByIDFn != nil {
		return m.GetByIDFn(id)
	}
	return nil, errors.New("not implemented")
}

// IMPLEMENTASI BARU: Create
func (m *MockStudentRepository) Create(ctx context.Context, s *models.Student) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, s)
	}
	return nil // Default sukses
}


/* =====================================================
   3. MOCK ROLE REPOSITORY (Updated)
   ===================================================== */
type MockRoleRepository struct {
	GetByIDFn          func(id uuid.UUID) (*models.Role, error)
	AssignPermissionFn func(ctx context.Context, roleID uuid.UUID, permissionID uuid.UUID) error // <--- INI DITAMBAHKAN
}

func (m *MockRoleRepository) GetByID(id uuid.UUID) (*models.Role, error) {
	if m.GetByIDFn != nil {
		return m.GetByIDFn(id)
	}
	return nil, errors.New("not implemented")
}

// IMPLEMENTASI BARU: AssignPermission
func (m *MockRoleRepository) AssignPermission(ctx context.Context, roleID uuid.UUID, permissionID uuid.UUID) error {
	if m.AssignPermissionFn != nil {
		return m.AssignPermissionFn(ctx, roleID, permissionID)
	}
	return nil // Default sukses
}


/* =====================================================
   MOCK LAINNYA (Reference, User, Lecturer) - Biarkan Saja
   ===================================================== */
// ... (Bagian Reference, User, Lecturer yg lama tidak perlu diubah kalau tidak error)
// Pastikan MockAchievementReferenceRepository tetap punya CountByStatus yg tadi kita tambahkan.

type MockAchievementReferenceRepository struct {
	CreateFn                func(ctx context.Context, ref *models.AchievementReference) error
	FindByIDFn              func(ctx context.Context, id uuid.UUID) (*models.AchievementReference, error)
	FindByStudentIDFn       func(ctx context.Context, studentID uuid.UUID, status string, page, limit int) ([]*models.AchievementReference, int, error)
	FindByAdvisorIDFn       func(ctx context.Context, advisorID uuid.UUID, status string, page, limit int) ([]*models.AchievementReference, int, error)
	FindAllFn               func(ctx context.Context, status string, page, limit int) ([]*models.AchievementReference, int, error)
	UpdateStatusFn          func(ctx context.Context, id uuid.UUID, status string, note *string, verifiedBy *uuid.UUID) error
	SoftDeleteFn            func(ctx context.Context, id uuid.UUID) error
	SubmitForVerificationFn func(ctx context.Context, id uuid.UUID) error
	VerifyAchievementFn     func(ctx context.Context, id uuid.UUID, verifiedBy uuid.UUID) error
	RejectAchievementFn     func(ctx context.Context, id uuid.UUID, rejectedBy uuid.UUID, note string) error
	CountByStatusFn         func(ctx context.Context, status string) (int64, error)
}

// ... method-method Reference Repository (Copy paste yg lama) ...
// Saya singkat disini agar tidak kepanjangan, 
// intinya method Create, FindByID, dll harus tetap ada.

func (m *MockAchievementReferenceRepository) Create(ctx context.Context, ref *models.AchievementReference) error {
	if m.CreateFn != nil { return m.CreateFn(ctx, ref) }
	return errors.New("not implemented")
}
func (m *MockAchievementReferenceRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.AchievementReference, error) {
	if m.FindByIDFn != nil { return m.FindByIDFn(ctx, id) }
	return nil, errors.New("not implemented")
}
func (m *MockAchievementReferenceRepository) FindByStudentID(ctx context.Context, studentID uuid.UUID, status string, page, limit int) ([]*models.AchievementReference, int, error) {
	if m.FindByStudentIDFn != nil { return m.FindByStudentIDFn(ctx, studentID, status, page, limit) }
	return []*models.AchievementReference{}, 0, nil
}
func (m *MockAchievementReferenceRepository) FindByAdvisorID(ctx context.Context, advisorID uuid.UUID, status string, page, limit int) ([]*models.AchievementReference, int, error) {
	if m.FindByAdvisorIDFn != nil { return m.FindByAdvisorIDFn(ctx, advisorID, status, page, limit) }
	return []*models.AchievementReference{}, 0, nil
}
func (m *MockAchievementReferenceRepository) FindAll(ctx context.Context, status string, page, limit int) ([]*models.AchievementReference, int, error) {
	if m.FindAllFn != nil { return m.FindAllFn(ctx, status, page, limit) }
	return []*models.AchievementReference{}, 0, nil
}
func (m *MockAchievementReferenceRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string, note *string, verifiedBy *uuid.UUID) error {
	if m.UpdateStatusFn != nil { return m.UpdateStatusFn(ctx, id, status, note, verifiedBy) }
	return errors.New("not implemented")
}
func (m *MockAchievementReferenceRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	if m.SoftDeleteFn != nil { return m.SoftDeleteFn(ctx, id) }
	return errors.New("not implemented")
}
func (m *MockAchievementReferenceRepository) SubmitForVerification(ctx context.Context, id uuid.UUID) error {
	if m.SubmitForVerificationFn != nil { return m.SubmitForVerificationFn(ctx, id) }
	return errors.New("not implemented")
}
func (m *MockAchievementReferenceRepository) VerifyAchievement(ctx context.Context, id uuid.UUID, verifiedBy uuid.UUID) error {
	if m.VerifyAchievementFn != nil { return m.VerifyAchievementFn(ctx, id, verifiedBy) }
	return errors.New("not implemented")
}
func (m *MockAchievementReferenceRepository) RejectAchievement(ctx context.Context, id uuid.UUID, rejectedBy uuid.UUID, note string) error {
	if m.RejectAchievementFn != nil { return m.RejectAchievementFn(ctx, id, rejectedBy, note) }
	return errors.New("not implemented")
}
func (m *MockAchievementReferenceRepository) CountByStatus(ctx context.Context, status string) (int64, error) {
	if m.CountByStatusFn != nil { return m.CountByStatusFn(ctx, status) }
	return 0, nil
}

type MockLecturerRepository struct {
	GetByUserFn func(userID uuid.UUID) (*models.Lecturer, error)
}
func (m *MockLecturerRepository) GetByUserID(userID uuid.UUID) (*models.Lecturer, error) {
	if m.GetByUserFn != nil { return m.GetByUserFn(userID) }
	return nil, errors.New("not implemented")
}

type MockUserRepository struct {
	GetByIDFn func(id uuid.UUID) (*models.User, error)
}
func (m *MockUserRepository) GetByID(id uuid.UUID) (*models.User, error) {
	if m.GetByIDFn != nil { return m.GetByIDFn(id) }
	return nil, errors.New("not implemented")
}