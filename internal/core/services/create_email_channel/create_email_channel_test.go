package createemailchannel

import (
	"context"
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
	EMAIL              = c.Email("test@test.test")
	USER_ID            = user.ID(42)
	VERIFICATION_TOKEN = channel.VerificationToken("test")
)

var Now time.Time = time.Now().UTC()

type testSuite struct {
	suite.Suite
	Logger         *logging.FakeLogger
	UnitOfWork     *uow.FakeUnitOfWork
	TokenGenerator *channel.FakeVerificationTokenGenerator
	service        services.Service[Input, Result]
}

func (suite *testSuite) SetupTest() {
	suite.Logger = logging.NewFakeLogger()
	suite.UnitOfWork = uow.NewFakeUnitOfWork()
	suite.TokenGenerator = channel.NewFakeVerificationTokenGenerator(VERIFICATION_TOKEN)
	suite.service = New(
		suite.Logger,
		suite.UnitOfWork,
		suite.TokenGenerator,
		func() time.Time { return Now },
	)
}

func (suite *testSuite) TearDownTest() {
	suite.UnitOfWork.Channels().Created = make([]channel.Channel, 0)
}

func TestCreateEmailChannelService(t *testing.T) {
	suite.Run(t, new(testSuite))
}

func (s *testSuite) TestSuccess() {
	cases := []struct {
		id          string
		limits      user.Limits
		actualCount uint
	}{
		{
			id:          "1",
			limits:      user.Limits{},
			actualCount: 0,
		},
		{
			id:          "2",
			limits:      user.Limits{},
			actualCount: 10,
		},
		{
			id:          "3",
			limits:      user.Limits{EmailChannelCount: c.NewOptional(uint32(1), true)},
			actualCount: 0,
		},
		{
			id:          "4",
			limits:      user.Limits{EmailChannelCount: c.NewOptional(uint32(2), true)},
			actualCount: 1,
		},
	}

	for _, testcase := range cases {
		s.Run(testcase.id, func() {
			s.SetupTest()
			s.UnitOfWork.Limits().Limits = testcase.limits
			s.UnitOfWork.Channels().CountChannels = testcase.actualCount

			result, err := s.service.Run(context.Background(), Input{
				Email:  EMAIL,
				UserID: USER_ID,
			})

			assert := s.Require()
			assert.Nil(err)

			options := s.UnitOfWork.Channels().Options
			assert.Len(options, 1)
			assert.Equal(c.NewOptional(USER_ID, true), options[0].UserIDEquals)
			assert.Equal(c.NewOptional(channel.Email, true), options[0].TypeEquals)

			createdChannels := s.UnitOfWork.Channels().Created
			assert.Len(createdChannels, 1)

			createdChannel := createdChannels[0]
			assert.Equal(createdChannel, result.Channel)
			assert.Equal(Now, createdChannel.CreatedAt)
			assert.Equal(USER_ID, createdChannel.CreatedBy)
			assert.Equal(channel.Email, createdChannel.Type)
			assert.Equal(EMAIL, createdChannel.Settings.(*channel.EmailSettings).Email)
			assert.Equal(c.NewOptional(VERIFICATION_TOKEN, true), createdChannel.VerificationToken)
			assert.Equal(VERIFICATION_TOKEN, result.VerificationToken)
			assert.False(createdChannel.IsVerified())

			assert.True(s.UnitOfWork.Context.WasCommitCalled)
		})
	}
}

func (s *testSuite) TestChannelIsNotCreatedIfLimitExceeded() {
	cases := []struct {
		id          string
		limits      user.Limits
		actualCount uint
	}{
		{
			id:          "1",
			limits:      user.Limits{EmailChannelCount: c.NewOptional(uint32(0), true)},
			actualCount: 0,
		},
		{
			id:          "2",
			limits:      user.Limits{EmailChannelCount: c.NewOptional(uint32(1), true)},
			actualCount: 1,
		},
		{
			id:          "3",
			limits:      user.Limits{EmailChannelCount: c.NewOptional(uint32(10), true)},
			actualCount: 20,
		},
	}

	for _, testcase := range cases {
		s.Run(testcase.id, func() {
			s.SetupTest()
			s.UnitOfWork.Limits().Limits = testcase.limits
			s.UnitOfWork.Channels().CountChannels = testcase.actualCount

			_, err := s.service.Run(context.Background(), Input{
				Email:  EMAIL,
				UserID: USER_ID,
			})

			assert := s.Require()
			assert.ErrorIs(err, user.ErrLimitEmailChannelCountExceeded)

			createdChannels := s.UnitOfWork.Channels().Created
			assert.Len(createdChannels, 0)

			assert.False(s.UnitOfWork.Context.WasCommitCalled)
		})
	}
}
