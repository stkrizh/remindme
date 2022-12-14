package signupwithemail

import (
	"context"
	"errors"
	c "remindme/internal/core/domain/common"
	"remindme/internal/core/domain/logging"
	uow "remindme/internal/core/domain/unit_of_work"
	"remindme/internal/core/domain/user"
	"remindme/internal/core/services"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

const (
	ACTIVATION_TOKEN = "test"
	EMAIL            = c.Email("test@test.test")
	RAW_PASSWORD     = user.RawPassword("test-password")
)

var NOW time.Time = time.Now().UTC()

type testSuite struct {
	suite.Suite
	Logger                   *logging.FakeLogger
	UnitOfWork               *uow.FakeUnitOfWork
	PasswordHasher           *user.FakePasswordHasher
	ActivationTokenGenerator *user.FakeActivationTokenGenerator
	Service                  services.Service[Input, Result]
}

func (suite *testSuite) SetupTest() {
	suite.Logger = logging.NewFakeLogger()
	suite.UnitOfWork = uow.NewFakeUnitOfWork()
	suite.PasswordHasher = user.NewFakePasswordHasher()
	suite.ActivationTokenGenerator = user.NewFakeActivationTokenGenerator(ACTIVATION_TOKEN)
	suite.Service = New(
		suite.Logger,
		suite.UnitOfWork,
		suite.PasswordHasher,
		suite.ActivationTokenGenerator,
		func() time.Time { return NOW },
	)
}

func TestSignUpWithEmailService(t *testing.T) {
	suite.Run(t, new(testSuite))
}

func (suite *testSuite) TestSuccess() {
	context := context.Background()
	result, err := suite.Service.Run(context, Input{Email: EMAIL, Password: RAW_PASSWORD})

	assert := suite.Require()
	assert.Nil(err)
	assert.NotEqual(user.ID(0), result.User.ID)
	assert.Equal(NOW, result.User.CreatedAt)
	assert.True(result.User.Email.IsPresent)
	assert.Equal(EMAIL, result.User.Email.Value)
	assert.True(result.User.PasswordHash.IsPresent)
	assert.NotEqual(RAW_PASSWORD, result.User.PasswordHash.Value)
	assert.False(result.User.Identity.IsPresent)
	assert.True(suite.UnitOfWork.Context.WasCommitCalled)
}

func (suite *testSuite) TestEmailAlreadyExistsError() {
	ctx := context.Background()
	suite.UnitOfWork.Context.UserRepository.Create(
		ctx,
		user.CreateUserInput{
			Email:        c.NewOptional(EMAIL, true),
			PasswordHash: c.NewOptional(user.PasswordHash("test"), true),
			CreatedAt:    NOW,
		},
	)

	_, err := suite.Service.Run(ctx, Input{Email: EMAIL, Password: RAW_PASSWORD})

	assert := suite.Require()
	assert.NotNil(err)
	assert.True(errors.Is(err, user.ErrEmailAlreadyExists))
	assert.False(suite.UnitOfWork.Context.WasCommitCalled)
	assert.True(suite.UnitOfWork.Context.WasRollbackCalled)
}
