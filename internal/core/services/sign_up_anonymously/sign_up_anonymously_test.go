package signupanonymously

import (
	"context"
	"net/netip"
	"remindme/internal/core/domain/channel"
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
	IDENTITY               = "test-identity"
	SESSION_TOKEN          = "test-session-token"
	INTERNAL_CHANNEL_TOKEN = "test-internal-channel-token"
)

var (
	Now    time.Time   = time.Now().UTC()
	Limits user.Limits = user.Limits{
		EmailChannelCount:    c.NewOptional(uint32(1), true),
		TelegramChannelCount: c.NewOptional(uint32(1), true),
	}
)

type testSuite struct {
	suite.Suite
	Logger                       *logging.FakeLogger
	UnitOfWork                   *uow.FakeUnitOfWork
	IdentityGenerator            *user.FakeIdentityGenerator
	SessionTokenGenerator        *user.FakeSessionTokenGenerator
	InternaChannelTokenGenerator *channel.TestInternalChannelTokenGenerator
	Service                      services.Service[Input, Result]
}

func (suite *testSuite) SetupTest() {
	suite.Logger = logging.NewFakeLogger()
	suite.UnitOfWork = uow.NewFakeUnitOfWork()
	suite.IdentityGenerator = user.NewFakeIdentityGenerator(IDENTITY)
	suite.SessionTokenGenerator = user.NewFakeSessionTokenGenerator(SESSION_TOKEN)
	suite.InternaChannelTokenGenerator = channel.NewTestInternalChannelTokenGenerator(INTERNAL_CHANNEL_TOKEN)
	suite.Service = New(
		suite.Logger,
		suite.UnitOfWork,
		suite.IdentityGenerator,
		suite.SessionTokenGenerator,
		suite.InternaChannelTokenGenerator,
		func() time.Time { return Now },
		Limits,
	)
}

func TestSignUpAnonymouslyService(t *testing.T) {
	suite.Run(t, new(testSuite))
}

func (suite *testSuite) TestSuccess() {
	ctx := context.Background()

	result, err := suite.Service.Run(ctx, Input{IP: netip.MustParseAddr("192.168.1.1")})
	createdUser := result.User
	sessionToken := result.Token

	assert := suite.Require()
	assert.Nil(err)
	assert.False(createdUser.Email.IsPresent)
	assert.False(createdUser.PasswordHash.IsPresent)
	assert.Equal(user.Identity(IDENTITY), createdUser.Identity.Value)
	assert.True(createdUser.Identity.IsPresent)
	assert.Equal(Now, createdUser.CreatedAt)
	assert.True(createdUser.ActivatedAt.IsPresent)
	assert.Equal(Now, createdUser.ActivatedAt.Value)

	assert.Equal(user.SessionToken(SESSION_TOKEN), sessionToken)

	u, _ := suite.UnitOfWork.Context.SessionRepository.GetUserByToken(ctx, sessionToken)
	assert.Equal(createdUser, u)

	assert.True(suite.UnitOfWork.Context.WasCommitCalled)
}

func (suite *testSuite) TestSuccessLimitsCreated() {
	ctx := context.Background()

	_, err := suite.Service.Run(ctx, Input{IP: netip.MustParseAddr("192.168.1.1")})

	assert := suite.Require()
	assert.Nil(err)
	assert.Equal(1, len(suite.UnitOfWork.Context.LimitsRepository.Created))
	assert.Equal(Limits, suite.UnitOfWork.Context.LimitsRepository.Created[0])
	assert.True(suite.UnitOfWork.Context.WasCommitCalled)
}

func (suite *testSuite) TestSuccessInternalChannelCreated() {
	ctx := context.Background()

	_, err := suite.Service.Run(ctx, Input{IP: netip.MustParseAddr("192.168.1.1")})

	assert := suite.Require()
	assert.Nil(err)
	assert.Equal(1, len(suite.UnitOfWork.Context.ChannelRepository.Created))
	settings := (suite.UnitOfWork.Context.ChannelRepository.Created[0].Settings).(*channel.InternalSettings)
	assert.Equal(channel.InternalChannelToken(INTERNAL_CHANNEL_TOKEN), settings.Token)
	assert.True(suite.UnitOfWork.Context.WasCommitCalled)
}
