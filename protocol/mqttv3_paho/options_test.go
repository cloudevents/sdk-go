package mqttv3_paho

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWithSubscribeMap(t *testing.T) {
	t.Run("should do nothing for nil map", func(t *testing.T) {
		p := &Protocol{subscriptions: map[string]byte{"existing/topic": 1}}
		err := WithSubscribeMap(nil)(p)

		require.NoError(t, err)
		require.Equal(t, map[string]byte{"existing/topic": 1}, p.subscriptions)
	})

	t.Run("should do nothing for nil protocol", func(t *testing.T) {
		err := WithSubscribeMap(map[string]byte{"test/topic": 1})(nil)

		require.NoError(t, err)
	})

	t.Run("should add subscriptions", func(t *testing.T) {
		p := &Protocol{}
		err := WithSubscribeMap(map[string]byte{"test/topic": 1, "another/topic": 0})(p)

		require.NoError(t, err)
		require.Equal(t, map[string]byte{"test/topic": 1, "another/topic": 0}, p.subscriptions)
	})
}

func TestWithSubscribeTopic(t *testing.T) {
	t.Run("should do nothing for empty topic", func(t *testing.T) {
		p := &Protocol{subscriptions: map[string]byte{"existing/topic": 1}}
		err := WithSubscribeTopic("", 1)(p)

		require.NoError(t, err)
		require.Equal(t, map[string]byte{"existing/topic": 1}, p.subscriptions)
	})

	t.Run("should do nothing for nil protocol", func(t *testing.T) {
		err := WithSubscribeTopic("test/topic", 1)(nil)

		require.NoError(t, err)
	})

	t.Run("should add subscription", func(t *testing.T) {
		p := &Protocol{}
		err := WithSubscribeTopic("test/topic", 1)(p)

		require.NoError(t, err)
		require.Equal(t, map[string]byte{"test/topic": 1}, p.subscriptions)
	})
}

func TestWithDisconnectQuiesce(t *testing.T) {
	t.Run("should do nothing for zero quiesce", func(t *testing.T) {
		p := &Protocol{quiesce: 5}
		err := WithDisconnectQuiesce(0)(p)

		require.NoError(t, err)
		require.Equal(t, uint(5), p.quiesce)
	})

	t.Run("should do nothing for nil protocol", func(t *testing.T) {
		err := WithDisconnectQuiesce(10)(nil)

		require.NoError(t, err)
	})

	t.Run("should set quiesce", func(t *testing.T) {
		p := &Protocol{}
		err := WithDisconnectQuiesce(10)(p)

		require.NoError(t, err)
		require.Equal(t, uint(10), p.quiesce)
	})
}

func TestWithPublishTopic(t *testing.T) {
	t.Run("should not set empty topic", func(t *testing.T) {
		p := &Protocol{}
		err := WithPublishTopic("", 1, true)(p)

		require.NoError(t, err)
		require.Equal(t, "", p.topic)
		require.Equal(t, byte(0), p.qos)
		require.Equal(t, false, p.retained)
	})

	t.Run("should do nothing for nil protocol", func(t *testing.T) {
		err := WithPublishTopic("test/topic", 1, true)(nil)

		require.NoError(t, err)
	})

	t.Run("should set topic", func(t *testing.T) {
		p := &Protocol{}
		err := WithPublishTopic("test/topic", 1, true)(p)

		require.NoError(t, err)
		require.Equal(t, "test/topic", p.topic)
		require.Equal(t, byte(1), p.qos)
		require.Equal(t, true, p.retained)
	})
}
