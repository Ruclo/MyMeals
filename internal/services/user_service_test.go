package services_test

import (
	"github.com/Ruclo/MyMeals/internal/apperrors"
	"github.com/Ruclo/MyMeals/internal/models"
	"github.com/Ruclo/MyMeals/internal/repositories"
	"github.com/Ruclo/MyMeals/internal/services"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/bcrypt"
	"testing"
)

type UserServiceTestSuite struct {
	suite.Suite
	userService services.UserService
	mockRepo    *MockUserRepository
}

func (s *UserServiceTestSuite) SetupTest() {
	// Create a fresh mock repo for each test
	s.mockRepo = new(MockUserRepository)
	s.userService = services.NewUserService(s.mockRepo)
}

// TearDownTest runs after each test
func (s *UserServiceTestSuite) TearDownTest() {
	// Verify all mock expectations were met
	s.mockRepo.AssertExpectations(s.T())
}

// TestLogin groups all login-related tests
func (s *UserServiceTestSuite) TestLogin() {
	// Define test cases
	testCases := []struct {
		name           string
		username       string
		password       string
		setupMock      func()
		expectedError  bool
		errorPredicate func(error) bool
		checkUser      func(*models.User)
	}{
		{
			name:     "Success",
			username: "testuser",
			password: "correctpassword",
			setupMock: func() {
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)
				mockUser := &models.User{
					Username: "testuser",
					Password: string(hashedPassword),
					Role:     models.AdminRole,
				}
				s.mockRepo.On("GetByUsername", "testuser").Return(mockUser, nil)
			},
			expectedError: false,
			checkUser: func(user *models.User) {
				s.Equal("testuser", user.Username)
				s.Equal(models.AdminRole, user.Role)
			},
		},
		{
			name:     "InvalidCredentials",
			username: "testuser",
			password: "wrongpassword",
			setupMock: func() {
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)
				mockUser := &models.User{
					Username: "testuser",
					Password: string(hashedPassword),
					Role:     models.AdminRole,
				}
				s.mockRepo.On("GetByUsername", "testuser").Return(mockUser, nil)
			},
			expectedError:  true,
			errorPredicate: apperrors.IsUnauthorizedErr,
		},
		{
			name:     "UserNotFound",
			username: "nonexistentuser",
			password: "anypassword",
			setupMock: func() {
				notFoundErr := apperrors.NewNotFoundErr("User not found", nil)
				s.mockRepo.On("GetByUsername", "nonexistentuser").Return(nil, notFoundErr)
			},
			expectedError:  true,
			errorPredicate: apperrors.IsUnauthorizedErr,
		},
	}

	// Run each test case as a subtest
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Create a fresh mock for each subtest
			s.SetupTest()

			// Setup mock expectations
			tc.setupMock()

			// Act
			user, err := s.userService.Login(tc.username, tc.password)

			// Assert
			if tc.expectedError {
				s.Error(err)
				if tc.errorPredicate != nil {
					s.True(tc.errorPredicate(err))
				}
				s.Nil(user)
			} else {
				s.NoError(err)
				s.NotNil(user)
				if tc.checkUser != nil {
					tc.checkUser(user)
				}
			}

			// Verify expectations (can also rely on TearDownTest)
			s.mockRepo.AssertExpectations(s.T())
		})
	}
}

