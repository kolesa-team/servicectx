package otel

import (
	"context"
	"github.com/kolesa-team/servicectx"
	"go.opentelemetry.io/otel/baggage"
)

// CreateBaggageMembers creates opentelemetry "baggage members" from properties
func CreateBaggageMembers(props servicectx.Properties) []baggage.Member {
	var result []baggage.Member

	for key, value := range props.HeaderMap() {
		member, err := baggage.NewMember(key, value)
		if err != nil {
			continue
		}

		result = append(result, member)
	}

	return result
}

// InjectIntoBaggage adds properties into opentelemetry baggage
func InjectIntoBaggage(bag baggage.Baggage, props servicectx.Properties) baggage.Baggage {
	for _, member := range CreateBaggageMembers(props) {
		bag, _ = bag.SetMember(member)
	}

	return bag
}

// InjectIntoContext adds the properties into OpenTelemetry Baggage, then adds the baggage into the context.
func InjectIntoContext(ctx context.Context, props servicectx.Properties) context.Context {
	bag := baggage.FromContext(ctx)
	bag = InjectIntoBaggage(bag, props)
	ctx = baggage.ContextWithBaggage(ctx, bag)

	return ctx
}

// FromBaggage retries properties from baggage
func FromBaggage(bag baggage.Baggage) servicectx.Properties {
	props := servicectx.New()

	for _, member := range bag.Members() {
		serviceName, option, ok := servicectx.ParsePropertyName(member.Key())
		if !ok {
			continue
		}

		props.Set(serviceName, option, member.Value())
	}

	return props
}

// FromContextAndBaggage retrieves properties from Go context and from baggage.
// This is convenient when the properties can be set both in application code via context and from outside world by opentelemetry.
// The properties from Go context have a preference over the baggage.
func FromContextAndBaggage(ctx context.Context, bag baggage.Baggage) servicectx.Properties {
	return FromBaggage(bag).Merge(servicectx.FromContext(ctx))
}
