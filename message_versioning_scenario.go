package messaging_spike

import "fmt"

const (
	v1      uint = 1
	v2      uint = 2
	vLatest uint = 3
)

type MessagePayload interface{}

type VersionableMessage struct {
	version uint
	content MessagePayload
}

func (v VersionableMessage) Version() uint {
	return v.version
}
func (v VersionableMessage) Content() MessagePayload {
	return v.content
}

type V1Payload string
type V2Payload struct {
	content map[string]string
}

type LatestPayload struct {
	foo, bar, value string
}

type Migration func(MessagePayload) (uint, MessagePayload, error)

type MessageUpgradeDecorator struct {
	c          *LatestMessageVersionOnlyConsumer
	migrations map[uint]Migration
}

func NewMessageUpgradeDecorator(c *LatestMessageVersionOnlyConsumer) *MessageUpgradeDecorator {
	d := &MessageUpgradeDecorator{c: c}
	d.migrations = map[uint]Migration{
		v1: migrateMessageV1ToV2,
		v2: migrateMessageV2ToVLatest,
	}
	return d
}

func (f *MessageUpgradeDecorator) OnEvent(e VersionableMessage) error {
	v := e.Version()
	c := e.Content()
	for upgrade, ok := f.migrations[v]; ok; upgrade, ok = f.migrations[v] {
		var err error
		if v, c, err = upgrade(c); err != nil {
			return fmt.Errorf("%T failed to upgrade: %v", upgrade, err)
		}
	}
	l, ok := c.(LatestPayload)
	if !ok {
		return fmt.Errorf("migration failed. unsupported content type: %T, Version: %d", c, v)
	}
	return f.c.OnEvent(l)
}

type LatestMessageVersionOnlyConsumer struct {
	state, foo, bar string
}

func (l *LatestMessageVersionOnlyConsumer) OnEvent(p LatestPayload) error {
	l.state = p.value
	l.foo = p.foo
	l.bar = p.bar
	return nil
}

func migrateMessageV1ToV2(s MessagePayload) (uint, MessagePayload, error) {
	p, ok := s.(V1Payload)
	if !ok {
		return 0, nil, fmt.Errorf("unsupported type %T for version %d", p, v1)
	}
	return v2, V2Payload{content: map[string]string{"foo": "defaultFoo", "value": string(p)}}, nil
}

func migrateMessageV2ToVLatest(s MessagePayload) (uint, MessagePayload, error) {
	p, ok := s.(V2Payload)
	if !ok {
		return 0, nil, fmt.Errorf("unsupported type %T for version %d", p, v2)
	}
	return vLatest, LatestPayload{foo: p.content["foo"], bar: "defaultBar", value: p.content["value"]}, nil
}
