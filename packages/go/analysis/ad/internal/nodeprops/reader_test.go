package nodeprops

import (
	"errors"
	"testing"

	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrorReadDomainSIDandNameAsString(t *testing.T) {
	tests := []struct {
		name    string
		message string
		want    string
	}{
		{
			name:    "empty message",
			message: "",
			want:    "failed to read domain SID and name: ",
		},
		{
			name:    "with message",
			message: "test error",
			want:    "failed to read domain SID and name: test error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errorReadDomainSIDandNameAsString(tt.message)
			assert.Equal(t, tt.want, err.Error())
		})
	}
}

func TestReadDomainIDandNameAsString(t *testing.T) {
	tests := []struct {
		name          string
		node          *graph.Node
		wantSID       string
		wantName      string
		wantErrSubstr string
	}{
		{
			name:          "nil node",
			node:          nil,
			wantSID:       "",
			wantName:      "",
			wantErrSubstr: "given nodeToRead is nil",
		},
		{
			name: "missing domain SID",
			node: &graph.Node{
				Properties: graph.NewProperties(),
			},
			wantSID:       "",
			wantName:      "",
			wantErrSubstr: "read domain SID property value is nil",
		},
		{
			name: "domain SID error",
			node: &graph.Node{
				Properties: func() *graph.Properties {
					props := graph.NewProperties()
					props.Set(ad.DomainSID.String(), errors.New("test error"))
					return props
				}(),
			},
			wantSID:       "",
			wantName:      "",
			wantErrSubstr: "failed to read domain SID and name",
		},
		{
			name: "empty domain SID",
			node: &graph.Node{
				Properties: func() *graph.Properties {
					props := graph.NewProperties()
					props.Set(ad.DomainSID.String(), "")
					return props
				}(),
			},
			wantSID:       "",
			wantName:      "",
			wantErrSubstr: "read domain SID is empty or blank",
		},
		{
			name: "empty domain name",
			node: &graph.Node{
				Properties: func() *graph.Properties {
					props := graph.NewProperties()
					props.Set(ad.DomainSID.String(), "S-1-5-21-123456789-123456789-123456789")
					props.Set(common.Name.String(), "")
					return props
				}(),
			},
			wantSID:       "S-1-5-21-123456789-123456789-123456789",
			wantName:      "",
			wantErrSubstr: "read domain name is empty or blank",
		},
		{
			name: "successful read",
			node: &graph.Node{
				Properties: func() *graph.Properties {
					props := graph.NewProperties()
					props.Set(ad.DomainSID.String(), "S-1-5-21-123456789-123456789-123456789")
					props.Set(common.Name.String(), "EXAMPLE.COM")
					return props
				}(),
			},
			wantSID:       "S-1-5-21-123456789-123456789-123456789",
			wantName:      "EXAMPLE.COM",
			wantErrSubstr: "",
		},
		{
			name: "missing domain name",
			node: &graph.Node{
				Properties: func() *graph.Properties {
					props := graph.NewProperties()
					props.Set(ad.DomainSID.String(), "S-1-5-21-123456789-123456789-123456789")
					// No name property set
					return props
				}(),
			},
			wantSID:       "S-1-5-21-123456789-123456789-123456789",
			wantName:      "",
			wantErrSubstr: "read domain name property value is nil",
		},
		{
			name: "domain SID with whitespace",
			node: &graph.Node{
				Properties: func() *graph.Properties {
					props := graph.NewProperties()
					props.Set(ad.DomainSID.String(), "  S-1-5-21-123456789-123456789-123456789  ")
					props.Set(common.Name.String(), "EXAMPLE.COM")
					return props
				}(),
			},
			wantSID:       "  S-1-5-21-123456789-123456789-123456789  ",
			wantName:      "EXAMPLE.COM",
			wantErrSubstr: "",
		},
		{
			name: "domain name with whitespace",
			node: &graph.Node{
				Properties: func() *graph.Properties {
					props := graph.NewProperties()
					props.Set(ad.DomainSID.String(), "S-1-5-21-123456789-123456789-123456789")
					props.Set(common.Name.String(), "  EXAMPLE.COM  ")
					return props
				}(),
			},
			wantSID:       "S-1-5-21-123456789-123456789-123456789",
			wantName:      "  EXAMPLE.COM  ",
			wantErrSubstr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSID, gotName, err := ReadDomainIDandNameAsString(tt.node)

			assert.Equal(t, tt.wantSID, gotSID, "SID value should match expected")
			assert.Equal(t, tt.wantName, gotName, "Name value should match expected")

			if tt.wantErrSubstr != "" {
				require.Error(t, err, "Expected an error but got none")
				assert.Contains(t, err.Error(), tt.wantErrSubstr, "Error message should contain expected substring")
			} else {
				assert.NoError(t, err, "Expected no error but got one")
			}
		})
	}
}
