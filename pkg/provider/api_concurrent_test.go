package provider

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/korotovsky/slack-mcp-server/pkg/provider/edge"
	"github.com/slack-go/slack"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

// MockSlackAPI implements SlackAPI interface for testing
type MockSlackAPI struct {
	users    []slack.User
	channels []slack.Channel
}

func (m *MockSlackAPI) AuthTest() (*slack.AuthTestResponse, error) {
	return &slack.AuthTestResponse{
		URL:    "https://test.slack.com",
		Team:   "Test Team",
		User:   "test",
		TeamID: "T123456",
		UserID: "U123456",
	}, nil
}

func (m *MockSlackAPI) AuthTestContext(ctx context.Context) (*slack.AuthTestResponse, error) {
	return m.AuthTest()
}

func (m *MockSlackAPI) GetUsersContext(ctx context.Context, options ...slack.GetUsersOption) ([]slack.User, error) {
	// Simulate some delay
	time.Sleep(10 * time.Millisecond)
	return m.users, nil
}

func (m *MockSlackAPI) GetUsersInfo(users ...string) (*[]slack.User, error) {
	return &m.users, nil
}

func (m *MockSlackAPI) PostMessageContext(ctx context.Context, channel string, options ...slack.MsgOption) (string, string, error) {
	return "123.456", "123.456", nil
}

func (m *MockSlackAPI) MarkConversationContext(ctx context.Context, channel, ts string) error {
	return nil
}

func (m *MockSlackAPI) GetConversationHistoryContext(ctx context.Context, params *slack.GetConversationHistoryParameters) (*slack.GetConversationHistoryResponse, error) {
	return &slack.GetConversationHistoryResponse{}, nil
}

func (m *MockSlackAPI) GetConversationRepliesContext(ctx context.Context, params *slack.GetConversationRepliesParameters) (msgs []slack.Message, hasMore bool, nextCursor string, err error) {
	return []slack.Message{}, false, "", nil
}

func (m *MockSlackAPI) SearchContext(ctx context.Context, query string, params slack.SearchParameters) (*slack.SearchMessages, *slack.SearchFiles, error) {
	return &slack.SearchMessages{}, &slack.SearchFiles{}, nil
}

func (m *MockSlackAPI) GetConversationsContext(ctx context.Context, params *slack.GetConversationsParameters) ([]slack.Channel, string, error) {
	// Simulate some delay
	time.Sleep(10 * time.Millisecond)
	return m.channels, "", nil
}

func (m *MockSlackAPI) ClientUserBoot(ctx context.Context) (*edge.ClientUserBootResponse, error) {
	return &edge.ClientUserBootResponse{}, nil
}

func (m *MockSlackAPI) IsBotToken() bool {
	return false
}

// TestConcurrentCacheAccess tests that concurrent access to caches is safe
func TestConcurrentCacheAccess(t *testing.T) {
	logger := zap.NewNop()

	ap := &ApiProvider{
		transport: "stdio",
		client: &MockSlackAPI{
			users: []slack.User{
				{ID: "U001", Name: "user1"},
				{ID: "U002", Name: "user2"},
			},
			channels: []slack.Channel{
				{
					GroupConversation: slack.GroupConversation{
						Conversation: slack.Conversation{ID: "C001"},
						Name:         "channel1",
					},
				},
				{
					GroupConversation: slack.GroupConversation{
						Conversation: slack.Conversation{ID: "C002"},
						Name:         "channel2",
					},
				},
			},
		},
		logger:        logger,
		rateLimiter:   rate.NewLimiter(rate.Every(100*time.Millisecond), 10),
		users:         make(map[string]slack.User),
		usersInv:      make(map[string]string),
		channels:      make(map[string]Channel),
		channelsInv:   make(map[string]string),
		usersCache:    "/tmp/test_users_cache.json",
		channelsCache: "/tmp/test_channels_cache.json",
	}

	ctx := context.Background()
	var wg sync.WaitGroup

	// Test concurrent refresh operations
	wg.Add(2)
	go func() {
		defer wg.Done()
		if err := ap.RefreshUsers(ctx); err != nil {
			t.Errorf("RefreshUsers failed: %v", err)
		}
	}()

	go func() {
		defer wg.Done()
		if err := ap.RefreshChannels(ctx); err != nil {
			t.Errorf("RefreshChannels failed: %v", err)
		}
	}()

	wg.Wait()

	// Test concurrent reads while refreshing
	wg.Add(4)

	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			_ = ap.ProvideUsersMap()
			time.Sleep(5 * time.Millisecond)
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			_ = ap.ProvideChannelsMaps()
			time.Sleep(5 * time.Millisecond)
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 5; i++ {
			if err := ap.RefreshUsers(ctx); err != nil {
				t.Errorf("RefreshUsers failed during concurrent access: %v", err)
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 5; i++ {
			if err := ap.RefreshChannels(ctx); err != nil {
				t.Errorf("RefreshChannels failed during concurrent access: %v", err)
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()

	wg.Wait()

	// Test IsReady concurrent access
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < 20; i++ {
			ready, _ := ap.IsReady()
			if !ready && i > 10 {
				t.Error("Expected IsReady to return true after initialization")
			}
			time.Sleep(2 * time.Millisecond)
		}
	}()

	go func() {
		defer wg.Done()
		ap.cacheMu.Lock()
		ap.usersReady = true
		ap.cacheMu.Unlock()

		ap.cacheMu.Lock()
		ap.channelsReady = true
		ap.cacheMu.Unlock()
	}()

	wg.Wait()
}

// TestConcurrentChannelOperations tests concurrent channel operations
func TestConcurrentChannelOperations(t *testing.T) {
	logger := zap.NewNop()

	ap := &ApiProvider{
		transport: "stdio",
		client: &MockSlackAPI{
			channels: []slack.Channel{
				{
					GroupConversation: slack.GroupConversation{
						Conversation: slack.Conversation{ID: "C001", IsPrivate: false},
						Name:         "public1",
					},
				},
				{
					GroupConversation: slack.GroupConversation{
						Conversation: slack.Conversation{ID: "C002", IsPrivate: true},
						Name:         "private1",
					},
				},
			},
		},
		logger:      logger,
		rateLimiter: rate.NewLimiter(rate.Every(10*time.Millisecond), 10),
		users:       make(map[string]slack.User),
		usersInv:    make(map[string]string),
		channels:    make(map[string]Channel),
		channelsInv: make(map[string]string),
	}

	ctx := context.Background()
	var wg sync.WaitGroup

	// Concurrent GetChannels calls
	wg.Add(5)
	for i := 0; i < 5; i++ {
		go func() {
			defer wg.Done()
			channels := ap.GetChannels(ctx, []string{"public_channel", "private_channel"})
			if len(channels) == 0 && ap.channelsReady {
				t.Error("Expected channels to be returned")
			}
		}()
	}

	wg.Wait()
}
