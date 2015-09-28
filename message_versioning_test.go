package messaging_spike

import "testing"

func TestUpgradeMessages(t *testing.T) {
	v1 := VersionableMessage{version: 1, content: V1Payload("first")}
	v2 := VersionableMessage{version: 2, content: V2Payload{content: map[string]string{"foo": "myFoo", "value": "second"}}}
	latestVersion := VersionableMessage{version: 3, content: LatestPayload{foo: "anotherFoo", bar: "anotherBar", value: "latest"}}
	latestVersionConsumer := &LatestMessageVersionOnlyConsumer{}
	c := NewMessageUpgradeDecorator(latestVersionConsumer)

	for _, spec := range []struct {
		m               VersionableMessage
		foo, bar, value string
	}{
		{v1, "defaultFoo", "defaultBar", "first"},
		{v2, "myFoo", "defaultBar", "second"},
		{latestVersion, "anotherFoo", "anotherBar", "latest"},
	} {
		// when
		if err := c.OnEvent(spec.m); err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		// then
		if got, exp := latestVersionConsumer.foo, spec.foo; got != exp {
			t.Errorf("expected %q but got %q", exp, got)
		}
		if got, exp := latestVersionConsumer.bar, spec.bar; got != exp {
			t.Errorf("expected %q but got %q", exp, got)
		}
		if got, exp := latestVersionConsumer.state, spec.value; got != exp {
			t.Errorf("expected %q but got %q", exp, got)
		}
	}
}
