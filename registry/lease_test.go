package registry

import (
	"reflect"
	"testing"
	"time"

	"github.com/coreos/fleet/etcd"
)

func TestSerializeLeaseMetadata(t *testing.T) {
	tests := []struct {
		machID string
		ver    int
		want   string
	}{
		{
			machID: "XXX",
			ver:    9,
			want:   `{"MachineID":"XXX","Version":9}`,
		},
		{
			machID: "XXX",
			ver:    0,
			want:   `{"MachineID":"XXX","Version":0}`,
		},
	}

	for i, tt := range tests {
		got, err := serializeLeaseMetadata(tt.machID, tt.ver)
		if err != nil {
			t.Errorf("case %d: unexpected err=%v", i, err)
			continue
		}
		if tt.want != got {
			t.Errorf("case %d: incorrect output from serializeLeaseMetadata\nwant=%s\ngot=%s", i, tt.want, got)
		}
	}
}

func TestLeaseFromResult(t *testing.T) {
	tests := []struct {
		res  etcd.Result
		want etcdLease
	}{
		// typical case
		{
			res: etcd.Result{
				Node: &etcd.Node{
					Key:           "/foo/bar",
					ModifiedIndex: 12,
					TTL:           9,
					Value:         `{"MachineID":"XXX","Version":19}`,
				},
			},
			want: etcdLease{
				key: "/foo/bar",
				idx: 12,
				ttl: time.Second * 9,
				meta: etcdLeaseMetadata{
					MachineID: "XXX",
					Version:   19,
				},
			},
		},

		// backwards-compatibility with unversioned engines
		{
			res: etcd.Result{
				Node: &etcd.Node{
					Key:           "/foo/bar",
					ModifiedIndex: 12,
					TTL:           9,
					Value:         "XXX",
				},
			},
			want: etcdLease{
				key: "/foo/bar",
				idx: 12,
				ttl: time.Second * 9,
				meta: etcdLeaseMetadata{
					MachineID: "XXX",
					Version:   0,
				},
			},
		},

		// json decode failures are treated like a nonversioned lease
		{
			res: etcd.Result{
				Node: &etcd.Node{
					Key:           "/foo/bar",
					ModifiedIndex: 12,
					TTL:           9,
					Value:         `{"MachineID":"XXX","Ver`,
				},
			},
			want: etcdLease{
				key: "/foo/bar",
				idx: 12,
				ttl: time.Second * 9,
				meta: etcdLeaseMetadata{
					MachineID: `{"MachineID":"XXX","Ver`,
					Version:   0,
				},
			},
		},
	}

	for i, tt := range tests {
		got := leaseFromResult(&tt.res, nil)
		if !reflect.DeepEqual(tt.want, *got) {
			t.Errorf("case %d: incorrect output from leaseFromResult\nwant=%#v\ngot=%#vs", i, tt.want, *got)
		}
	}
}