// TestCreate tests the Create method (to be implemented)
func (s *UserServiceTestSuite) TestCreate() {
	// Define test cases for Create method
	testCases := []struct {
		name           string
		user           *models.User
		setupMock      func()
		expectedError  bool
		errorPredicate func(error) bool
	}{
		{
			name: "Success",
			user: &models.User{
				Username: "newuser",
				Password: "plainpassword",
			},
			setupMock: func() {
				s.mockRepo.On("Exists", "newuser").Return(false, nil)
				s.mockRepo.On("Create", mock.AnythingOfType("*models.User")).Return(nil)
			},
			expectedError: false,
		},
		{
			name: "UserAlreadyExists",
			user: &models.User{
				Username: "existinguser",
				Password: "plainpassword",
			},
			setupMock: func() {
				s.mockRepo.On("Exists", "existinguser").Return(true, nil)
			},
			expectedError:  true,
			errorPredicate: apperrors.IsAlreadyExistsErr,
		},
		{
			name: "RepositoryError",
			user: &models.User{
				Username: "erroruser",
				Password: "plainpassword",
			},
			setupMock: func() {
				s.mockRepo.On("Exists", "erroruser").Return(false, nil)
				repoErr := apperrors.NewInternalServerErr("Database error", nil)
				s.mockRepo.On("Create", mock.AnythingOfType("*models.User")).Return(repoErr)
			},
			expectedError: true,
		},
	}

	// Run each test case as a subtest
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Setup fresh mocks
			s.SetupTest()

			// Setup mock expectations
			tc.setupMock()

			// Make a copy to avoid modifications between tests
			userCopy := &models.User{
				Username: tc.user.Username,
				Password: tc.user.Password,
			}

			// Act
			err := s.userService.Create(userCopy)

			// Assert
			if tc.expectedError {
				s.Error(err)
				if tc.errorPredicate != nil {
					s.True(tc.errorPredicate(err))
				}
			} else {
				s.NoError(err)
				s.NotEqual(tc.user.Password, userCopy.Password) // Password should be hashed
				s.Equal(models.RegularStaffRole, userCopy.Role)
			}
		})
	}
}

// TestChangePassword tests the ChangePassword method
func (s *UserServiceTestSuite) TestChangePassword() {
	testCases := []struct {
		name           string
		username       string
		oldPassword    string
		newPassword    string
		setupMock      func()
		expectedError  bool
		errorPredicate func(error) bool
	}{
		{
			name:        "Success",
			username:    "testuser",
			oldPassword: "oldpassword",
			newPassword: "newpassword",
			setupMock: func() {
				hashedOldPassword, _ := bcrypt.GenerateFromPassword([]byte("oldpassword"), bcrypt.DefaultCost)
				mockUser := &models.User{
					Username: "testuser",
					Password: string(hashedOldPassword),
					Role:     models.RegularStaffRole,
				}
				s.mockRepo.On("GetByUsername", "testuser").Return(mockUser, nil)
				s.mockRepo.On("Replace", mock.AnythingOfType("*models.User")).Return(nil)
			},
			expectedError: false,
		},
		{
			name:        "WrongOldPassword",
			username:    "testuser",
			oldPassword: "wrongpassword",
			newPassword: "newpassword",
			setupMock: func() {
				hashedOldPassword, _ := bcrypt.GenerateFromPassword([]byte("oldpassword"), bcrypt.DefaultCost)
				mockUser := &models.User{
					Username: "testuser",
					Password: string(hashedOldPassword),
					Role:     models.RegularStaffRole,
				}
				s.mockRepo.On("GetByUsername", "testuser").Return(mockUser, nil)
			},
			expectedError:  true,
			errorPredicate: apperrors.IsUnauthorizedErr,
		},
		{
			name:        "UserNotFound",
			username:    "nonexistentuser",
			oldPassword: "oldpassword",
			newPassword: "newpassword",
			setupMock: func() {
				notFoundErr := apperrors.NewNotFoundErr("User not found", nil)
				s.mockRepo.On("GetByUsername", "nonexistentuser").Return(nil, notFoundErr)
			},
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Setup fresh mocks
			s.SetupTest()

			// Setup mock expectations
			tc.setupMock()

			// Act
			err := s.userService.ChangePassword(tc.username, tc.oldPassword, tc.newPassword)

			// Assert
			if tc.expectedError {
				s.Error(err)
				if tc.errorPredicate != nil {
					s.True(tc.errorPredicate(err))
				}
			} else {
				s.NoError(err)
			}
		})
	}
}

// Run the test suite
func TestUserServiceSuite(t *testing.T) {
	suite.Run(t, new(UserServiceTestSuite))
}

// Mock implementation of UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) WithTransaction(fn func(repository repositories.UserRepository) error) error {
	return fn(m)
}

func (m *MockUserRepository) Create(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByUsername(username string) (*models.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Update(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) Exists(username string) (bool, error) {
	args := m.Called(username)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) GetByRole(role models.Role) ([]*models.User, error) {
	args := m.Called(role)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockUserRepository) DeleteByUsername(username string) error {
	args := m.Called(username)
	return args.Error(0)
}
