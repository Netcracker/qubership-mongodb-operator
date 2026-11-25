// Copyright 2024-2025 NetCracker Technology Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	"reflect"
	"testing"
)

func TestGetDATAReplicaSetHostName(t *testing.T) {
	type args struct {
		replicaSize int
		shardIndex  int
		domain      string
		namespace   string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "Standard datars1 names",
			args: args{
				replicaSize: 3,
				shardIndex:  0,
				namespace:   "mongodb",
				domain:      "cluster.local",
			},
			want: []string{"datars10-0.datars1.mongodb.svc.cluster.local:27017",
				"datars11-0.datars1.mongodb.svc.cluster.local:27017",
				"datars12-0.datars1.mongodb.svc.cluster.local:27017"},
		},
		{
			name: "Standard datars2 names",
			args: args{
				replicaSize: 3,
				shardIndex:  1,
				namespace:   "mongodb",
				domain:      "cluster.local",
			},
			want: []string{"datars20-0.datars2.mongodb.svc.cluster.local:27017",
				"datars21-0.datars2.mongodb.svc.cluster.local:27017",
				"datars22-0.datars2.mongodb.svc.cluster.local:27017"},
		},
		{
			name: "Domain cluster-2",
			args: args{
				replicaSize: 3,
				shardIndex:  1,
				namespace:   "mongodb",
				domain:      "cluster-2.local",
			},
			want: []string{"datars20-0.datars2.mongodb.svc.cluster-2.local:27017",
				"datars21-0.datars2.mongodb.svc.cluster-2.local:27017",
				"datars22-0.datars2.mongodb.svc.cluster-2.local:27017"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetDATAReplicaSetHostName(tt.args.replicaSize, tt.args.shardIndex, tt.args.domain, tt.args.namespace); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetDATAReplicaSetHostName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetReplicaSetHostNames(t *testing.T) {
	type args struct {
		replicaSize int
		namePattern string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "Standard datars1 name",
			args: args{
				replicaSize: 3,
				namePattern: "datars1%v-0.datars1.mongodb.svc.cluster.local:27017",
			},
			want: []string{"datars10-0.datars1.mongodb.svc.cluster.local:27017", "datars11-0.datars1.mongodb.svc.cluster.local:27017", "datars12-0.datars1.mongodb.svc.cluster.local:27017"},
		},
		{
			name: "Standard cnfrs name",
			args: args{
				replicaSize: 2,
				namePattern: "cnfrs%v-0.cnfrs.mongodb.svc.cluster.local:27017",
			},
			want: []string{"cnfrs0-0.cnfrs.mongodb.svc.cluster.local:27017", "cnfrs1-0.cnfrs.mongodb.svc.cluster.local:27017"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetReplicaSetHostNames(tt.args.replicaSize, tt.args.namePattern); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetReplicaSetHostNames() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetCNFReplicaSetHostNames(t *testing.T) {
	type args struct {
		replicaSize int
		domain      string
		namespace   string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "Standard cnfrs names",
			args: args{
				replicaSize: 3,
				namespace:   "mongodb",
				domain:      "cluster.local",
			},
			want: []string{"cnfrs0-0.cnfrs.mongodb.svc.cluster.local:27017",
				"cnfrs1-0.cnfrs.mongodb.svc.cluster.local:27017",
				"cnfrs2-0.cnfrs.mongodb.svc.cluster.local:27017"},
		},
		{
			name: "Domain cluster-2",
			args: args{
				replicaSize: 3,
				namespace:   "mongodb",
				domain:      "cluster-2.local",
			},
			want: []string{"cnfrs0-0.cnfrs.mongodb.svc.cluster-2.local:27017",
				"cnfrs1-0.cnfrs.mongodb.svc.cluster-2.local:27017",
				"cnfrs2-0.cnfrs.mongodb.svc.cluster-2.local:27017"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetCNFReplicaSetHostNames(tt.args.replicaSize, tt.args.domain, tt.args.namespace); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetCNFReplicaSetHostNames() = %v, want %v", got, tt.want)
			}
		})
	}
}
