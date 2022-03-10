package otel

import (
	"github.com/kolesa-team/servicectx"
	"go.opentelemetry.io/otel/baggage"
)

func CreateBaggageMembers(opts servicectx.Options) []baggage.Member {
	var result []baggage.Member

	for key, value := range opts.HeaderMap() {
		member, err := baggage.NewMember(key, value)
		if err != nil {
			continue
		}

		result = append(result, member)
	}

	return result
}

func InjectIntoBaggage(bag baggage.Baggage, opts servicectx.Options) baggage.Baggage {
	for _, member := range CreateBaggageMembers(opts) {
		bag, _ = bag.SetMember(member)
	}

	return bag
}

func FromBaggage(bag baggage.Baggage) servicectx.Options {
	opts := servicectx.New()

	for _, member := range bag.Members() {
		serviceName, option, ok := servicectx.ParseOptionName(member.Key())
		if !ok {
			continue
		}

		opts.Set(serviceName, option, member.Value())
	}

	return opts
}
