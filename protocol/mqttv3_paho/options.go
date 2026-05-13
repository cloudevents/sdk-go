package mqttv3_paho

type Option func(*Protocol) error

func WithSubscribeMap(subscribeMap map[string]byte) Option {
	return func(p *Protocol) error {
		if subscribeMap == nil || p == nil {
			return nil
		}

		if p.subscriptions == nil {
			p.subscriptions = make(map[string]byte)
		}

		for topic, qos := range subscribeMap {
			p.subscriptions[topic] = qos
		}
		return nil
	}
}

func WithSubscribeTopic(topic string, qos byte) Option {
	return func(p *Protocol) error {
		if topic == "" || p == nil {
			return nil
		}

		if p.subscriptions == nil {
			p.subscriptions = make(map[string]byte)
		}

		p.subscriptions[topic] = qos
		return nil
	}
}

func WithPublishTopic(topic string, qos byte, retained bool) Option {
	return func(p *Protocol) error {
		if topic == "" || p == nil {
			return nil
		}

		p.topic = topic
		p.qos = qos
		p.retained = retained

		return nil
	}
}

func WithDisconnectQuiesce(quiesce uint) Option {
	return func(p *Protocol) error {
		if quiesce == 0 || p == nil {
			return nil
		}

		p.quiesce = quiesce

		return nil
	}
}
