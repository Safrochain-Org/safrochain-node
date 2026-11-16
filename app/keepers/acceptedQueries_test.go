package keepers_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/cosmos/gogoproto/proto"

	"github.com/Safrochain_Org/safrochain/app/keepers"
)

// TestAcceptedQueriesResponseTypes asserts that every accepted query path maps
// to the exact response type declared by the gRPC method.
func TestAcceptedQueriesResponseTypes(t *testing.T) {
	for path, newMsg := range keepers.AcceptedQueries() {
		t.Run(path, func(t *testing.T) {
			require.True(t, strings.HasPrefix(path, "/"))
			parts := strings.Split(path[1:], "/")
			require.Len(t, parts, 2, "path must be /<service>/<method>")

			desc, err := proto.HybridResolver.FindDescriptorByName(protoreflect.FullName(parts[0]))
			require.NoError(t, err, "service not found")
			svc, ok := desc.(protoreflect.ServiceDescriptor)
			require.True(t, ok, "%s is not a service", parts[0])

			method := svc.Methods().ByName(protoreflect.Name(parts[1]))
			require.NotNil(t, method, "method %s not found", parts[1])

			require.Equal(t, string(method.Output().FullName()), proto.MessageName(newMsg()))
		})
	}
}
